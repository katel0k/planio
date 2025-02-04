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
