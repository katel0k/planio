import { ReactNode, useContext, useEffect, useState } from "react"
import { join as joinPB } from 'join.proto'
import { msg as msgPB } from 'msg.proto'
import { IdContext } from './App'

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

function Messages({ user }: {user: joinPB.IUser | undefined}): ReactNode {
    const id = useContext<number>(IdContext);
    const [messages, setMessages] = useState<msgPB.IMsgResponse[]>([]);
    useEffect(() => {
        const controller = new AbortController()
        const signal = controller.signal
        fetch("http://localhost:5000/message", {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json;charset=utf-8'
            },
            body: JSON.stringify(msgPB.AllMessagesRequest.create({
                senderId: id,
                receiverId: user?.id ?? 0
            })),
            signal
        })
            .then((response: Response) => response.arrayBuffer())
            .then((buffer: ArrayBuffer) => msgPB.AllMessagesResponse.decode(new Uint8Array(buffer)))
            .then((resp: msgPB.AllMessagesResponse) => setMessages(resp.messages))
            .catch(_ => {})
        return () => { controller.abort('Use effect cancelled') }
    }, []);
    return (
        <div className="messages">
            {
                messages.map((a: msgPB.IMsgResponse, i: number) =>
                    <MessageComponent text={a.text ?? ''} author={a.authorId ?? 0} key={i}/>)
            }
        </div>
    )
}

interface UserProps {
    user: joinPB.IUser,
    isChosen: boolean,
    handleClick: (id: number | undefined) => void
}

function UserComponent({user, isChosen, handleClick}: UserProps): ReactNode {
    return (
        <div className={"user" + (isChosen ? " user_chosen" : "")} onClick={_ => handleClick(user.id ?? undefined)}>
            <span className="user-name">{user.username}</span>
            <span className="user-id">{user.id}</span>
        </div>
    )
}

export function Messenger(): ReactNode {
    const [chosenUser, setChosenUser] = useState<number | undefined>(0);
    const [userList, setUserList] = useState<joinPB.IUser[]>([]);

    useEffect(() => {
        const controller = new AbortController()
        const signal = controller.signal
        fetch("http://localhost:5000/join", {
            method: 'GET',
            signal
        })
            .then((response: Response) => response.arrayBuffer())
            .then((buffer: ArrayBuffer) => joinPB.JoinedUsersResponse.decode(new Uint8Array(buffer)))
            .then((userList: joinPB.IJoinedUsersResponse) => setUserList(userList.users ?? []))
            .catch(_ => {})
        return () => { controller.abort('Use effect cancelled') }
    }, []);

    function handleUserChoosing(id: number | undefined) {
        setChosenUser(id);
    }

    return (
        <div className="messenger">
            <div className="user-list">
                {userList.map((a: joinPB.IUser, i: number) => 
                    <UserComponent user={a} isChosen={chosenUser == i} handleClick={handleUserChoosing} key={i} />
                    )}
            </div>
            <Messages user={chosenUser ? userList[chosenUser] : undefined} />
        </div>
    )
}
