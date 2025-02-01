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
