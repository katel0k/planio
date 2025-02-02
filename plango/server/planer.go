package main

import (
	"log"
	"net/http"

	eventPB "github.com/katel0k/planio/server/build/event"
	planPB "github.com/katel0k/planio/server/build/plan"
	"google.golang.org/protobuf/proto"
)

func planHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		id, _ := getId(r)
		userPlans, _ := r.Context().Value(DB).(Database).GetAllPlans(id)
		agenda, _ := r.Context().Value(DB).(Database).GetAgenda(id)
		userPlans.Structure = agenda
		marsh, _ := proto.Marshal(userPlans)
		w.Write(marsh)
	case "POST":
		defer r.Body.Close()
		id, _ := getId(r)
		var planReq planPB.NewPlanRequest
		if err := getRequest(r, &planReq); err != nil {
			log.Default().Print(err)
			return
		}
		plan, _ := r.Context().Value(DB).(Database).CreateNewPlan(id, &planReq)
		marsh, _ := proto.Marshal(plan)
		w.Write(marsh)
	case "PATCH":
		defer r.Body.Close()
		var planReq planPB.ChangePlanRequest
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
		var planReq planPB.DeletePlanRequest
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
		calendar := eventPB.Calendar{
			Body: make([]*eventPB.Event, 0),
		}
		calendar.Body = append(calendar.Body, events...)
		marsh, _ := proto.Marshal(&calendar)
		w.Write(marsh)
	case "POST":
		defer r.Body.Close()
		id, _ := getId(r)
		var eventReq eventPB.NewEventRequest
		if err := getRequest(r, &eventReq); err != nil {
			log.Default().Print(err)
			return
		}
		plan, _ := r.Context().Value(DB).(Database).CreateEvent(id, &eventReq)
		marsh, _ := proto.Marshal(plan)
		w.Write(marsh)
	}
}
