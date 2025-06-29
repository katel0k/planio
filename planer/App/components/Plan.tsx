import { ReactNode, useContext, useState } from 'react';
import { plan as planPB } from 'plan.proto'
import debugContext from 'App/lib/debugContext';
import { convertScaleToString, PlanObject } from 'App/lib/util';
import "./Plan.module.css"

export interface PlanProps {
    plan: PlanObject
}

export default function Plan({ plan }: PlanProps): ReactNode {
    let handleChange = (_: planPB.ChangePlanRequest) => {};
    let handleDelete = (_: planPB.DeletePlanRequest) => {};
    const [synopsis, setSynopsis] = useState<string>(plan.synopsis);
    const [isEditing, setIsEditing] = useState<boolean>(false);
    const debug = useContext(debugContext);

    return (
        <div styleName="plan">
            { debug && <div><span>{plan.id}</span></div> }
            <div styleName="plan__body">
                <div styleName="plan__info">
                    {isEditing ? 
                        <input styleName="plan__synopsis-editor" type="text"
                            value={synopsis}
                            onChange={e => setSynopsis(e.target.value)}
                            name="plan__synopsis-editor" /> :
                        <span styleName="plan__synopsis">{synopsis}</span>}
                    <div styleName="plan__description">{plan.description ?? ""}</div>
                    <div styleName="plan__time-scale">{convertScaleToString(plan.scale)}</div>
                </div>
                <div styleName="plan__settings">
                    <input type="button" styleName="plan__settings-change"
                        onClick={_ => {
                            if (isEditing) {
                                handleChange(planPB.ChangePlanRequest.create({
                                    id: plan.id,
                                    synopsis
                                }));
                                setIsEditing(false);
                            } else {
                                setIsEditing(true);
                            }
                        }} value={isEditing ? 'save' : 'edit'} />
                    <input type="button" styleName="plan__settings-delete" onClick={_ => 
                        handleDelete(planPB.DeletePlanRequest.create({id: plan.id}))} value="delete" />
                </div>
            </div>
        </div>
    )
}
