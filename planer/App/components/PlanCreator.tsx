import { ReactNode, useContext, useEffect, useState } from "react";
import "./PlanCreator.module.css"
import { plan as planPB } from "plan.proto";
import { timeframe as timeframePB, google } from "timeframe.proto";
import { upscale } from "App/lib/util";
import { APIContext } from "App/lib/api";
import DatePicker from 'react-datepicker';
import "react-datepicker/dist/react-datepicker-cssmodules.css";

const TIME_UPDATE_TIMER: number = 60 * 1000;
const MISTAKE_VANISHING_TIMEOUT: number = 3 * 1000;

export default function PlanCreator({ handleSubmit, handleCancel, context }: {
    context?: planPB.Plan,
    handleSubmit: (request: planPB.NewPlanRequest) => void,
    handleCancel: () => void
}): ReactNode {
    const [synopsis, setSynopsis] = useState<string>('');
    const [description, setDescription] = useState<string>('');
    const [scale, setScale] = useState<planPB.TimeScale>(upscale(context?.scale ?? planPB.TimeScale.life));
    const [time, setTime] = useState<Date>(new Date());
    const [timeframeStart, setTimeframeStart] = useState<Date | null>(new Date());
    const [timeframeEnd, setTimeframeEnd] = useState<Date | null>(new Date());
    const [isTimed, setIsTimed] = useState<boolean>(false);
    const [mistake, setMistake] = useState<string | null>(null);

    useEffect(() => {
        const timerId = setTimeout(() => {
            const timeCopy = new Date(time);
            timeCopy.setMinutes(time.getMinutes() + 1);
            setTime(timeCopy);
        }, TIME_UPDATE_TIMER);
        return () => { clearTimeout(timerId) }
    }, [time]);

    useEffect(() => {
        if (mistake != null && mistake.length != 0) {
            const timerId = setTimeout(() => {
                setMistake(null);
            }, MISTAKE_VANISHING_TIMEOUT);
            return () => { clearTimeout(timerId) }
        }
        return;
    }, [mistake]);

    const handleSave = () => {
        if (synopsis.length == 0) {
            setMistake("Synopsis can't be empty");
            return;
        }
        handleSubmit(planPB.NewPlanRequest.create({
            synopsis,
            description,
            scale,
            parent: context?.id,
            timeframe: isTimed ? timeframePB.Timeframe.create({
                start: timeframeStart ? google.protobuf.Timestamp.fromObject(timeframeStart) : undefined,
                end: timeframeEnd ? google.protobuf.Timestamp.fromObject(timeframeEnd) : undefined
            }) : null
        }));
    }
    return (
        <div styleName="plan-creator">
            { mistake != null && <span>{mistake}</span> }
            <input styleName="plan-creator__synopsis" type="text" name="synopsis"
                    onChange={e => setSynopsis(e.target.value)} value={synopsis} />
            <input styleName="plan-creator__description" type="text" name="description"
                    onChange={e => setDescription(e.target.value)} value={description} />
            {
                context != undefined &&
                <select styleName="plan-creator__scale" name="scale" value={scale}
                        onChange={e => setScale(Number(e.target.value) as planPB.TimeScale)}>
                    <option value={planPB.TimeScale.life}>Life</option>
                    <option value={planPB.TimeScale.year}>Year</option>
                    <option value={planPB.TimeScale.month}>Month</option>
                    <option value={planPB.TimeScale.week}>Week</option>
                    <option value={planPB.TimeScale.day}>Day</option>
                    <option value={planPB.TimeScale.hour}>Hour</option>
                </select>
            }
            <div styleName="plan-creator__date">
                {
                    isTimed ? <>
                    <DatePicker
                        showIcon
                        selected={timeframeStart}
                        onChange={(date: Date | null) => setTimeframeStart(date)} />
                    - <DatePicker
                        showIcon
                        selected={timeframeEnd}
                        onChange={(date: Date | null) => setTimeframeEnd(date)} />
                        </> : <></>
                }
                timed <input type="checkbox" checked={isTimed} onChange={_ => setIsTimed(!isTimed)} />
            </div>
            <div styleName="plan-creator__time">
                Creation time: {time.toLocaleString('ru-ru', {
                    timeStyle: "short",
                    dateStyle: "short"
                })}
            </div>
            <div styleName="plan-creator__controls">
                <input styleName="plan-creator__cancel" type="button" value="Cancel" onClick={handleCancel} />
                <input styleName="plan-creator__submit" type="button" value="Save" onClick={handleSave} />
            </div>
        </div>
    )
}

export function PlanCreatorButton({ context }: { context?: planPB.Plan }): ReactNode {
    const [isPlanCreating, setIsPlanCreating] = useState<boolean>(false);
    const api = useContext(APIContext);
    return (
        isPlanCreating ?
            <PlanCreator handleSubmit={(request: planPB.NewPlanRequest) => {
                api?.createPlan(request);
                setIsPlanCreating(false);
            }} handleCancel={() => setIsPlanCreating(false)} context={context} /> :
            <input type="button" onClick={_ => setIsPlanCreating(true)} value="Plan" />
    )
}
