import { ReactNode, createContext} from 'react'

const id: number = await fetch("http://0.0.0.0:5000/join/artem").then(response => response.text()).then(parseInt);
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
