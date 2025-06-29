import { ReactNode, useState, useEffect, createContext } from 'react';
import { plan as planPB } from 'plan.proto'
import { API, APIContext } from 'App/lib/api';
import { PlanCreatorButton } from './PlanCreator'
import { convertScaleToString, downscale, upscale, getAgendaRoots, Agenda } from 'App/lib/util'
import { ScaleTree } from './Agenda';
import './Planer.module.css'

export const ScaleContext = createContext<planPB.TimeScale>(planPB.TimeScale.life);

export default function Planer({api}: {api: API}): ReactNode {
    const [agendaRoots, setAgendaRoots] = useState<Agenda[]>([]);
    const [scale, setScale] = useState<planPB.TimeScale>(planPB.TimeScale.life);
    useEffect(() => {
        const controller = new AbortController()
        api.getPlans({ signal: controller.signal })
            .then((res: planPB.UserPlans) => {
                setAgendaRoots(getAgendaRoots(res));
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
                        {
                            agendaRoots.map((a: Agenda) => <ScaleTree key={a.id} root={a} />)
                        }
                    </div>
                </div>
            </APIContext.Provider>
        </ScaleContext.Provider>
    )
}
