import { useState } from 'react'
import { join as joinPB } from 'join.proto'
import './App.module.css'
import IdContext, { ID_UNSET, getNameCookie } from './lib/api';
import Planer from 'App/components/Planer';
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
                <Planer />
                <Messenger />
            </div>
        </IdContext.Provider>
    )
};
