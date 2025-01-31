import { plan as planPB } from "plan.proto";
import { Context, createContext, PropsWithChildren, ReactNode, useContext, useState } from "react";
import { ScaleContext } from "./Planer";
import Plan from "./Plan";

import "./Agenda.module.css"

export const PlanDBContext: Context<Map<number, planPB.Plan>> = createContext(new Map());

export default function Agenda({agenda, parent}: {agenda: planPB.Plan[], parent?: planPB.Plan}): ReactNode {
    // const [agenda, setAgenda] = useState<planPB.Plan[]>(agendaProp);

    const scale = useContext(ScaleContext);
    return (
        <div styleName={parent == undefined ? "agenda" : "agenda_parented"}>
            {
                parent == undefined ? "" : <Plan plan={parent} handleChange={_=>{}} handleDelete={_=>{}}/>
            }
            {
                agenda.filter((p: planPB.Plan) => {
                    let doRender: boolean = p.scale <= scale || p.subplans.reduce((hasSubplans: boolean, pl: planPB.IPlan) =>
                        hasSubplans || new planPB.Plan(pl).scale >= scale, false);
                    console.log(p.synopsis, doRender);
                    return doRender;
                }).map((p: planPB.Plan) => 
                    p.scale == scale ?
                    <Plan plan={p} key={p.id} handleChange={_=>{}} handleDelete={_=>{}}/> :
                    <Agenda agenda={p.subplans.map(a => new planPB.Plan(a))} parent={p} key={p.id}/>)
            }
        </div>
    )
}

type scaleTree = {
    id: number,
    scale: planPB.TimeScale,
    subplans: scaleTree[]
};

export function ScaleTree({
    converter, 
    tree
}: {
    converter: (id: number) => ReactNode,
    tree: scaleTree[],
} & PropsWithChildren): ReactNode {
    const scale = useContext(ScaleContext);

    return (
        tree.filter((p: scaleTree) => {
            if (p.scale == scale) {
                return true;
            }
            if (p.scale < scale) {
                return function findAppropriateScale(t: scaleTree): boolean {
                    return t.subplans.reduce((hasIt: boolean, tt: scaleTree) => 
                        hasIt || tt.scale == scale || findAppropriateScale(tt), false);
                }(p);
            }
            return false;
        })
        .map(({ id }: scaleTree) => converter(id))
    )
}
