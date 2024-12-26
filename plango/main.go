package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	msg_pb "github.com/katel0k/planio/build/msg"
)

var dbpool *pgxpool.Pool

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
	w        *http.ResponseWriter
}

func joinHandler(w http.ResponseWriter, r *http.Request) {
	row := dbpool.QueryRow(context.Background(), "INSERT INTO users(nickname) VALUES ($1) RETURNING id", r.PathValue("nickname"))
	var id int
	err := row.Scan(&id)
	if err != nil {
		log.Default().Print(err)
		log.Default().Printf("Failed to add user in database")
	} else {
		// response := join_pb.JoinResponse{Id: int32(id)}
		activeUsersMutex.Lock()
		activeUsers[id] = activeUser{
			msgQueue: make(chan *msg_pb.MsgResponse),
			w:        &w,
		}
		activeUsersMutex.Unlock()
		log.Default().Printf("Got join request for %d", id)
		msg := <-activeUsers[id].msgQueue
		(*activeUsers[id].w).Write([]byte(msg.String()))
		log.Default().Printf("got response %s", msg.String())
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
	_, err = r.Body.Read(bytes)
	text := string(bytes)
	if err != nil && err != io.EOF {
		log.Default().Print("error ", err)
		return
	} else {
		dbpool.Exec(context.Background(),
			"INSERT INTO messages(author_id, receiver_id, body) VALUES ($1, $2, $3)", "1", receiver, text)
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

func main() {
	dbpool = connectDB()
	defer dbpool.Close()
	s := &http.Server{Addr: ":5000"}
	http.HandleFunc("/join/{nickname}", joinHandler)
	http.Handle("/message/{receiver_id}", conversationHandler{})
	log.Fatal(s.ListenAndServe())
}
