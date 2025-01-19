import { createContext, ReactNode, useState } from 'react'
import { join as joinPB } from 'join.proto'

const NAME_COOKIE_KEY: string = 'name';
const ID_UNSET: number = -1;
function getNameCookie(): number {
    let matches = document.cookie.match(new RegExp(
        "(?:^|; )" + NAME_COOKIE_KEY.replace(/([\.$?*|{}\(\)\[\]\\\/\+^])/g, '\\$1') + "=([^;]*)"
    ));
    return matches ? parseInt(decodeURIComponent(matches[1])) : ID_UNSET;
}
    
export const IdContext: React.Context<number> = createContext(ID_UNSET);

import { Plans } from './Plan';
import { Messenger } from './Message';

function AuthComponent({ handleAuth }: { handleAuth: (nickname: string) => void }): ReactNode {
    const [nickname, setNickname] = useState<string>('');
    return (
        <div className="auth">
            <input type="text" value={nickname} onChange={e => setNickname(e.target.value)} />
            <input type="submit" onClick={_ => handleAuth(nickname)} />
        </div>
    )
}

export default function App() {
    const [id, setId] = useState<number>(getNameCookie());
    
    if (id == ID_UNSET) {
        return (
            <AuthComponent handleAuth={
                (nickname: string) => 
                    fetch("http://0.0.0.0:5000/join", {
                        method: "POST",
                        headers: {
                            'Content-Type': 'application/json;charset=utf-8'
                        },
                        body: JSON.stringify(joinPB.JoinRequest.create({
                            username: nickname
                        }).toJSON()),
                    })
                    .then((response: Response) => response.arrayBuffer())
                    .then((buffer: ArrayBuffer) => joinPB.JoinResponse.decode(new Uint8Array(buffer)))
                    .then((res: joinPB.JoinResponse) =>setId(res.id))
            } />
        )
    }
    return (
        <IdContext.Provider value={id}>
            <div className="wrapper">
                <Plans />
                <Messenger />
            </div>
        </IdContext.Provider>
    )
};
