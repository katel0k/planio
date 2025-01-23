import { ReactNode, useState, useEffect, useContext } from 'react';
import { plan as planPB } from 'plan.proto'
import { apiFactory, IdContext } from 'App/lib/api';
import './Plan.module.css'
import Plan from './Plan'
import PlanCreator from './PlanCreator'

export default function Planer(): ReactNode {
    const id = useContext<number>(IdContext);
    const { getPlans, createPlan, changePlan, deletePlan } = apiFactory(id);
    const [agenda, setAgenda] = useState<planPB.IPlan[]>([]);
    useEffect(() => {
        const controller = new AbortController()
        const signal = controller.signal
        getPlans({ signal })
            .then((res: planPB.Agenda) => setAgenda(res.plans));
        return () => { controller.abort("Use effect cancelled") }
    }, []);

    const handleCreatePlan: (synopsis: string) => void = (synopsis) => {
        createPlan(synopsis)
            .then((res: planPB.Plan) => setAgenda([res, ...agenda]))
            .catch(_ => {});
    }

    const handleChangePlan: (change: planPB.ChangePlanRequest) => void = (change) => {
        changePlan(change)
            .then((res: planPB.Plan) => setAgenda(
                agenda.map((a: planPB.IPlan) => a.id == res.id ? res : a)
            ))
            .catch(_ => {});
    }

    const handleDeletePlan: (del: planPB.DeletePlanRequest) => void = (del) => {
        deletePlan(del)
            .then((res: planPB.Plan) => setAgenda(
                agenda.filter((a: planPB.IPlan) => a.id != res.id)
            ))
            .catch(_ => agenda);
    }

    return (
        <div styleName="plans">
            <PlanCreator handleSubmit={handleCreatePlan} />
            <div styleName="plans-body">
                {agenda.map((props: planPB.IPlan, index: number) =>
                    <Plan
                        synopsis={props.synopsis ?? ''}
                        handleChange={handleChangePlan}
                        handleDelete={handleDeletePlan}
                        id={props.id ?? 0}
                        key={index} />
                )}
            </div>
        </div>
    )
}
