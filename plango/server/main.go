package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	joinPB "github.com/katel0k/planio/server/build/join"
	msgPB "github.com/katel0k/planio/server/build/msg"
	planPB "github.com/katel0k/planio/server/build/plan"
)

type contextKey int

const (
	DB contextKey = iota
	ONLINE_USERS
	USE_COOKIES
)

type userOnline struct {
	msgChan chan *msgPB.MsgResponse
}

type onlineUsers struct {
	sync.RWMutex
	body map[int]userOnline
}

func (user *userOnline) sendMessage(msg *msgPB.MsgResponse) {
	user.msgChan <- msg
}

func (users *onlineUsers) addUser(id int) {
	users.Lock()
	users.body[id] = userOnline{msgChan: make(chan *msgPB.MsgResponse)}
	users.Unlock()
}

// FIXME: that is a temporary solution because I was too lazy to setup a proper server
// If you are serving html from file://, cookies just dont work
// So instead, I'm just gonna send an "Id" header from frontend (REALLY SAFE METHOD TRUST ME)
// It is also good for testing, so in the future I'm probably going to hide it behind an interface
func getId(r *http.Request) (int, error) {
	useCookies, ok := r.Context().Value(USE_COOKIES).(bool)
	if !ok {
		useCookies = DEFAULT_USE_COOKIES
	}
	if useCookies {
		idStr, err := r.Cookie("id")
		if err == http.ErrNoCookie {
			return 0, err
		}
		return strconv.Atoi(idStr.Value)
	} else {
		idStr := r.Header.Get("Id")
		if idStr == "" {
			return 0, nil
		}
		return strconv.Atoi(idStr)
	}
}

// gets proto request base on content type. If application/json - gets it from JSON, otherwise - from bytes
func getRequest(r *http.Request, m proto.Message) error {
	var err error
	headerContentType := r.Header.Get("Content-Type")
	buffer := make([]byte, 1024)
	n, _ := r.Body.Read(buffer)
	if strings.Contains(headerContentType, "application/json") {
		err = protojson.Unmarshal(buffer[0:n], m)
	} else {
		err = proto.Unmarshal(buffer[0:n], m)
	}
	return err
}

func joinHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var joinReq joinPB.JoinRequest
	if getRequest(r, &joinReq) != nil {
		return
	}

	id, err := r.Context().Value(DB).(Database).FindUser(joinReq.Username)
	var isNew bool = false

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			id, err = r.Context().Value(DB).(Database).CreateNewUser(joinReq.Username)
			if err != nil {
				return
			}
			isNew = true
		} else {
			return
		}
	}

	onlineUsers, _ := r.Context().Value(ONLINE_USERS).(*onlineUsers)
	onlineUsers.addUser(id)
	if r.Context().Value(USE_COOKIES).(bool) {
		cookie := http.Cookie{
			Name:   "id",
			Value:  strconv.Itoa(id),
			MaxAge: 300,
		}
		http.SetCookie(w, &cookie)
	}
	marsh, _ := proto.Marshal(&joinPB.JoinResponse{Id: int32(id), IsNew: isNew})
	w.Write(marsh)
}

func messageHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var msg msgPB.MsgRequest
	if getRequest(r, &msg) != nil {
		return
	}
	receiver := int(msg.ReceiverId)

	id, _ := getId(r)
	msgId, err := r.Context().Value(DB).(Database).CreateNewMessage(id, receiver, msg.Text)
	if err != nil {
		return
	}
	response := msgPB.MsgResponse{
		Id:       int32(msgId),
		Text:     msg.Text,
		AuthorId: int32(id),
	}
	onlineUsers, _ := r.Context().Value(ONLINE_USERS).(*onlineUsers)
	onlineUsers.RLock()
	if user, isOnline := onlineUsers.body[receiver]; isOnline {
		go user.sendMessage(&response)
	}
	onlineUsers.RUnlock()
}

func messagesHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var msg msgPB.AllMessagesRequest
	if getRequest(r, &msg) != nil {
		return
	}
	msgs, _ := r.Context().Value(DB).(Database).GetAllMessages(&msg)
	marsh, _ := proto.Marshal(msgs)
	w.Write(marsh)
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := getId(r)
	select {
	case msg := <-r.Context().Value(ONLINE_USERS).(*onlineUsers).body[id].msgChan:
		marsh, _ := proto.Marshal(msg)
		w.Write(marsh)
	case <-time.After(PING_RESPONSE_TIME):
		w.Write([]byte("pong"))
	}
}

func onlineUsersHandler(w http.ResponseWriter, r *http.Request) {
	onlineUsers, _ := r.Context().Value(ONLINE_USERS).(*onlineUsers)
	onlineUsers.RLock()
	var resp joinPB.JoinedUsersResponse
	for userId := range onlineUsers.body {
		resp.Users = append(resp.Users, &joinPB.User{
			Id: int32(userId),
		})
	}
	onlineUsers.RUnlock()
	marsh, _ := proto.Marshal(&resp)
	w.Write(marsh)
}

func planHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		id, _ := getId(r)
		agenda, _ := r.Context().Value(DB).(Database).GetAllPlans(id)
		marsh, _ := proto.Marshal(agenda)
		w.Write(marsh)
	case "POST":
		defer r.Body.Close()
		id, _ := getId(r)
		var planReq planPB.NewPlanRequest
		if err := getRequest(r, &planReq); err != nil {
			log.Default().Print(err)
			return
		}
		plan, _ := r.Context().Value(DB).(Database).CreateNewPlan(id, &planReq)
		marsh, _ := proto.Marshal(plan)
		w.Write(marsh)
	case "PATCH":
		defer r.Body.Close()
		var planReq planPB.ChangePlanRequest
		if getRequest(r, &planReq) != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err := r.Context().Value(DB).(Database).ChangePlan(&planReq)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	case "DELETE":
		defer r.Body.Close()
		var planReq planPB.DeletePlanRequest
		if getRequest(r, &planReq) != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err := r.Context().Value(DB).(Database).DeletePlan(int(planReq.Id))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Id")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		if r.Method == "OPTIONS" {
			return
		}
		next.ServeHTTP(w, r)
	})
}

const PING_RESPONSE_TIME time.Duration = time.Second * 4
const DEFAULT_STATIC_DIR string = "../../planer/dist"
const DEFAULT_DATABASE_PORT int = 32768
const DEFAULT_USE_COOKIES bool = false
const DEFAULT_SERVER_PORT int = 5000

func main() {
	staticDir := flag.String("static", DEFAULT_STATIC_DIR, "Directory with static files")
	databasePort := flag.Int("dbp", DEFAULT_DATABASE_PORT, "Database port")
	useCookies := flag.Bool("c", DEFAULT_USE_COOKIES, "Use cookies or simple join id and a header")
	serverPort := flag.Int("p", DEFAULT_SERVER_PORT, "Server port")
	flag.Parse()
	db := Database{
		Pool: ConnectDB(*databasePort),
	}
	defer db.Pool.Close()

	onlineUsers := onlineUsers{
		body: make(map[int]userOnline),
	}
	logger := Logging(log.New(os.Stdout, "", log.LstdFlags))

	s := &http.Server{
		Addr: fmt.Sprintf(":%d", *serverPort),
		ConnContext: func(ctx context.Context, _ net.Conn) context.Context {
			ctx = context.WithValue(ctx, DB, db)
			ctx = context.WithValue(ctx, ONLINE_USERS, &onlineUsers)
			ctx = context.WithValue(ctx, USE_COOKIES, *useCookies)
			return ctx
		},
		Handler: logger(cors(http.DefaultServeMux)),
	}
	http.HandleFunc("/join", joinHandler)
	http.HandleFunc("/ping", pingHandler)
	http.HandleFunc("/message", messageHandler)
	http.HandleFunc("/messages", messagesHandler)

	http.HandleFunc("/online", onlineUsersHandler)

	http.HandleFunc("/plan", planHandler)

	http.Handle("/", http.FileServer(http.Dir(*staticDir)))

	log.Fatal(s.ListenAndServe())
}
