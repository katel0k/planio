import { ReactNode, useState, useEffect, createContext } from 'react';
import { plan as planPB } from 'plan.proto'
import { API, APIContext } from 'App/lib/api';
import './Planer.module.css'
import Plan from './Plan'
import { PlanCreatorButton } from './PlanCreator'
import { convertScaleToString, downscale, upscale } from 'App/lib/util'
import { ScaleTree } from './Agenda';

export const ScaleContext = createContext<planPB.TimeScale>(planPB.TimeScale.life);

export default function Planer({api}: {api: API}): ReactNode {
    const [agenda, setAgenda] = useState<planPB.Agenda[]>([]);
    type plansType = {
        [id: number]: planPB.Plan
    }
    const [plans, setPlans] = useState<plansType>({});
    const [scale, setScale] = useState<planPB.TimeScale>(planPB.TimeScale.life);
    useEffect(() => {
        const controller = new AbortController()
        api.getPlans({ signal: controller.signal })
            .then((res: planPB.UserPlans) => {
                setAgenda(res.structure ? res.structure.subplans : []);
                setPlans(res.body.map(a => new planPB.Plan(a)).reduce((res: plansType, a: planPB.Plan) => ({
                    ...res,
                    [a.id]: a
                }), {}));
            })
            .catch(_ => {});
            return () => { controller.abort("Use effect cancelled") }
        }, []);

    return (
        <ScaleContext.Provider value={scale}>
            <APIContext.Provider value={api}>
                <div styleName="planer">
                    <PlanCreatorButton />
                    <div styleName="planer__controls">
                        <span>Scale: {convertScaleToString(scale)}</span>
                        <div>
                            <input type="button" name="planer__controls-zoom-in" value="in" onClick={_ => 
                                setScale(upscale(scale))
                            }/>
                            <input type="button" name="planer__controls-zoom-out" value="out" onClick={_ => 
                                setScale(downscale(scale))
                            }/>
                        </div>
                    </div>
                    <div styleName="planer__plans">
                        <ScaleTree converter={id => id in plans ?
                            <Plan
                                handleChange={api.changePlan}
                                handleDelete={api.deletePlan}
                                plan={plans[id] as planPB.Plan}
                                key={id}/> : null}
                            tree={agenda}/>
                    </div>
                </div>
            </APIContext.Provider>
        </ScaleContext.Provider>
    )
}
