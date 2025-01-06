import { ReactNode } from "react"

interface MessageProps {
    text: string,
    author: number
}

export default function MessageComponent({ text, author }: MessageProps): ReactNode {
    return (
        <div className="message-wrapper">
            <div className="message">
                <div className="author"><span>{author}</span></div>
                <div className="text"><span>{text}</span></div>
            </div>
        </div>
    )
}
