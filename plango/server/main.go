package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/katel0k/planio/server/lib"
	"google.golang.org/protobuf/proto"

	msg_pb "github.com/katel0k/planio/server/build/msg"
	plan_pb "github.com/katel0k/planio/server/build/plan"
)

func getIdFromCookie(r *http.Request) (int, error) {
	idStr, err := r.Cookie("id")
	if err == http.ErrNoCookie {
		return 0, err
	}
	return strconv.Atoi(idStr.Value)
}

var db lib.Database

var userMessageChannels map[int]chan *msg_pb.MsgResponse = make(map[int]chan *msg_pb.MsgResponse)
var userChannelsMutex sync.RWMutex

func joinHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
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
		id, _ := getIdFromCookie(r)
		msgId, err := db.CreateNewMessage(id, receiver, msg.Text)
		if err != nil {
			return
		}
		msg := msg_pb.MsgResponse{
			Id:       int32(msgId),
			Text:     msg.Text,
			AuthorId: int32(id),
		}
		userChannelsMutex.RLock()
		if user, isOnline := userMessageChannels[receiver]; isOnline {
			go (func() {
				user <- &msg
			})()
			log.Default().Printf("Sent message %s", msg.Text)
		}
		userChannelsMutex.RUnlock()
	}
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	log.Default().Printf("GET %s", r.URL)

	id, _ := getIdFromCookie(r)
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
		w.Write([]byte(fmt.Sprint(user) + " "))
	}
	userChannelsMutex.RUnlock()
}

func listPlans(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	id, _ := getIdFromCookie(r)
	agenda, _ := db.GetAllPlans(id)
	marsh, _ := proto.Marshal(agenda)
	w.Write(marsh)
}

func addPlan(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	id, _ := getIdFromCookie(r)
	buffer := make([]byte, 1024)
	n, _ := r.Body.Read(buffer)
	var planReq plan_pb.PlanRequest
	err := proto.Unmarshal(buffer[0:n], &planReq)
	if err != nil {
		return
	}
	plan, _ := db.CreateNewPlan(id, &planReq)
	marsh, _ := proto.Marshal(plan)
	w.Write(marsh)
}

func main() {
	db = lib.Database{
		Pool: lib.ConnectDB(),
	}
	defer db.Pool.Close()
	s := &http.Server{Addr: ":5000"}
	http.HandleFunc("/join/{nickname}", joinHandler)
	http.HandleFunc("/ping", pingHandler)
	http.HandleFunc("/message", messageHandler)

	http.HandleFunc("/users", listUsers)

	http.HandleFunc("/plans", listPlans)
	http.HandleFunc("/plan", addPlan)
	log.Fatal(s.ListenAndServe())
}
