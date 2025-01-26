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
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	joinPB "github.com/katel0k/planio/server/build/join"
)

type contextKey int

const (
	DB contextKey = iota
	ONLINE_USERS
	USE_COOKIES
)

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
