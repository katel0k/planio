import { ReactNode, useState, useEffect, useContext } from 'react';
import planPB from 'plan.proto'
import { makeIdFetch } from './serv';
import { IdContext } from './App'
import 'plan.css'

interface PlanProps {
    synopsis: string,
    id: number
}

function PlanComponent({ synopsis, id }: PlanProps): ReactNode {
    return (
        <div className="plan-wrapper">
            <div className="plan">
                <div className="plan-id-wrapper"><span className="plan-id">{id}</span></div>
                <div className="synopsis-wrapper"><span className="synopsis">{synopsis}</span></div>
            </div>
        </div>
    )
}

function PlanControls({ handleSubmit }: {
    handleSubmit: (synopsisValue: string) => void
}): ReactNode {
    const [synopsis, setSynopsis] = useState('');
    return (
        <div className="plans-controls">
            <input className="plan-synopsis__text plans-controls__synopsis-input" type="text" name="synopsis" onChange={e => setSynopsis(e.target.value)} />
            <input type="button" value="new plan" onClick={
                () => handleSubmit(synopsis)
            } />
        </div>
    )
}

export function Plans(): ReactNode {
    const id = useContext(IdContext);
    const f = makeIdFetch(id);
    const [agenda, setAgenda] = useState<planPB.plan.IPlan[]>([]);
    useEffect(() => {
        const controller = new AbortController()
        const signal = controller.signal
        f("http://0.0.0.0:5000/plans", {headers:{}, signal})
            .then(response => response.arrayBuffer())
            .then(buffer => planPB.plan.Agenda.decode(new Uint8Array(buffer)))
            .then(res => setAgenda(res.plans))
            .catch(_ => {})
        return () => { controller.abort("Use effect cancelled") }
    }, []);

    function makeNewPlan(synopsis: string) {
        const plan = planPB.plan.Plan.create({
            synopsis
        });
        const f = makeIdFetch(id);
        f("http://0.0.0.0:5000/plan", {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json;charset=utf-8'
            },
            body: JSON.stringify(plan.toJSON()),
        })
            .then(response => response.arrayBuffer())
            .then(buffer => planPB.plan.Plan.decode(new Uint8Array(buffer)))
            .then(res => setAgenda([res, ...agenda]))
            .catch(_ => {})
    }

    return (
        <div className="plans">
            <PlanControls handleSubmit={makeNewPlan} />
            <div className="plans-body-wrapper">
                <div className="plans-body">
                    {agenda.map((props, index) =>
                        <PlanComponent
                            synopsis={props.synopsis ?? ''}
                            id={props.id ?? 0}
                            key={index} />
                    )}
                </div>
            </div>
        </div>
    )
}