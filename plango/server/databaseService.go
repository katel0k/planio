package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	msgPB "github.com/katel0k/planio/server/build/msg"
	planPB "github.com/katel0k/planio/server/build/plan"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Database struct {
	Pool *pgxpool.Pool
}

func ConnectDB(port int) *pgxpool.Pool {
	url := fmt.Sprintf("postgres://postgres:postgres@localhost:%d/planbook", port)
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		log.Fatalf("Unable to parse DB config: %v\n", err)
	}

	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		dataType, err := conn.LoadType(context.Background(), "time_scale")
		if err != nil {
			return err
		}
		conn.TypeMap().RegisterType(dataType)

		return nil
	}

	dbpool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	return dbpool
}

var ErrNotFound = errors.New("not found")

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

func (db Database) CreateNewMessage(authorId int, receiverId int, text string) (int, error) {
	row := db.Pool.QueryRow(context.Background(),
		"INSERT INTO messages(author_id, receiver_id, body) VALUES ($1, $2, $3) RETURNING id", authorId, receiverId, text)
	var id int
	err := row.Scan(&id)
	if err != nil {
		log.Default().Print(err)
		log.Default().Printf("Failed to add message in database")
		return 0, err
	}
	return id, nil
}

func (db Database) GetAllMessages(req *msgPB.AllMessagesRequest) (*msgPB.AllMessagesResponse, error) {
	rows, err := db.Pool.Query(context.Background(),
		"SELECT id, author_id, text FROM messages WHERE author_id=$1 AND receiver_id=$2", req.SenderId, req.ReceiverId)
	var resp msgPB.AllMessagesResponse
	for rows.Next() {
		var msg msgPB.MsgResponse
		rows.Scan(&msg.Id, &msg.AuthorId, &msg.Text)
		resp.Messages = append(resp.Messages, &msg)
	}
	return &resp, err
}

func (db Database) GetAllPlans(userId int) (*planPB.Agenda, error) {
	rows, err := db.Pool.Query(context.Background(),
		`SELECT id, synopsis, creation_dttm, parent_id, scale, body as description
		FROM plans FULL OUTER JOIN descriptions ON id=plan_id
		WHERE author_id=$1`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var agenda planPB.Agenda
	for rows.Next() {
		var plan planPB.Plan
		var scale string
		var creationTime time.Time
		rows.Scan(&plan.Id, &plan.Synopsis, &creationTime, &plan.Parent, &scale, &plan.Description)
		plan.CreationTime = timestamppb.New(creationTime)
		plan.Scale = planPB.TimeScale(planPB.TimeScale_value[scale])
		agenda.Plans = append(agenda.Plans, &plan)
	}
	return &agenda, nil
}

func (db Database) CreateNewPlan(authorId int, plan *planPB.NewPlanRequest) (*planPB.Plan, error) {
	row := db.Pool.QueryRow(context.Background(),
		`INSERT INTO plans(author_id, synopsis, parent_id, scale) VALUES ($1, $2, $3, $4)
		RETURNING id, synopsis, creation_dttm, parent_id, scale`,
		authorId, plan.Synopsis, plan.Parent, plan.Scale)
	var res planPB.Plan
	var creationTime time.Time
	var scale string
	err := row.Scan(&res.Id, &res.Synopsis, &creationTime, &res.Parent, &scale)
	res.CreationTime = timestamppb.New(creationTime)
	res.Scale = planPB.TimeScale(planPB.TimeScale_value[scale])
	if err != nil {
		log.Default().Print(err)
		log.Default().Printf("Failed to add message in database")
		return nil, err
	}
	return &res, nil
}

func (db Database) ChangePlan(plan *planPB.ChangePlanRequest) error {
	_, err := db.Pool.Exec(context.Background(), "UPDATE plans SET synopsis=$1 WHERE id=$2", plan.Synopsis, plan.Id)
	return err
}

func (db Database) DeletePlan(plan_id int) error {
	_, err := db.Pool.Exec(context.Background(), "DELETE FROM plans WHERE id=$1", plan_id)
	return err
}
