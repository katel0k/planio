import { ReactNode, useState } from "react";
import "./Auth.module.css"

export default function Auth({ handleAuth, error }: 
        { handleAuth: (nickname: string) => void, error: string }): ReactNode {
    const [nickname, setNickname] = useState<string>('');
    return (
        <div className="auth">
            <div className="error">{error}</div>
            <input type="text" value={nickname} onChange={e => setNickname(e.target.value)} />
            <input type="submit" onClick={_ => handleAuth(nickname)} />
        </div>
    )
}
