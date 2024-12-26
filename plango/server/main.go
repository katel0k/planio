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

var userMessageChannels map[int]chan *msg_pb.MsgResponse = make(map[int]chan *msg_pb.MsgResponse)
var userChannelsMutex sync.RWMutex

func joinHandler(w http.ResponseWriter, r *http.Request) {
	id, err := db.CreateNewUser(r.PathValue("nickname"))
	if err != nil {
		log.Default().Print(err)
		log.Default().Printf("Failed to add user in database")
	} else {
		userChannelsMutex.Lock()
		userMessageChannels[id] = make(chan *msg_pb.MsgResponse)
		userChannelsMutex.Unlock()
		log.Default().Printf("Got join request for %d", id)
		cookie := http.Cookie{
			Name:   "id",
			Value:  strconv.Itoa(id),
			MaxAge: 300,
		}
		http.SetCookie(w, &cookie)
		w.Write([]byte(fmt.Sprintf("%d", id)))
	}
}

func messageHandler(w http.ResponseWriter, r *http.Request) {
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
		userChannelsMutex.RLock()
		if user, isOnline := userMessageChannels[receiver]; isOnline {
			go (func() {
				user <- &msg
			})()
			log.Default().Printf("Sent message %s", text)
		}
		userChannelsMutex.RUnlock()
	}
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	log.Default().Printf("GET %s", r.URL)

	idStr, err := r.Cookie("id")
	if err == http.ErrNoCookie {
		return
	}
	id, _ := strconv.Atoi(idStr.Value)
	select {
	case msg := <-userMessageChannels[id]:
		marsh, _ := proto.Marshal(msg)
		w.Write(marsh)
		log.Default().Printf("Pong message %s", msg.String())
	case <-time.After(time.Second * 4):
		w.Write([]byte("pong"))
		log.Default().Print("pong")
	}
}

func listUsers(w http.ResponseWriter, _ *http.Request) {
	userChannelsMutex.RLock()
	for user := range userMessageChannels {
		w.Write([]byte(fmt.Sprint(user)))
	}
	userChannelsMutex.RUnlock()
}

func main() {
	db = lib.Database{
		Pool: connectDB(),
	}
	defer db.Pool.Close()
	s := &http.Server{Addr: ":5000"}
	http.HandleFunc("/join/{nickname}", joinHandler)
	http.HandleFunc("/ping", pingHandler)
	http.HandleFunc("/message/{receiver_id}", messageHandler)

	http.HandleFunc("/users", listUsers)
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
