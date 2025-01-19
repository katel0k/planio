import { ReactNode, useState, useEffect, useContext } from 'react';
import { plan as planPB } from 'plan.proto'
import { makeIdFetch, fetchFunc } from 'App/lib/serv';
import { IdContext } from 'App/App'
import './Plan.module.css'

interface PlanProps {
    synopsis: string,
    id: number,
    handleDelete: (plan: planPB.DeletePlanRequest) => void,
    handleChange: (plan: planPB.ChangePlanRequest) => void
}

function PlanComponent({ synopsis, id, handleChange, handleDelete }: PlanProps): ReactNode {
    const [synopsisInput, setSynopsisInput] = useState<string>(synopsis);
    const [isEditing, setIsEditing] = useState<boolean>(false);

    return (
        <div className="plan-wrapper">
            <div className="plan">
                <div className="plan-id-wrapper"><span className="plan-id">{id}</span></div>
                <div className="plan-synopsis-wrapper">
                    {isEditing ? 
                        <input className="plan-synopsis-editor" type="text"
                            value={synopsisInput}
                            onChange={e => setSynopsisInput(e.target.value)}
                            name="plan-synopsis-editor" /> :
                        <span className="plan-synopsis">{synopsis}</span>}
                </div>
                <div className="plan-settings-wrapper">
                    <div className="plan-settings">
                        <button className="plan-change"
                            onClick={_ => {
                                if (isEditing) {
                                    handleChange(planPB.ChangePlanRequest.create({
                                        id,
                                        synopsis: synopsisInput
                                    }));
                                    setIsEditing(false);
                                } else {
                                    setIsEditing(true);
                                }
                            }}>{isEditing ? 'save' : 'edit'}</button>
                        <button className="plan-delete" onClick={_ => 
                            handleDelete(planPB.DeletePlanRequest.create({id}))}>delete</button>
                    </div>
                </div>
            </div>
        </div>
    )
}

function PlanControls({ handleSubmit }: {
    handleSubmit: (synopsisValue: string) => void
}): ReactNode {
    const [synopsis, setSynopsis] = useState<string>('');
    return (
        <div className="plans-controls">
            <input className="plan-synopsis__text plans-controls__synopsis-input"
                   type="text" name="synopsis" onChange={e => setSynopsis(e.target.value)} />
            <input type="button" value="new plan" onClick={
                () => handleSubmit(synopsis)
            } />
        </div>
    )
}

export default function Plans(): ReactNode {
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
            <PlanControls handleSubmit={makeNewPlan} />
            <div className="plans-body-wrapper">
                <div className="plans-body">
                    {agenda.map((props: planPB.IPlan, index: number) =>
                        <PlanComponent
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
