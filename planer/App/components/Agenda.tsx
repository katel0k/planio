import { plan as planPB } from "plan.proto";
import { ReactNode, useContext } from "react";
import { ScaleContext } from "./Planer";

import "./Agenda.module.css"

type scaleTree = {
    body: {
        id: number,
        scale: planPB.TimeScale,
    } | null,
    subplans: scaleTree[]
};

export function ScaleTree({
    converter, 
    tree
}: {
    converter: (id: number) => ReactNode,
    tree: scaleTree[],
}): ReactNode {
    const scale = useContext(ScaleContext);

    return (
        tree.filter((p: scaleTree) => {
            if (p.body && p.body.scale == scale) {
                return true;
            }
            if (p.body && p.body.scale < scale) {
                return function findAppropriateScale(t: scaleTree): boolean {
                    return t.subplans.reduce((hasIt: boolean, tt: scaleTree) => 
                        hasIt || (tt.body && tt.body.scale == scale) || findAppropriateScale(tt), false);
                }(p);
            }
            return false;
        })
        .map((a: scaleTree) => a.body ?
            <div key={a.body.id} style={{paddingLeft: '30px'}}>
                {converter(a.body.id)}
                <ScaleTree converter={converter} tree={a.subplans}/>
            </div> : <></>)
    )
}
