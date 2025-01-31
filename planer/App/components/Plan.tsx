import { ReactNode, useContext, useState } from 'react';
import { plan as planPB } from 'plan.proto'
import "./Plan.module.css"
import debugContext from 'App/lib/debugContext';
import { convertScaleToString } from 'App/lib/util';
import PlanCreator from './PlanCreator';
import { APIContext } from 'App/lib/api';
import { ScaleContext } from './Planer';

export interface PlanProps {
    plan: planPB.Plan,
    handleDelete: (plan: planPB.DeletePlanRequest) => void,
    handleChange: (plan: planPB.ChangePlanRequest) => void
}

export default function Plan({ plan, handleChange, handleDelete }: PlanProps): ReactNode {
    const [synopsis, setSynopsis] = useState<string>(plan.synopsis);
    const [isEditing, setIsEditing] = useState<boolean>(false);
    const [isCreatingSubplan, setIsCreatingSubplan] = useState<boolean>(false);
    const debug = useContext(debugContext);
    const api = useContext(APIContext);
    const scale = useContext(ScaleContext);

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
                    <button styleName="plan__settings-change"
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
                        }}>{isEditing ? 'save' : 'edit'}</button>
                    <button styleName="plan__settings-delete" onClick={_ => 
                        handleDelete(planPB.DeletePlanRequest.create({id: plan.id}))}>delete</button>
                    <button styleName="plan__settings-subplan" onClick={_ => setIsCreatingSubplan(true)}>subplan</button>
                </div>
            </div>
            <div styleName="plan__subplans">
                {isCreatingSubplan && <PlanCreator 
                    handleSubmit={(request: planPB.NewPlanRequest) => {
                        request.parent = plan.id;
                        api?.createPlan(request);
                        setIsCreatingSubplan(false);
                    }}
                    handleCancel={() => setIsCreatingSubplan(false)}
                    />}
                {
                    plan.subplans
                    .map((p: planPB.IPlan) => new planPB.Plan(p))
                    .filter((p: planPB.Plan) => p.scale == scale)
                    .map((p: planPB.Plan) => <Plan plan={p} key={p.id} handleChange={_=>{}} handleDelete={_=>{}} />)
                }
            </div>
        </div>
    )
}
