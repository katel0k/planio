import { ReactNode, useEffect, useState } from "react";
import "./PlanCreator.module.css"
import { plan as planPB } from "plan.proto";

export default function PlanCreator({ handleSubmit, handleCancel }: {
    handleSubmit: (request: planPB.NewPlanRequest) => void,
    handleCancel: () => void
}): ReactNode {
    const [synopsis, setSynopsis] = useState<string>('');
    const [description, setDescription] = useState<string>('');
    const [scale, setScale] = useState<planPB.TimeScale>(planPB.TimeScale.life);
    const [time, setTime] = useState<Date>(new Date());
    useEffect(() => {
        const timerId = setTimeout(() => {
            const timeCopy = new Date(time);
            timeCopy.setMinutes(time.getMinutes() + 1);
            setTime(timeCopy);
        }, 60 * 1000);
        return () => {
            clearTimeout(timerId);
        }
    }, [time]);
    const handleSave = () => {
        handleSubmit(planPB.NewPlanRequest.create({
            synopsis,
            description,
            scale
        }));
    }
    return (
        <div styleName="plan-creator">
            <input styleName="plan-creator__synopsis" type="text" name="synopsis"
                    onChange={e => setSynopsis(e.target.value)} value={synopsis} />
            <input styleName="plan-creator__description" type="text" name="description"
                    onChange={e => setDescription(e.target.value)} value={description} />
            <select styleName="plan-creator__scale" name="scale" value={scale} onChange={e => setScale(Number(e.target.value) as planPB.TimeScale)}>
                <option value={planPB.TimeScale.life}>Life</option>
                <option value={planPB.TimeScale.year}>Year</option>
                <option value={planPB.TimeScale.month}>Month</option>
                <option value={planPB.TimeScale.week}>Week</option>
                <option value={planPB.TimeScale.day}>Day</option>
                <option value={planPB.TimeScale.hour}>Hour</option>
            </select>
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
