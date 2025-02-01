import { ReactNode, useState, useEffect, useContext, createContext } from 'react';
import { plan as planPB } from 'plan.proto'
import { APIContext, apiFactory, IdContext } from 'App/lib/api';
import './Planer.module.css'
import Plan from './Plan'
import PlanCreator from './PlanCreator'
import { agendaTree, convertIPlanToPlan, convertScaleToString } from 'App/lib/util';
import Agenda, { ScaleTree } from './Agenda';

export const ScaleContext = createContext<planPB.TimeScale>(planPB.TimeScale.life);

export default function Planer(): ReactNode {
    const id = useContext<number>(IdContext);
    const api = apiFactory(id);
    const { getPlans, createPlan, changePlan, deletePlan } = api;
    const [agenda, setAgenda] = useState<agendaTree[]>([]);
    function findPlan(a: agendaTree[], id: number): agendaTree | null {
        return a.reduce((res: agendaTree | null, b: agendaTree) => res ?? (b.id == id ? b : findPlan(b.subplans, id)), null);
    }
    const [isPlanCreating, setIsPlanCreating] = useState<boolean>(false);
    const [scale, setScale] = useState<planPB.TimeScale>(planPB.TimeScale.life);
    useEffect(() => {
        const controller = new AbortController()
        getPlans({ signal: controller.signal })
            .then((res: planPB.Agenda) => setAgenda(res.plans.map(convertIPlanToPlan)))
            .catch(_ => {});
        return () => { controller.abort("Use effect cancelled") }
    }, []);

    const handleCreatePlan: (plan: planPB.NewPlanRequest) => void = (plan) => {
        createPlan(plan)
            .then((res: planPB.Plan) => setAgenda([...agenda, convertIPlanToPlan(res)]))
            .catch(_ => {});
    }

    const handleChangePlan: (change: planPB.ChangePlanRequest) => void = (change) => {
        changePlan(change)
            .then((res: planPB.Plan) => setAgenda(
                agenda.map((a: agendaTree) => a.id == res.id ? convertIPlanToPlan(res) : a)
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
        <ScaleContext.Provider value={scale}>
            <div styleName="planer">
                {
                    isPlanCreating ?
                        <PlanCreator handleSubmit={(request: planPB.NewPlanRequest) => {
                            handleCreatePlan(request);
                            setIsPlanCreating(false);
                        }} handleCancel={() => setIsPlanCreating(false)} /> :
                        <input type="button" onClick={_ => setIsPlanCreating(true)} value="Create new plan" />
                    }
                <div styleName="planer__controls">
                    <span>Scale: {convertScaleToString(scale)}</span>
                    <div>
                        <input type="button" name="planer__controls-zoom-in" value="in" onClick={_ => 
                            setScale(scale == planPB.TimeScale.hour ? scale : scale + 1)
                        }/>
                        <input type="button" name="planer__controls-zoom-out" value="out" onClick={_ => 
                            setScale(scale == planPB.TimeScale.life ? scale : scale - 1)
                        }/>
                    </div>
                </div>
                <div styleName="planer__plans">
                    <APIContext.Provider value={api}>
                        <ScaleTree converter={id => <Plan key={id} handleChange={_=>{}} handleDelete={_=>{}} plan={findPlan(agenda, id) as agendaTree}/>}
                            tree={agenda}/>
                        {/* <Agenda agenda={agenda}/> */}
                        {/* {agenda
                        .filter((plan: planPB.Plan) => plan.scale == scale)
                        .map((plan: planPB.Plan) =>
                            <Plan
                                plan={plan}
                                handleChange={handleChangePlan}
                                handleDelete={handleDeletePlan}
                                key={plan.id} />
                        )} */}
                    </APIContext.Provider>
                </div>
            </div>
        </ScaleContext.Provider>
    )
}
