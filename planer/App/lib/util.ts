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

export type agenda = {
    body: {
        id: number,
        scale: planPB.TimeScale
    } | null
    subplans: agenda[]
}

export function convertIAgendaToAgenda(agenda: planPB.IAgenda): agenda {
    let ag = new planPB.Agenda(agenda) as agenda;
    ag.body = ag.body ? new planPB.Agenda.AgendaNode(ag.body) : null;
    ag.subplans = ag.subplans.map(convertIAgendaToAgenda);
    console.log(ag);
    return ag;
}
