package lib

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	plan_pb "github.com/katel0k/planio/server/build/plan"
)

type Database struct {
	Pool *pgxpool.Pool
}

func ConnectDB() *pgxpool.Pool {
	url := "postgres://postgres:postgres@localhost:32771/planbook"
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

func (db Database) CreateNewMessage(author_id int, receiver_id int, text string) (int, error) {
	row := db.Pool.QueryRow(context.Background(),
		"INSERT INTO messages(author_id, receiver_id, body) VALUES ($1, $2, $3) RETURNING id", author_id, receiver_id, text)
	var id int
	err := row.Scan(&id)
	if err != nil {
		log.Default().Print(err)
		log.Default().Printf("Failed to add message in database")
		return 0, err
	}
	return id, nil
}

func (db Database) GetAllPlans(user_id int) (*plan_pb.Agenda, error) {
	rows, err := db.Pool.Query(context.Background(),
		"SELECT id, synopsis FROM plans WHERE author_id=$1", user_id)
	if err != nil {
		return nil, err
	}
	var agenda plan_pb.Agenda
	defer rows.Close()
	for rows.Next() {
		var plan plan_pb.Plan
		rows.Scan(&plan.Id, &plan.Synopsis)
		agenda.Plans = append(agenda.Plans, &plan)
	}
	return &agenda, nil
}

func (db Database) CreateNewPlan(author_id int, plan *plan_pb.Plan) (int, error) {
	row := db.Pool.QueryRow(context.Background(),
		"INSERT INTO plans(author_id, synopsis) VALUES ($1, $2) RETURNING id", author_id, plan.Synopsis)
	var id int
	err := row.Scan(&id)
	if err != nil {
		log.Default().Print(err)
		log.Default().Printf("Failed to add message in database")
		return 0, err
	}
	return id, nil
}
