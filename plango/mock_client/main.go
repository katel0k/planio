package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	msg_pb "github.com/katel0k/planio/mock_client/build/msg"
	"google.golang.org/protobuf/proto"
)

var PONG_TIMEOUT time.Duration = time.Second * 5

func main() {
	printPingInfo := flag.Bool("d", false, "Log ping pong debug info")
	flag.Parse()

	SERVER_URL, _ := url.Parse("http://0.0.0.0:5000")
	joinURL := SERVER_URL.JoinPath("/join/mock")
	pingURL := SERVER_URL.JoinPath("/ping")

	jar, _ := cookiejar.New(nil)
	c := http.Client{
		Timeout: PONG_TIMEOUT,
		Jar:     jar,
	}
	defer c.CloseIdleConnections()
	resp, err := c.Get(joinURL.String())
	if err != nil {
		log.Println(err)
		log.Println("Server error, quitting")
		return
	}
	c.Jar.SetCookies(joinURL, resp.Cookies())
	buffer := make([]byte, 1024)
	n, err := resp.Body.Read(buffer)
	if err != io.EOF && n != 0 {
		log.Println(err)
		log.Println("Server error, quitting")
	}

	for {
		pong, err := c.Get(pingURL.String())

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
				waiter <- buffer[0:n]
			}
		})()
		buf := <-waiter
		msg := msg_pb.MsgResponse{}
		err = proto.Unmarshal(buf, &msg)
		if err == nil {
			log.Printf("%d sent: %s", msg.AuthorId, msg.Text)
		} else if string(buf) == "pong" {
			if *printPingInfo {
				log.Printf("pong")
			}
		} else {
			log.Print(err)
		}
	}
}
