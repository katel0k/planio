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
        <div styleName="plan">
            {/* <div styleName="plan-id-wrapper"><span styleName="plan-id">{id}</span></div> */}
            <div styleName="plan__synopsis-wrapper">
                {isEditing ? 
                    <input styleName="plan__synopsis-editor" type="text"
                        value={synopsisInput}
                        onChange={e => setSynopsisInput(e.target.value)}
                        name="plan__synopsis-editor" /> :
                    <span styleName="plan__synopsis">{synopsis}</span>}
            </div>
            <div styleName="plan__settings">
                <button styleName="plan__settings-change"
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
                <button styleName="plan__settings-delete" onClick={_ => 
                    handleDelete(planPB.DeletePlanRequest.create({id}))}>delete</button>
            </div>
        </div>
    )
}
