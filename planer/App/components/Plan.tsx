import { ReactNode, useState } from 'react';
import { plan as planPB } from 'plan.proto'
import "./Plan.module.css"

export interface PlanProps {
    synopsis: string,
    id: number,
    handleDelete: (plan: planPB.DeletePlanRequest) => void,
    handleChange: (plan: planPB.ChangePlanRequest) => void
}

export default function Plan({ synopsis, id, handleChange, handleDelete }: PlanProps): ReactNode {
    const [synopsisInput, setSynopsisInput] = useState<string>(synopsis);
    const [isEditing, setIsEditing] = useState<boolean>(false);

    return (
        <div className="plan-wrapper">
            <div className="plan">
                <div className="plan-id-wrapper"><span className="plan-id">{id}</span></div>
                <div className="plan-synopsis-wrapper">
                    {isEditing ? 
                        <input className="plan-synopsis-editor" type="text"
                            value={synopsisInput}
                            onChange={e => setSynopsisInput(e.target.value)}
                            name="plan-synopsis-editor" /> :
                        <span className="plan-synopsis">{synopsis}</span>}
                </div>
                <div className="plan-settings-wrapper">
                    <div className="plan-settings">
                        <button className="plan-change"
                            onClick={_ => {
                                if (isEditing) {
                                    handleChange(planPB.ChangePlanRequest.create({
                                        id,
                                        synopsis: synopsisInput
                                    }));
                                    setIsEditing(false);
                                } else {
                                    setIsEditing(true);
                                }
                            }}>{isEditing ? 'save' : 'edit'}</button>
                        <button className="plan-delete" onClick={_ => 
                            handleDelete(planPB.DeletePlanRequest.create({id}))}>delete</button>
                    </div>
                </div>
            </div>
        </div>
    )
}
