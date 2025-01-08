import { ReactNode, createContext} from 'react'
import { join as joinPB } from 'join.proto'

const id: number =
    await fetch("http://0.0.0.0:5000/join/artem")
            .then((response: Response) => response.arrayBuffer())
            .then((buffer: ArrayBuffer) => joinPB.JoinResponse.decode(new Uint8Array(buffer)))
            .then((res: joinPB.JoinResponse) => res.id)
export const IdContext: React.Context<number> = createContext(id);

import { Plans } from './Plan'

function AppState(): ReactNode {
    return (
        <></>
    )
}

function Navbar(): ReactNode {
    return (
        <nav>
            <AppState />
        </nav>
    )
}

export default function App() {

    return (
        <div className="wrapper">
            <Navbar />
            <Plans />
        </div>
    )
};
