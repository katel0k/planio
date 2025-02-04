package main

import (
	"context"
	"log"
	"net/http"
	"time"

	PB "github.com/katel0k/planio/protos"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func planHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		id, _ := getId(r)
		userPlans, err := r.Context().Value(DB).(Database).GetAllPlans(id)
		if err != nil {
			log.Default().Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		marsh, _ := proto.Marshal(userPlans)
		w.Write(marsh)
	case "POST":
		defer r.Body.Close()
		id, _ := getId(r)
		var planReq PB.NewPlanRequest
		if err := getRequest(r, &planReq); err != nil {
			log.Default().Print(err)
			return
		}
		plan, _ := r.Context().Value(DB).(Database).CreateNewPlan(id, &planReq)
		marsh, _ := proto.Marshal(plan)
		w.Write(marsh)
	case "PATCH":
		defer r.Body.Close()
		var planReq PB.ChangePlanRequest
		if getRequest(r, &planReq) != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err := r.Context().Value(DB).(Database).ChangePlan(&planReq)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	case "DELETE":
		defer r.Body.Close()
		var planReq PB.DeletePlanRequest
		if getRequest(r, &planReq) != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err := r.Context().Value(DB).(Database).DeletePlan(int(planReq.Id))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func eventHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		id, _ := getId(r)
		events, _ := r.Context().Value(DB).(Database).GetEvents(id)
		calendar := PB.Calendar{
			Body: make([]*PB.Event, 0),
		}
		calendar.Body = append(calendar.Body, events...)
		marsh, _ := proto.Marshal(&calendar)
		w.Write(marsh)
	case "POST":
		defer r.Body.Close()
		id, _ := getId(r)
		var eventReq PB.NewEventRequest
		if err := getRequest(r, &eventReq); err != nil {
			log.Default().Print(err)
			return
		}
		plan, _ := r.Context().Value(DB).(Database).CreateEvent(id, &eventReq)
		marsh, _ := proto.Marshal(plan)
		w.Write(marsh)
	}
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
	ag, err := db.GetAgenda(userId)
	if err != nil {
		return nil, err
	}
	res.Structure = ag
	evs, err := db.GetEvents(userId)
	if err != nil {
		return nil, err
	}
	res.Calendar = &PB.Calendar{
		Body: evs,
	}
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
	ctx := context.TODO()
	conn, err := db.Pool.Acquire(ctx)

	if err != nil {
		return nil, err
	}

	row := conn.QueryRow(ctx,
		`INSERT INTO plans(author_id, synopsis, parent_id, scale) VALUES ($1, $2, $3, $4)
		RETURNING id, creation_dttm`,
		authorId, plan.Synopsis, plan.Parent, plan.Scale)
	var res PB.Plan
	var creationTime time.Time
	err = row.Scan(&res.Id, &creationTime)

	if err != nil {
		return nil, err
	}

	if plan.Timeframe != nil {
		conn.Query(ctx, `UPDATE plans SET start_dttm=$1, end_dttm=$2 WHERE id=$3`,
			plan.Timeframe.Start, plan.Timeframe.End, res.Id)
	}
	if plan.Description != nil && len(*plan.Description) != 0 {
		conn.Query(ctx, `INSERT INTO descriptions(plan_id, body) VALUES ($1, $2)`,
			res.Id, plan.Description)
	}

	res.CreationTime = timestamppb.New(creationTime)
	res.Description = plan.Description
	res.Scale = *plan.Scale
	res.Timeframe = plan.Timeframe
	res.Parent = plan.Parent
	return &res, nil
}

func (db Database) ChangePlan(plan *PB.ChangePlanRequest) error {
	ctx := context.TODO()
	_, err := db.Pool.Exec(ctx, "UPDATE plans SET synopsis=$1 WHERE id=$2", plan.Synopsis, plan.Id)
	return err
}

func (db Database) DeletePlan(plan_id int) error {
	ctx := context.TODO()
	_, err := db.Pool.Exec(ctx, "DELETE FROM plans WHERE id=$1", plan_id)
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
