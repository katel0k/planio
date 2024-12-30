package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	msg_pb "github.com/katel0k/planio/mock_client/build/msg"
	plan_pb "github.com/katel0k/planio/mock_client/build/plan"
	"google.golang.org/protobuf/proto"
)

var PONG_TIMEOUT time.Duration = time.Second * 5

func interactWithUser(ctx context.Context, c *http.Client) {
	SERVER_URL, _ := url.Parse("http://0.0.0.0:5000")
	msgURL := SERVER_URL.JoinPath("/message")
	allUsersURL := SERVER_URL.JoinPath("/users")
	allPlansURL := SERVER_URL.JoinPath("/plans")
	newPlanURL := SERVER_URL.JoinPath("/plan")
	buffer := make([]byte, 1024)
	for {
		select {
		case <-ctx.Done():
			return
		default:

			fmt.Println("Choose next action: /m - message, /p - plans, /n - new plan")
			var cmd rune
			fmt.Scanf("/%c\n", &cmd)
			switch cmd {
			case 'm':
				resp, _ := c.Get(allUsersURL.String())
				defer resp.Body.Close()
				n, _ := resp.Body.Read(buffer)
				fmt.Println("Active users you can write to:")
				allUsers := strings.Split(string(buffer[0:n]), " ")
				fmt.Println(string(buffer[0:n]))
				fmt.Println("Which user do you want to write to:")
				var id string
				for {
					var possibleId int
					fmt.Scanf("%d\n", &possibleId)
					id = strconv.Itoa(possibleId)
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
			case 'p':
				resp, _ := c.Get(allPlansURL.String())
				defer resp.Body.Close()
				n, _ := resp.Body.Read(buffer)
				var agenda plan_pb.Agenda
				proto.Unmarshal(buffer[0:n], &agenda)
				for plan := range agenda.Plans {
					fmt.Println(agenda.Plans[plan].String())
				}
				fmt.Println()
			case 'n':
				fmt.Println("Write synopsis of your plan:")
				synopsis, _ := bufio.NewReader(os.Stdin).ReadString('\n')
				plan := plan_pb.Plan{
					Synopsis: synopsis,
				}
				marsh, _ := proto.Marshal(&plan)
				resp, _ := c.Post(newPlanURL.String(), "text/plain", bytes.NewReader(marsh))
				n, _ := resp.Body.Read(buffer)
				fmt.Printf("Created plan: %s\n", string(buffer[0:n]))
			default:
				fmt.Printf("Unrecognised command: %c\n", cmd)
			}
		}
	}
}

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
	defer resp.Body.Close()
	c.Jar.SetCookies(joinURL, resp.Cookies())
	buffer := make([]byte, 1024)
	n, err := resp.Body.Read(buffer)
	if err != io.EOF && n != 0 {
		log.Println(err)
		log.Println("Server error, quitting")
	}

	pingPongContext, cancel := context.WithCancel(context.Background())
	defer cancel()

	go interactWithUser(pingPongContext, &c)
	for {
		pong, err := c.Get(pingURL.String())

		if err != nil {
			log.Println(err)
			log.Println("Server error, quitting")
			return
		}
		defer pong.Body.Close()
		n, _ := pong.Body.Read(buffer)
		buf := buffer[0:n]
		msg := msg_pb.MsgResponse{}
		err = proto.Unmarshal(buf, &msg)
		if err == nil {
			log.Printf("%d received msg %d: %s", msg.AuthorId, msg.Id, msg.Text)
		} else if string(buf) == "pong" {
			if *printPingInfo {
				log.Printf("pong")
			}
		} else {
			log.Print(err)
		}
	}
}
