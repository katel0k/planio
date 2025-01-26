package main

import (
	"net/http"
	"sync"
	"time"

	joinPB "github.com/katel0k/planio/server/build/join"
	msgPB "github.com/katel0k/planio/server/build/msg"
	"google.golang.org/protobuf/proto"
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
