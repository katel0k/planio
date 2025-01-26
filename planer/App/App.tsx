import { useState } from 'react'
import { auth as authPB } from 'auth.proto'
import './App.module.css'
import { IdContext, ID_UNSET, getNameCookie, serverUrl } from './lib/api';
import debugContext from './lib/debugContext';
import Planer from 'App/components/Planer';
import Messenger from 'App/components/Messenger';
import Auth from 'App/components/Auth';

export default function App() {
    const [id, setId] = useState<number>(getNameCookie());
    const [apiError, setApiError] = useState<string | null>(null);

    if (id == ID_UNSET) {
        return (
            <>
                {apiError && <div>{apiError}</div>}
                <Auth handleAuth={
                    (username: string) => 
                        fetch(new URL("/auth", serverUrl), {
                            method: "POST",
                            headers: {
                                'Content-Type': 'application/json;charset=utf-8'
                            },
                            body: JSON.stringify(authPB.AuthRequest.create({
                                username,
                                password: ""
                            }).toJSON()),
                        })
                        .then((response: Response) => response.arrayBuffer())
                        .then((buffer: ArrayBuffer) => authPB.AuthResponse.decode(new Uint8Array(buffer)))
                        .then((res: authPB.AuthResponse) => {
                            if (res.successful) {
                                setId(res.id as number);
                            } else {
                                setApiError(res.reason as string);
                            }
                        })
                } />
            </>
        )
    }
    return (
        <IdContext.Provider value={id}>
            <debugContext.Provider value={false}>
                <Planer />
                <Messenger />
            </debugContext.Provider>
        </IdContext.Provider>
    )
};
