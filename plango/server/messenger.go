package main

import (
	"net/http"
	"sync"
	"time"

	// PB "github.com/katel0k/planio/server/build/join"
	PB "github.com/katel0k/planio/server/protos"
	"google.golang.org/protobuf/proto"
)

type userOnline struct {
	msgChan chan *PB.MsgResponse
}

type onlineUsers struct {
	sync.RWMutex
	body map[int]userOnline
}

func (user *userOnline) sendMessage(msg *PB.MsgResponse) {
	user.msgChan <- msg
}

func (users *onlineUsers) addUser(id int) {
	users.Lock()
	users.body[id] = userOnline{msgChan: make(chan *PB.MsgResponse)}
	users.Unlock()
}

func joinHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var joinReq PB.JoinRequest // TODO: take that from cookie instead
	if getRequest(r, &joinReq) != nil {
		return
	}
	id, _ := r.Context().Value(DB).(Database).FindUser(joinReq.Username)

	onlineUsers, _ := r.Context().Value(ONLINE_USERS).(*onlineUsers)
	onlineUsers.addUser(id)
	w.WriteHeader(http.StatusOK)
}

func messageHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var msg PB.MsgRequest
	if getRequest(r, &msg) != nil {
		return
	}
	receiver := int(msg.ReceiverId)

	id, _ := getId(r)
	msgId, err := r.Context().Value(DB).(Database).CreateNewMessage(id, receiver, msg.Text)
	if err != nil {
		return
	}
	response := PB.MsgResponse{
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
	var msg PB.AllMessagesRequest
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
	var resp PB.JoinedUsersResponse
	for userId := range onlineUsers.body {
		resp.Users = append(resp.Users, &PB.User{
			Id: int32(userId),
		})
	}
	onlineUsers.RUnlock()
	marsh, _ := proto.Marshal(&resp)
	w.Write(marsh)
}
