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

func main() {
	s := http.Client{
		Timeout: time.Second * 5, // hanoseconds -> seconds
	}
	defer s.CloseIdleConnections()
	resp, err := s.Get("http://0.0.0.0:5000/join/mock")
	if err != nil {
		log.Println(err)
		log.Println("Server error, quitting")
		return
	}
	buffer := make([]byte, 1024)
	n, err := resp.Body.Read(buffer)
	if err != io.EOF && n != 0 {
		log.Println(err)
		log.Panicln("Server error, quitting")
	}
	id, err := strconv.Atoi(string(buffer[0:n]))
	if err != nil {
		log.Println(err)
		log.Println("Server error, quitting")
		return
	}

	for {
		pong, err := s.Get(fmt.Sprintf("http://0.0.0.0:5000/ping/%d", id))
		if err != nil {
			log.Println(err)
			log.Println("Server error, quitting")
			return
		}
		n, _ = pong.Body.Read(buffer)
		if n != 0 {
			msg := msg_pb.MsgResponse{}
			proto.Unmarshal(buffer[0:n], &msg)
			text := msg.Text
			// text := msg_pb.MsgResponse
			if text != "pong" {
				log.Printf("Got message: \n%s\n", text)
			} else {
				log.Println("pong")
			}
		}
	}
	// join and get id
	// send ping request to server
	// wait for pong with data
	// timeout if it didnt reach client
}
