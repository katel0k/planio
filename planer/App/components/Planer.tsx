import { ReactNode, useState, useEffect, useContext } from 'react';
import { plan as planPB } from 'plan.proto'
import { apiFactory, IdContext } from 'App/lib/api';
import './Planer.module.css'
import Plan from './Plan'
import PlanCreator from './PlanCreator'

export default function Planer(): ReactNode {
    const id = useContext<number>(IdContext);
    const { getPlans, createPlan, changePlan, deletePlan } = apiFactory(id);
    const [agenda, setAgenda] = useState<planPB.Plan[]>([]);
    const [isPlanCreating, setIsPlanCreating] = useState<boolean>(false);
    useEffect(() => {
        const controller = new AbortController()
        getPlans({ signal: controller.signal })
            .then((res: planPB.Agenda) => setAgenda(res.plans.map((a: planPB.IPlan) => new planPB.Plan(a))));
        return () => { controller.abort("Use effect cancelled") }
    }, []);

    const handleCreatePlan: (plan: planPB.NewPlanRequest) => void = (plan) => {
        createPlan(plan)
            .then((res: planPB.Plan) => setAgenda([...agenda, res]))
            .catch(_ => {});
    }

    const handleChangePlan: (change: planPB.ChangePlanRequest) => void = (change) => {
        changePlan(change)
            .then((res: planPB.Plan) => setAgenda(
                agenda.map((a: planPB.Plan) => a.id == res.id ? res : a)
            ))
            .catch(_ => {});
    }

    const handleDeletePlan: (del: planPB.DeletePlanRequest) => void = (del) => {
        deletePlan(del)
            .then((res: planPB.Plan) => setAgenda(
                agenda.filter((a: planPB.Plan) => a.id != res.id)
            ))
            .catch(_ => agenda);
    }

    return (
        <div styleName="planer">
            {
                isPlanCreating ?
                    <PlanCreator handleSubmit={(request: planPB.NewPlanRequest) => {
                        handleCreatePlan(request);
                        setIsPlanCreating(false);
                    }} handleCancel={() => setIsPlanCreating(false)} /> :
                    <button onClick={_ => setIsPlanCreating(true)}>Create new plan</button>
            }
            <div styleName="planer__plans">
                {agenda.map((plan: planPB.Plan) =>
                    <Plan
                        plan={plan}
                        handleChange={handleChangePlan}
                        handleDelete={handleDeletePlan}
                        key={plan.id} />
                )}
            </div>
        </div>
    )
}
