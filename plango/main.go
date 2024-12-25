package main

import (
	"log"
	"net/http"

	join_pb "github.com/katel0k/planio/build/join"
)

func joinHandler(w http.ResponseWriter, _ *http.Request) {
	response := join_pb.JoinResponse{Tag: "Hello"}
	w.Write([]byte(response.String()))
	log.Default().Print("Got join request")
}

func main() {
	s := &http.Server{Addr: ":5000"}
	http.HandleFunc("/join", joinHandler)
	log.Fatal(s.ListenAndServe())
}
