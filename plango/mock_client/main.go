package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	msg_pb "github.com/katel0k/planio/mock_client/build/msg"
	"google.golang.org/protobuf/proto"
)

var SERVER_URL string = "http://0.0.0.0:5000"
var PONG_TIMEOUT time.Duration = time.Second * 5

func main() {
	s := http.Client{
		Timeout: PONG_TIMEOUT,
	}
	defer s.CloseIdleConnections()
	resp, err := s.Get(SERVER_URL + "/join/mock")
	if err != nil {
		log.Println(err)
		log.Println("Server error, quitting")
		return
	}
	buffer := make([]byte, 1024)
	n, err := resp.Body.Read(buffer)
	if err != io.EOF && n != 0 {
		log.Println(err)
		log.Println("Server error, quitting")
	}
	id, err := strconv.Atoi(string(buffer[0:n]))
	if err != nil {
		log.Println(err)
		log.Println("Server error, quitting")
		return
	}

	for {
		pong, err := s.Get(fmt.Sprintf("%s/ping/%d", SERVER_URL, id))
		if err != nil {
			log.Println(err)
			log.Println("Server error, quitting")
			return
		}
		var waiter chan []byte = make(chan []byte)
		go (func() {
			defer pong.Body.Close()
			n, _ = pong.Body.Read(buffer)
			if n != 0 {
				fmt.Println(n)
				waiter <- buffer[0:n]
			}
		})()
		buf := <-waiter
		msg := msg_pb.MsgResponse{}
		err = proto.Unmarshal(buf, &msg)
		if err == nil {
			text := msg.Text
			if text != "pong" {
				log.Printf("Got message: %s", text)
			}
		} else {
			log.Println("pong")
		}
	}
}
