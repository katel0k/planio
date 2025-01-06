import { ReactNode, useState, useEffect, createContext, useContext } from 'react'
import planPB from 'plan.proto'
import { makeIdFetch } from './serv';

const id = await fetch("http://0.0.0.0:5000/join/artem").then(response => response.text()).then(parseInt);
console.log(id);
const IdContext = createContext(id);

interface MessageProps {
    text: string,
    author: number
}

function MessageComponent({ text, author }: MessageProps): ReactNode {
    return (
        <div className="message-wrapper">
            <div className="message">
                <div className="author"><span>{author}</span></div>
                <div className="text"><span>{text}</span></div>
            </div>
        </div>
    )
}

interface PlanProps {
    synopsis: string,
    id: number
}

function PlanComponent({ synopsis, id }: PlanProps): ReactNode {
    return (
        <div className="plan-wrapper">
            <div className="plan">
                <div className="id"><span>{id}</span></div>
                <div className="synopsis"><span>{synopsis}</span></div>
            </div>
        </div>
    )
}

function PlanControls({ handleSubmit }: {
    handleSubmit: (synopsisValue: string) => void
}): ReactNode {
    const [synopsis, setSynopsis] = useState('');
    return (
        <div className="plans-control">
            <input type="text" name="synopsis" onChange={e => setSynopsis(e.target.value)} />
            <input type="button" value="new plan" onClick={
                () => handleSubmit(synopsis)
            } />
        </div>
    )
}

function Plans(): ReactNode {
    const id = useContext(IdContext);
    const f = makeIdFetch(id);
    const [agenda, setAgenda] = useState<planPB.plan.IAgenda>({});
    useEffect(() => {
        f("http://0.0.0.0:5000/plans", {headers:{}})
            .then(response => response.arrayBuffer())
            .then(buffer => planPB.plan.Agenda.decode(new Uint8Array(buffer)))
            .then(res => setAgenda(res))
    }, []);

    function makeNewPlan(synopsis: string) {
        const plan = planPB.plan.Plan.create({
            synopsis
        });
        console.log(id);
        const f = makeIdFetch(id);
        f("http://0.0.0.0:5000/plan", {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json;charset=utf-8'
            },
            body: JSON.stringify(plan.toJSON()),
        }).then(response => {
            console.log(response);
        })
    }

    return (
        <div className="plans">
            <PlanControls handleSubmit={makeNewPlan} />
            <div className="plans-body">
                {agenda.plans?.map((props, index) =>
                    <PlanComponent
                        synopsis={props.synopsis ?? ''}
                        id={props.id ?? 0}
                        key={index} />
                )}
            </div>
        </div>
    )
}

export default function App() {

    return (
        <div className="wrapper">
            <Plans />
            <div className="messages">
                <MessageComponent text={'Hi:)'} author={3} />
            </div>
        </div>
    )
};
