import { ReactNode, useState, useEffect } from 'react'
import planPB from 'plan.proto'

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

function Plans(): ReactNode {
    const [agenda, setAgenda] = useState<planPB.plan.IAgenda>({});
    useEffect(() => {
        fetch("http://0.0.0.0:5000/plans")
            .then(response => response.arrayBuffer())
            .then(buffer => planPB.plan.Agenda.decode(new Uint8Array(buffer)))
            .then(res => setAgenda(res))
    });

    return (
        <div className="plans">
            {agenda.plans?.map((props, index) =>
                <PlanComponent
                    synopsis={props.synopsis ?? ''}
                    id={props.id ?? 0}
                    key={index} />
            )}
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
