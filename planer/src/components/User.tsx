import { ReactNode } from "react"
import { join as joinPB } from 'join.proto'
import "./User.module.css"

interface UserProps {
    user: joinPB.IUser,
    isChosen: boolean,
    handleClick: (id: number | undefined) => void
}

export default function User({user, isChosen, handleClick}: UserProps): ReactNode {
    return (
        <div className={"user" + (isChosen ? " user_chosen" : "")} onClick={_ => handleClick(user.id ?? undefined)}>
            <span className="user-name">{user.username}</span>
            <span className="user-id">{user.id}</span>
        </div>
    )
}
