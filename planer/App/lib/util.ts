import { plan as planPB } from 'plan.proto'

export function convertScaleToString(scale: planPB.TimeScale): string {
    switch (scale) {
        case planPB.TimeScale.life: return 'life';
        case planPB.TimeScale.year: return 'year';
        case planPB.TimeScale.month: return 'month';
        case planPB.TimeScale.week: return 'week';
        case planPB.TimeScale.day: return 'day';
        case planPB.TimeScale.hour: return 'hour';
        case planPB.TimeScale.unknown: return 'unknown';
    }
}

export function upscale(scale: planPB.TimeScale): planPB.TimeScale {
    return scale == planPB.TimeScale.hour ? scale : scale + 1;
}

export function downscale(scale: planPB.TimeScale): planPB.TimeScale {
    return scale == planPB.TimeScale.life ? scale : scale - 1;
}

export type PlanObject = Omit<planPB.Plan, "toJSON">;

export type Agenda = PlanObject & {
    subplans: SubAgenda
};

export type SubAgenda = {
    [id: number]: Agenda,
    ids: number[]
} & Iterable<Agenda>;

export function getAgendaRoots(userPlans: planPB.UserPlans): Agenda[] {
    function convertPBAgendaToAgenda(agenda: planPB.IAgenda): Agenda {
        const subplans: SubAgenda = {
            ...new planPB.Agenda(agenda).subplans.map(convertPBAgendaToAgenda).reduce(
                (res: {ids: number[]}, a: Agenda) => ({
                    ...res,
                    ids: res.ids.concat(a.id),
                    [a.id]: a
                }), { ids: [] }),
            [Symbol.iterator](): Iterator<Agenda> {
                let iter = this.ids[Symbol.iterator]();
                let self = this;
                return {
                    next() {
                        let res = iter.next();
                        return {
                            done: res.done,
                            value: self[res.value] as Agenda
                        }
                    }
                }
            }
        };
        let pl: planPB.Plan | undefined = userPlans.body.find((a: planPB.Plan) => a.id == agenda.body?.id);
        if (pl === undefined) {
            throw new Error("Plan from agenda was not found in plan list");
        }
        return {
            ...pl,
            subplans
        }
    }
    return userPlans.structure?.subplans.map(convertPBAgendaToAgenda) ?? [];
}
