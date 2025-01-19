import { createContext, useState } from 'react'
import { join as joinPB } from 'join.proto'
import './App.module.css'

const NAME_COOKIE_KEY: string = 'name';
const ID_UNSET: number = -1;
function getNameCookie(): number {
    let matches = document.cookie.match(new RegExp(
        "(?:^|; )" + NAME_COOKIE_KEY.replace(/([\.$?*|{}\(\)\[\]\\\/\+^])/g, '\\$1') + "=([^;]*)"
    ));
    return matches ? parseInt(decodeURIComponent(matches[1])) : ID_UNSET;
}

export const IdContext: React.Context<number> = createContext(ID_UNSET);

import Plans from 'App/components/Plan';
import Messenger from 'App/components/Messenger';
import Auth from 'App/components/Auth';

export default function App() {
    const [id, setId] = useState<number>(getNameCookie());
    
    if (id == ID_UNSET) {
        return (
            <Auth handleAuth={
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
