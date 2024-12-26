package lib

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Pool *pgxpool.Pool
}

func (db Database) CreateNewUser(username string) (int, error) {
	row := db.Pool.QueryRow(context.Background(), "INSERT INTO users(nickname) VALUES ($1) RETURNING id", username)
	var id int
	err := row.Scan(&id)
	if err != nil {
		log.Default().Print(err)
		log.Default().Printf("Failed to add user in database")
		return 0, err
	} else {
		return id, nil
	}
}

func (db Database) CreateNewMessage(author_id int, receiver_id int, text string) error {
	_, err := db.Pool.Exec(context.Background(),
		"INSERT INTO messages(author_id, receiver_id, body) VALUES ($1, $2, $3)", author_id, receiver_id, text)
	return err
}
