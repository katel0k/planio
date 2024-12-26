package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/katel0k/planio/server/lib"
	"google.golang.org/protobuf/proto"

	msg_pb "github.com/katel0k/planio/server/build/msg"
)

var db lib.Database

func connectDB() *pgxpool.Pool {
	url := "postgres://postgres:postgres@localhost:32770/planbook"
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		log.Fatalf("Unable to parse DB config: %v\n", err)
	}

	dbpool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	return dbpool
}

type activeUser struct {
	msgQueue chan *msg_pb.MsgResponse
	id       int
}

func joinHandler(w http.ResponseWriter, r *http.Request) {
	id, err := db.CreateNewUser(r.PathValue("nickname"))
	if err != nil {
		log.Default().Print(err)
		log.Default().Printf("Failed to add user in database")
	} else {
		activeUsersMutex.Lock()
		activeUsers[id] = activeUser{
			msgQueue: make(chan *msg_pb.MsgResponse),
			id:       id,
		}
		activeUsersMutex.Unlock()
		log.Default().Printf("Got join request for %d", id)
		w.Write([]byte(fmt.Sprintf("%d", id)))
	}
}

var activeUsers map[int]activeUser = make(map[int]activeUser)
var activeUsersMutex sync.RWMutex

type conversationHandler struct {
}

func sendMsg(to activeUser, msg *msg_pb.MsgResponse) {
	to.msgQueue <- msg
}

func (h conversationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var receiver, err = strconv.Atoi(r.PathValue("receiver_id"))
	if err != nil {
		return
	}
	var bytes []byte = make([]byte, 1024)
	n, err := r.Body.Read(bytes)
	text := string(bytes[0:n])
	if err != nil && err != io.EOF {
		log.Default().Print("error ", err)
		return
	} else {
		db.CreateNewMessage(1, receiver, text)
		msg := msg_pb.MsgResponse{
			Text:     text,
			AuthorId: int32(receiver),
		}
		activeUsersMutex.RLock()
		if user, isOnline := activeUsers[receiver]; isOnline {
			go sendMsg(user, &msg)
		}
		log.Default().Printf("Sent message %s", text)
	}
}

func listUsers(w http.ResponseWriter, _ *http.Request) {
	activeUsersMutex.RLock()
	for user := range activeUsers {
		w.Write([]byte(fmt.Sprint(user)))
	}
	activeUsersMutex.RUnlock()
}

type Signal int

const (
	Stop Signal = iota
)

type pingHandler struct {
	control chan Signal
}

func (p pingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Default().Printf("GET %s", r.URL)
	id, _ := strconv.Atoi(r.PathValue("id"))
	timer := time.AfterFunc(time.Second*4, func() {
		w.Write([]byte("pong"))
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		p.control <- Stop
	})
	select {
	case msg := <-activeUsers[id].msgQueue:
		marsh, _ := proto.Marshal(msg)
		w.Write(marsh)
		log.Default().Printf("got response %s", msg.String())
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		timer.Stop()
	case sig := <-p.control:
		if sig == Stop {
			return
		}
	}
}

func main() {
	db = lib.Database{
		Pool: connectDB(),
	}
	defer db.Pool.Close()
	s := &http.Server{Addr: ":5000"}
	http.HandleFunc("/users", listUsers)
	http.HandleFunc("/join/{nickname}", joinHandler)
	http.Handle("/message/{receiver_id}", conversationHandler{})
	http.Handle("/ping/{id}", pingHandler{})
	log.Fatal(s.ListenAndServe())
}

// plan:
// i want to write tests because using 3 fucking terminals is god awful for testing
// for me to write tests i need to further decompose the code into testeable modules
// furthermore, i woould need to write functional tests for that purpose
// problem is that i want to write rust client, but i really dont want to deal with that right now
// as that is not the point of this specific project currently
// so i would rather write it in go, but that would go against the spirit of this application.
// soultion is that i'll write mock client for testing in go and a real client later down the line in rust.
