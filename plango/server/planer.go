package main

import (
	"log"
	"net/http"

	planPB "github.com/katel0k/planio/server/build/plan"
	"google.golang.org/protobuf/proto"
)

func planHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		id, _ := getId(r)
		agenda, _ := r.Context().Value(DB).(Database).GetAllPlans(id)
		marsh, _ := proto.Marshal(agenda)
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
