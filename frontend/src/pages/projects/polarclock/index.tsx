import Arc from "../../../components/Arc"
import Header from "../../../components/Header"
import React, { useEffect, useMemo, useState } from "react"


function getDayOfYear(timeObj: Date) : {yearAngle: number, daysPassed: number, days: number} {
    const year = timeObj.getFullYear()
    const isLeapYear = (year % 4 === 0 && year % 100 !== 0) || (year % 400 === 0)
    const days = isLeapYear? 366 : 365

    // calculate time passed since start of year
    const yearStart = new Date(year, 0, 1)
    const diff = timeObj.getTime() - yearStart.getTime()
    const daysPassed = diff / (1000 * 60 * 60 * 24)

    const yearAngle = (daysPassed / days) * 360
    // return the angle needed
    return {yearAngle, daysPassed, days}
}

const PolarClock = () => {
    const [time, setTime] = useState(new Date())
    const [hovered, setHovered] = useState<string | null>(null)

    useEffect(() => {
        const interval = setInterval(() => {
            setTime(new Date())
        }, 100)

        return () => clearInterval(interval) // Cleanup interval on component unmount
    }, [])

    console.log(time.getSeconds(), time.getMinutes(), time.getHours())

    const daysInMonth = new Date(time.getFullYear(), time.getMonth() + 1, 0).getDate()


    const secAng = (time.getSeconds() / 60) * 360
    const minAng = (time.getMinutes() / 60) * 360
    const hourAng = ((time.getHours() + time.getMinutes() / 60) / 24) * 360
    const monthAng = (((time.getDate() - 1 + (time.getHours() / 24)) / daysInMonth) * 360)
    const {yearAngle, daysPassed, days} = getDayOfYear(time)

    const hoveredArc = useMemo(() => {
        if (!hovered) return null
        switch (hovered) {
            case 'seconds':
                return { l: 's', c: time.getSeconds().toString(), t: 60 };
            case 'minutes':
                return { l: 'm', c: time.getMinutes().toString(), t: 60 };
            case 'hours':
                return { l: 'h', c: time.getHours().toString(), t: 24 };
            case 'day':
                return { l: ' day', c: (time.getDate() - 1 + time.getHours() / 24).toFixed(2), t: daysInMonth };
            case 'days':
                return { l: ' days', c: daysPassed.toFixed(2), t: days }
            default:
                return null;
        }
        // days, daysInMonth, and daysPassed are all currently calculated based on `time`, which updates at 10hz.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [hovered, time])

    return (
        <div className="min-h-screen grid grid-rows-[auto_1fr] bg-slate-400">
            <Header className="row-span-1" />
            <div className="w-full h-full row-span-1 flex flex-col justify-center items-center relative">
                <div className="h-[90%] aspect-square">
                    <svg width="100%" height="100%" viewBox="0 0 1000 1000" preserveAspectRatio="xMidYMid meet">

                        <Arc dev={true} r={250} sAng={0} eAng={secAng} width={40} color="red" omh={() => setHovered('seconds')} oml={() => setHovered(null)}/>

                        <Arc dev={true} r={210} sAng={0} eAng={minAng} width={40} color="blue" omh={() => setHovered('minutes')} oml={() => setHovered(null)}/>

                        <Arc dev={true} r={170} sAng={0} eAng={hourAng} width={40} color="green" omh={() => setHovered('hours')} oml={() => setHovered(null)}/>

                        <Arc dev={true} r={130} sAng={0} eAng={monthAng} width={40} color="purple" omh={() => setHovered('day')} oml={() => setHovered(null)}/>

                        <Arc dev={true} r={90} sAng={0} eAng={yearAngle} width={40} color="black" omh={() => setHovered('days')} oml={() => setHovered(null)} />
                    </svg>
                </div>
                {hoveredArc && (
                    <div className="absolute bottom-4 left-1/2 transform -translate-x-1/2 text-center bg-slate-500 bg-opacity-75 px-2 py-1 rounded-md">
                        {hoveredArc.c}/{hoveredArc.t}{hoveredArc.l}
                    </div>
                )}
            </div>
        </div>
    )
}

export default PolarClock