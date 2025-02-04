package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	PB "github.com/katel0k/planio/protos"
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

func (db Database) GetAllMessages(req *PB.AllMessagesRequest) (*PB.AllMessagesResponse, error) {
	rows, err := db.Pool.Query(context.Background(),
		"SELECT id, author_id, text FROM messages WHERE author_id=$1 AND receiver_id=$2", req.SenderId, req.ReceiverId)
	var resp PB.AllMessagesResponse
	for rows.Next() {
		var msg PB.MsgResponse
		rows.Scan(&msg.Id, &msg.AuthorId, &msg.Text)
		resp.Messages = append(resp.Messages, &msg)
	}
	return &resp, err
}

func (db Database) GetAllPlans(userId int) (*PB.UserPlans, error) {
	rows, err := db.Pool.Query(context.Background(),
		`SELECT id, synopsis, creation_dttm, parent_id, scale, body as description, start_dttm, end_dttm
		FROM plans FULL OUTER JOIN descriptions d ON id=d.plan_id
					FULL OUTER JOIN timeframes t ON id=t.plan_id
		WHERE author_id=$1`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res PB.UserPlans
	for rows.Next() {
		var plan PB.Plan
		var scale string
		var creationTime time.Time
		var startTime *time.Time
		var endTime *time.Time
		rows.Scan(&plan.Id, &plan.Synopsis, &creationTime, &plan.Parent, &scale, &plan.Description, &startTime, &endTime)
		plan.CreationTime = timestamppb.New(creationTime)
		if startTime != nil && endTime != nil {
			plan.Timeframe = &PB.Timeframe{
				Start: timestamppb.New(*startTime),
				End:   timestamppb.New(*endTime),
			}
		}
		plan.Scale = PB.TimeScale(PB.TimeScale_value[scale])
		res.Body = append(res.Body, &plan)
	}
	res.UserId = int32(userId)
	evs, err := db.GetEvents(userId)
	if err != nil {
		return nil, err
	}
	res.Events = evs
	return &res, nil
}

type agendaNodePrototype struct {
	body     *PB.Agenda_AgendaNode
	parent   *int32
	subplans []*agendaNodePrototype
}

func (db Database) GetAgenda(userId int) (*PB.Agenda, error) {
	rows, err := db.Pool.Query(context.Background(),
		`SELECT id, parent_id, scale FROM plans WHERE author_id=$1`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	prototype := make([]*agendaNodePrototype, 0)
	for rows.Next() {
		node := agendaNodePrototype{
			body:   &PB.Agenda_AgendaNode{},
			parent: nil,
		}
		var scale string
		rows.Scan(&node.body.Id, &node.parent, &scale)
		node.body.Scale = PB.TimeScale(PB.TimeScale_value[scale])
		prototype = append(prototype, &node)
	}
	prototype = getScaleTreePrototype(prototype)
	res := PB.Agenda{
		Body: nil, // root
	}
	for i := range prototype {
		res.Subplans = append(res.Subplans, convertPrototypeToAgenda(prototype[i]))
	}
	return &res, nil
}

func getScaleTreePrototype(plans []*agendaNodePrototype) []*agendaNodePrototype {
	// converts plain tree into an actual tree. Works in O(number of edges)
	// at least if we believe in good go compiler
	q := make([]int, 0)
	m := make(map[int32]*agendaNodePrototype)
	for p := range plans {
		q = append(q, p)
		m[plans[p].body.Id] = plans[p]
	}
	for len(q) > 0 {
		ind := q[0]
		p := plans[ind]
		q = q[1:]
		if p.parent != nil {
			if pl, ok := m[*p.parent]; ok {
				pl.subplans = append(pl.subplans, p)
			} else {
				q = append(q, ind)
			}
		} else {
			m[p.body.Id] = p
		}
	}
	res := make([]*agendaNodePrototype, 0)
	for p := range m {
		if m[p].parent == nil {
			res = append(res, m[p])
		}
	}
	return res
}

func convertPrototypeToAgenda(prototype *agendaNodePrototype) *PB.Agenda {
	subplans := make([]*PB.Agenda, 0)
	for i := range prototype.subplans {
		subplans = append(subplans, convertPrototypeToAgenda(prototype.subplans[i]))
	}
	return &PB.Agenda{
		Body:     prototype.body,
		Subplans: subplans,
	}
}

func (db Database) CreateNewPlan(authorId int, plan *PB.NewPlanRequest) (*PB.Plan, error) {
	row := db.Pool.QueryRow(context.Background(),
		`INSERT INTO plans(author_id, synopsis, parent_id, scale) VALUES ($1, $2, $3, $4)
		RETURNING id, synopsis, creation_dttm, parent_id, scale`,
		authorId, plan.Synopsis, plan.Parent, plan.Scale)
	var res PB.Plan
	var creationTime time.Time
	var scale string
	err := row.Scan(&res.Id, &res.Synopsis, &creationTime, &res.Parent, &scale)
	res.CreationTime = timestamppb.New(creationTime)
	res.Scale = PB.TimeScale(PB.TimeScale_value[scale])
	if err != nil {
		log.Default().Print(err)
		log.Default().Printf("Failed to add message in database")
		return nil, err
	}
	return &res, nil
}

func (db Database) ChangePlan(plan *PB.ChangePlanRequest) error {
	_, err := db.Pool.Exec(context.Background(), "UPDATE plans SET synopsis=$1 WHERE id=$2", plan.Synopsis, plan.Id)
	return err
}

func (db Database) DeletePlan(plan_id int) error {
	_, err := db.Pool.Exec(context.Background(), "DELETE FROM plans WHERE id=$1", plan_id)
	return err
}

func (db Database) GetEvents(authorId int) ([]*PB.Event, error) {
	rows, err := db.Pool.Query(context.Background(),
		`SELECT id, synopsis, creation_dttm, dttm FROM events WHERE author_id=$1`, authorId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := make([]*PB.Event, 0)
	for rows.Next() {
		var event PB.Event
		var creationTime time.Time
		var dttm time.Time
		rows.Scan(&event.Id, &event.Synopsis, &creationTime, &dttm)
		event.CreationTime = timestamppb.New(creationTime)
		event.Time = timestamppb.New(dttm)
		events = append(events, &event)
	}
	return events, nil
}

func (db Database) CreateEvent(authorId int, newEvent *PB.NewEventRequest) (*PB.Event, error) {
	row := db.Pool.QueryRow(context.Background(),
		`INSERT INTO events(author_id, synopsis, dttm) VALUE ($1, $2, $3) RETURNING id, synopsis, creation_dttm, dttm`,
		authorId, newEvent.Synopsis, newEvent.Time)
	var res PB.Event
	var creationTime time.Time
	var dttm time.Time
	err := row.Scan(&res.Id, &res.Synopsis, &creationTime, &dttm)
	res.CreationTime = timestamppb.New(creationTime)
	res.Time = timestamppb.New(dttm)
	return &res, err
}
