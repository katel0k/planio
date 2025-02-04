package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	PB "github.com/katel0k/planio/protos"
	"google.golang.org/protobuf/proto"
)

func authHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var authReq PB.AuthRequest
	if getRequest(r, &authReq) != nil {
		return
	}

	id, err := r.Context().Value(DB).(Database).FindUser(authReq.Username)
	var success bool

	if err != nil {
		success = false
		if errors.Is(err, ErrNotFound) {
			id, err = r.Context().Value(DB).(Database).CreateNewUser(authReq.Username)
			if err != nil {
				return
			}
		} else {
			return
		}
	} else {
		success = true
	}

	if r.Context().Value(USE_COOKIES).(bool) {
		cookie := http.Cookie{
			Name:   "id",
			Value:  strconv.Itoa(id),
			MaxAge: 300,
		}
		http.SetCookie(w, &cookie)
	}
	response := PB.AuthResponse{Successful: success}
	if success {
		response.Response = &PB.AuthResponse_Id{
			Id: int32(id),
		}
	} else {
		response.Response = &PB.AuthResponse_Reason{
			Reason: "Incorrect username",
		}
	}
	marsh, _ := proto.Marshal(&response)
	w.Write(marsh)
}

// @brief returns user id if it was found, else ErrNotFound
func (db Database) FindUser(username string) (int, error) {
	row := db.Pool.QueryRow(context.Background(), "SELECT id FROM users WHERE nickname=$1", username)
	var id int
	err := row.Scan(&id)
	if err != nil {
		return 0, ErrNotFound
	} else {
		return id, nil
	}
}

func (db Database) CreateNewUser(username string) (int, error) {
	row := db.Pool.QueryRow(context.Background(), "INSERT INTO users(nickname) VALUES ($1) RETURNING id", username)
	var id int
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	} else {
		return id, nil
	}
}
