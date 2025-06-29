import { ReactNode, useContext } from "react";
import { ScaleContext } from "./Planer";
import Plan from "./Plan";
import { PlanCreatorButton } from "./PlanCreator";
import { Agenda } from "App/lib/util";
import { plan as planPB } from "plan.proto";
import "./Agenda.module.css"

type AgendaState = Agenda & {
    show: boolean
};

function getShow(scale: planPB.TimeScale, ag: Agenda): boolean {
    if (ag.scale == scale) {
        return true;
    }
    if (ag.scale < scale) {
        return function findAppropriateScale(t: Agenda): boolean {
            return Array.from(t.subplans).reduce((hasIt: boolean, tt: Agenda) => 
                hasIt || tt.scale == scale || findAppropriateScale(tt), false);
        }(ag);
    }
    return false;
}

function convertAgendaToScaleTreeFactory(scale: planPB.TimeScale): (ag: Agenda) => AgendaState {
    return (ag: Agenda): AgendaState =>
        ({
            ...ag,
            show: getShow(scale, ag)
        })
}


export function ScaleTree({ root }: { root: Agenda, forceShow?: boolean }): ReactNode {
    const scale = useContext(ScaleContext);
    let ags: AgendaState[] = Array.from(root.subplans).map(convertAgendaToScaleTreeFactory(scale));

    return (((root.scale ?? planPB.TimeScale.life) <= scale) && 
        <div styleName="scale-tree_parented">
            <Plan plan={root} />
            <ScaleContext.Provider value={ root.scale }>
                <div styleName="scale-tree__subplans">
                    { ags
                        .filter((ag: AgendaState) => ag.show)
                        .map((ag: AgendaState) => <ScaleTree root={ag} key={ag.id} />) }
                    <PlanCreatorButton context={root}/>
                </div>
            </ScaleContext.Provider>
        </div>
    );
}
