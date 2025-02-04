import { ReactNode, useState, useEffect, useContext, createContext } from 'react';
// import { event as eventPB } from 'event.proto'
import { plan as planPB } from 'plan.proto'
import { APIContext, apiFactory, IdContext } from 'App/lib/api';
import './Planer.module.css'
import Plan from './Plan'
import { PlanCreatorButton } from './PlanCreator'
import { convertScaleToString, downscale, upscale } from 'App/lib/util'
import { ScaleTree } from './Agenda';

export const ScaleContext = createContext<planPB.TimeScale>(planPB.TimeScale.life);

export default function Planer(): ReactNode {
    const id = useContext<number>(IdContext);
    const api = apiFactory(id);
    const { getPlans } = api;
    const [agenda, setAgenda] = useState<planPB.Agenda[]>([]);
    const [plans, setPlans] = useState<Map<number, planPB.Plan>>(new Map());
    // const [events, setEvents] = useState<eventPB.Event[]>([]);
    const [scale, setScale] = useState<planPB.TimeScale>(planPB.TimeScale.life);
    useEffect(() => {
        const controller = new AbortController()
        getPlans({ signal: controller.signal })
            .then((res: planPB.UserPlans) => {
                setAgenda(res.structure ? res.structure.subplans : []);
                setPlans(new Map(res.body.map(a => new planPB.Plan(a)).map((a: planPB.Plan) => [a.id, a])));
                // setEvents(res.events);
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
                        <ScaleTree converter={id => plans.has(id) ?
                            <Plan
                                handleChange={api?.changePlan}
                                handleDelete={api?.deletePlan}
                                plan={plans.get(id) as planPB.Plan}/> : null}
                            tree={agenda}/>
                    </div>
                </div>
            </APIContext.Provider>
        </ScaleContext.Provider>
    )
}
