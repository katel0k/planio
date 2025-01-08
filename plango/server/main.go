package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/katel0k/planio/server/lib"
	"google.golang.org/protobuf/proto"

	join_pb "github.com/katel0k/planio/server/build/join"
	msg_pb "github.com/katel0k/planio/server/build/msg"
	plan_pb "github.com/katel0k/planio/server/build/plan"
)

type contextKey int

const (
	DB contextKey = iota
	ACTIVE_USERS
	USE_COOKIES
)

type ActiveUsers struct {
	sync.RWMutex
	body map[int]chan *msg_pb.MsgResponse
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

func joinHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	headerContentType := r.Header.Get("Content-Type")
	joinReq := join_pb.JoinRequest{}
	if strings.Contains(headerContentType, "application/json") {
		err = json.NewDecoder(r.Body).Decode(&joinReq)
	} else {
		buffer := make([]byte, 1024)
		n, _ := r.Body.Read(buffer)
		err = proto.Unmarshal(buffer[0:n], &joinReq)
	}
	if err != nil {
		return
	}

	id, err := r.Context().Value(DB).(lib.Database).FindUser(joinReq.Username)
	var isNew bool = false

	if err != nil {
		if errors.Is(err, lib.ErrNotFound) {
			log.Default().Printf("creating new user %s", joinReq.Username)
			id, err = r.Context().Value(DB).(lib.Database).CreateNewUser(joinReq.Username)
			log.Default().Print(err)
			if err != nil {
				return
			}
			isNew = true
		} else {
			return
		}
	}

	activeUsers, _ := r.Context().Value(ACTIVE_USERS).(*ActiveUsers)
	activeUsers.Lock()
	activeUsers.body[id] = make(chan *msg_pb.MsgResponse)
	activeUsers.Unlock()
	log.Default().Printf("Got join request for %d", id)
	if r.Context().Value(ACTIVE_USERS).(bool) {
		cookie := http.Cookie{
			Name:   "id",
			Value:  strconv.Itoa(id),
			MaxAge: 300,
		}
		http.SetCookie(w, &cookie)
	}
	marsh, _ := proto.Marshal(&join_pb.JoinResponse{Id: int32(id), IsNew: isNew})
	w.Write(marsh)
}

func messageHandler(w http.ResponseWriter, r *http.Request) {
	var bytes []byte = make([]byte, 1024)
	n, err := r.Body.Read(bytes)
	msg := msg_pb.MsgRequest{}
	err2 := proto.Unmarshal(bytes[0:n], &msg)
	if err2 != nil {
		log.Print(err2)
		return
	}
	receiver := int(msg.ReceiverId)
	if err != nil && err != io.EOF {
		log.Default().Print("error ", err)
		return
	} else {
		id, _ := getId(r)
		msgId, err := r.Context().Value(DB).(lib.Database).CreateNewMessage(id, receiver, msg.Text)
		if err != nil {
			return
		}
		msg := msg_pb.MsgResponse{
			Id:       int32(msgId),
			Text:     msg.Text,
			AuthorId: int32(id),
		}
		activeUsers, _ := r.Context().Value(ACTIVE_USERS).(*ActiveUsers)
		activeUsers.RLock()
		if user, isOnline := activeUsers.body[receiver]; isOnline {
			go (func() {
				user <- &msg
			})()
			log.Default().Printf("Sent message %s", msg.Text)
		}
		activeUsers.RUnlock()
	}
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	log.Default().Printf("GET %s", r.URL)

	id, _ := getId(r)
	select {
	case msg := <-r.Context().Value(ACTIVE_USERS).(*ActiveUsers).body[id]:
		marsh, _ := proto.Marshal(msg)
		w.Write(marsh)
		log.Default().Printf("Pong message %s", msg.String())
	case <-time.After(time.Second * 4):
		w.Write([]byte("pong"))
		log.Default().Print("pong")
	}
}

func listUsersHandler(w http.ResponseWriter, r *http.Request) {
	activeUsers, _ := r.Context().Value(ACTIVE_USERS).(*ActiveUsers)
	activeUsers.RLock()
	for user := range activeUsers.body {
		w.Write([]byte(fmt.Sprint(user) + " "))
	}
	activeUsers.RUnlock()
}

func listPlansHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := getId(r)
	agenda, _ := r.Context().Value(DB).(lib.Database).GetAllPlans(id)
	marsh, _ := proto.Marshal(agenda)
	w.Write(marsh)
}

func addPlanHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	defer r.Body.Close()
	headerContentType := r.Header.Get("Content-Type")
	id, _ := getId(r)
	var planReq plan_pb.PlanRequest

	if strings.Contains(headerContentType, "application/json") {
		json.NewDecoder(r.Body).Decode(&planReq)
	} else {
		buffer := make([]byte, 1024)
		n, _ := r.Body.Read(buffer)
		err := proto.Unmarshal(buffer[0:n], &planReq)
		if err != nil {
			return
		}
	}
	plan, _ := r.Context().Value(DB).(lib.Database).CreateNewPlan(id, &planReq)
	marsh, _ := proto.Marshal(plan)
	w.Write(marsh)
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Id")
		if r.Method == "OPTIONS" {
			return
		}
		next.ServeHTTP(w, r)
	})
}

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
	db := lib.Database{
		Pool: lib.ConnectDB(*databasePort),
	}
	defer db.Pool.Close()

	activeUsers := ActiveUsers{
		body: make(map[int]chan *msg_pb.MsgResponse),
	}

	s := &http.Server{
		Addr: fmt.Sprintf(":%d", *serverPort),
		ConnContext: func(ctx context.Context, _ net.Conn) context.Context {
			ctx = context.WithValue(ctx, DB, db)
			ctx = context.WithValue(ctx, ACTIVE_USERS, &activeUsers)
			ctx = context.WithValue(ctx, USE_COOKIES, *useCookies)
			return ctx
		},
	}
	http.Handle("/join", cors(http.HandlerFunc(joinHandler)))
	http.Handle("/ping", cors(http.HandlerFunc(pingHandler)))
	http.Handle("/message", cors(http.HandlerFunc(messageHandler)))

	http.Handle("/users", cors(http.HandlerFunc(listUsersHandler)))

	http.Handle("/plans", cors(http.HandlerFunc(listPlansHandler)))
	http.Handle("/plan", cors(http.HandlerFunc(addPlanHandler)))
	fileServer := http.FileServer(http.Dir(*staticDir))
	http.Handle("/", cors(http.RedirectHandler("/static/index.html", http.StatusMovedPermanently)))
	http.Handle("/static/", cors(http.StripPrefix("/static", fileServer)))

	log.Fatal(s.ListenAndServe())
}
