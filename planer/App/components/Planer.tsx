import { ReactNode, useState, useEffect, useContext } from 'react';
import { plan as planPB } from 'plan.proto'
import { makeIdFetch, fetchFunc } from 'App/lib/api';
import IdContext from 'App/lib/api'
import './Plan.module.css'
import Plan from './Plan'
import PlanCreator from './PlanCreator'

export default function Planer(): ReactNode {
    const id = useContext<number>(IdContext);
    const f: fetchFunc = makeIdFetch(id);
    const [agenda, setAgenda] = useState<planPB.IPlan[]>([]);
    useEffect(() => {
        const controller = new AbortController()
        const signal = controller.signal
        f("http://0.0.0.0:5000/plan", { signal })
            .then((response: Response) => response.arrayBuffer())
            .then((buffer: ArrayBuffer) => planPB.Agenda.decode(new Uint8Array(buffer)))
            .then((res: planPB.Agenda) => setAgenda(res.plans))
            .catch(_ => {})
        return () => { controller.abort("Use effect cancelled") }
    }, []);

    function makeNewPlan(synopsis: string) {
        const plan = planPB.Plan.create({
            synopsis
        });
        const f: fetchFunc = makeIdFetch(id);
        f("http://0.0.0.0:5000/plan", {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json;charset=utf-8'
            },
            body: JSON.stringify(plan.toJSON()),
        })
            .then((response: Response) => response.arrayBuffer())
            .then((buffer: ArrayBuffer) => planPB.Plan.decode(new Uint8Array(buffer)))
            .then((res: planPB.Plan) => setAgenda([res, ...agenda]))
            .catch(_ => {})
    }

    function changePlan(plan: planPB.ChangePlanRequest) {
        const f: fetchFunc = makeIdFetch(id);
        f("http://0.0.0.0:5000/plan", {
            method: 'PATCH',
            headers: {
                'Content-Type': 'application/json;charset=utf-8'
            },
            body: JSON.stringify(plan.toJSON())
        })
            .then((response: Response) => {
                if (response.ok) {
                    setAgenda(agenda.map((a: planPB.IPlan) => 
                        a.id == plan.id ? planPB.Plan.fromObject({...a, synopsis: plan.synopsis}) : a));
                } else {
                    alert("error");
                }
            })
            .catch(_ => {})
    }

    function deletePlan(plan: planPB.DeletePlanRequest) {
        const f: fetchFunc = makeIdFetch(id);
        f("http://0.0.0.0:5000/plan", {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json;charset=utf-8'
            },
            body: JSON.stringify(plan.toJSON())
        })
            .then((response: Response) => {
                if (response.ok) {
                    setAgenda(agenda.filter((a: planPB.IPlan) => a.id != plan.id))
                } else {
                    alert("error");
                }
            })
            .catch(_ => {})
    }

    return (
        <div className="plans">
            <PlanCreator handleSubmit={makeNewPlan} />
            <div className="plans-body-wrapper">
                <div className="plans-body">
                    {agenda.map((props: planPB.IPlan, index: number) =>
                        <Plan
                            synopsis={props.synopsis ?? ''}
                            handleChange={changePlan}
                            handleDelete={deletePlan}
                            id={props.id ?? 0}
                            key={index} />
                    )}
                </div>
            </div>
        </div>
    )
}
