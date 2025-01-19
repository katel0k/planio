import { ReactNode, useState } from "react";

export default function Auth({ handleAuth }: { handleAuth: (nickname: string) => void }): ReactNode {
    const [nickname, setNickname] = useState<string>('');
    return (
        <div className="auth">
            <input type="text" value={nickname} onChange={e => setNickname(e.target.value)} />
            <input type="submit" onClick={_ => handleAuth(nickname)} />
        </div>
    )
}
