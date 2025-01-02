import { ReactNode } from 'react'
import planPB from './protos/plan.proto'

interface MessageProps {
    text: string,
    author: number
}

function MessageComponent({ text, author } : MessageProps): ReactNode {
    fetch("localhost:5000/").then()
    return  (
        <div className="message-wrapper">
            <div className="message">
                <div className="author"><span className="author">{author}</span></div>
                <div className="text"><span>{text}</span></div>
            </div>
        </div>
        )
}

export default function App() {
    return (
        <MessageComponent text="hi" author={1} />
    )
};
