import { ReactNode } from 'react'

interface MessageProps {
    text: string,
    author: number
}

function Message({ text, author } : MessageProps): ReactNode {
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
        <Message text="hi" author={1} />
    )
};
