import { useCallback, useState } from 'react'
import { auth as authPB } from 'auth.proto'
import './App.module.css'
import { IdContext, ID_UNSET, handleAuth, apiFactory, API } from './lib/api';
import debugContext from './lib/debugContext';
import Planer from 'App/components/Planer';
import Messenger from 'App/components/Messenger';
import Auth from 'App/components/Auth';

export default function App({ id, setId }: { id: number, setId: (id: number) => void }) {
    const [apiError, setApiError] = useState<string>("");
    const handleOnAuth = useCallback((username: string) => {
        handleAuth(authPB.AuthRequest.create({
            username,
            password: ""
        }))
        .then((res: authPB.AuthResponse) => {
            if (res.successful) {
                setId(res.id as number);
                setApiError("");
            } else {
                setApiError(res.reason as string);
            }
        });
    }, []);
    const api: API = apiFactory(id);
    return (id == ID_UNSET) ?
        <Auth handleAuth={handleOnAuth} error={apiError}/> : 
        <IdContext.Provider value={id}>
            <debugContext.Provider value={false}>
                <Planer api={api}/>
                <Messenger />
            </debugContext.Provider>
        </IdContext.Provider>
};
