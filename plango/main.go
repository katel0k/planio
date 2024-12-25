package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	join_pb "github.com/katel0k/planio/build/join"
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

func joinHandler(w http.ResponseWriter, r *http.Request) {
	row := dbpool.QueryRow(context.Background(), "INSERT INTO users(nickname) VALUES ($1) RETURNING id", r.PathValue("nickname"))
	var id int
	err := row.Scan(&id)
	if err != nil {
		log.Default().Print(err)
		log.Default().Printf("Failed to add user in database")
	} else {
		response := join_pb.JoinResponse{Tag: fmt.Sprint(id)}
		w.Write([]byte(response.String()))
		log.Default().Printf("Got join request for %d", id)
	}
}

func main() {
	dbpool = connectDB()
	defer dbpool.Close()
	s := &http.Server{Addr: ":5000"}
	http.HandleFunc("/join/{nickname}", joinHandler)
	log.Fatal(s.ListenAndServe())
}
