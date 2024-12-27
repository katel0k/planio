package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	msg_pb "github.com/katel0k/planio/mock_client/build/msg"
	"google.golang.org/protobuf/proto"
)

var PONG_TIMEOUT time.Duration = time.Second * 5

func pingPong(c *http.Client, pingURL *url.URL, printPingInfo bool) {
	buffer := make([]byte, 1024)
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
			n, _ := pong.Body.Read(buffer)
			if n != 0 {
				waiter <- buffer[0:n]
			}
		})()
		buf := <-waiter
		msg := msg_pb.MsgResponse{}
		err = proto.Unmarshal(buf, &msg)
		if err == nil {
			log.Printf("%d received msg %d: %s", msg.AuthorId, msg.Id, msg.Text)
		} else if string(buf) == "pong" {
			if printPingInfo {
				log.Printf("pong")
			}
		} else {
			log.Print(err)
		}
	}
}

func main() {
	printPingInfo := flag.Bool("d", false, "Log ping pong debug info")
	flag.Parse()

	SERVER_URL, _ := url.Parse("http://0.0.0.0:5000")
	joinURL := SERVER_URL.JoinPath("/join/mock")
	pingURL := SERVER_URL.JoinPath("/ping")
	msgURL := SERVER_URL.JoinPath("/message")
	allUsersURL := SERVER_URL.JoinPath("/users")

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

	go pingPong(&c, pingURL, *printPingInfo)
	for {
		resp, _ := c.Get(allUsersURL.String())
		n, _ := resp.Body.Read(buffer)
		fmt.Println("Active users you can write to:")
		allUsers := strings.Split(string(buffer[0:n]), " ")
		fmt.Println(string(buffer[0:n]))
		fmt.Println("Which user do you want to write to:")
		var id string
		for {
			fmt.Scanf("%s\n", &id)
			if slices.Contains(allUsers, id) {
				break
			}
			fmt.Printf("User not found, try again:\n")
		}
		fmt.Println("Write message:")
		var msg string
		fmt.Scanln(&msg)
		iid, _ := strconv.Atoi(id)
		m := msg_pb.MsgRequest{
			Text:       msg,
			ReceiverId: int32(iid),
		}
		res, _ := proto.Marshal(&m)

		c.Post(msgURL.String(), "text/plain", bytes.NewReader(res))
	}
}
