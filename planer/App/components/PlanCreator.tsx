import { ReactNode, useState } from "react";
import "./PlanCreator.module.css"

export default function PlanCreator({ handleSubmit }: {
    handleSubmit: (synopsisValue: string) => void
}): ReactNode {
    const [synopsis, setSynopsis] = useState<string>('');
    return (
        <div className="plans-controls">
            <input className="plan-synopsis__text plans-controls__synopsis-input"
                   type="text" name="synopsis" onChange={e => setSynopsis(e.target.value)} />
            <input type="button" value="new plan" onClick={
                () => handleSubmit(synopsis)
            } />
        </div>
    )
}
