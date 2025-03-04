import React, { useEffect, useRef} from "react"

interface ArcProps {
    r: number;
    sAng: number;
    eAng: number;
    width: number;
    color: string;
    dev?: boolean;
    omh?: () => void; // onMouseHover
    oml?: () => void; // onMouseLeave
    // these are shortened to keep arc components on a single (albeit long) line
}

const Arc: React.FC<ArcProps> = ({ r, sAng, eAng, width, color, dev, omh, oml }) => {
    const ref = useRef<SVGSVGElement>(null)

    useEffect(() => {
        const updateDim = () => {
            if (ref.current?.parentElement) {
                const { offsetWidth, offsetHeight } = ref.current.parentElement
                console.log(offsetWidth, offsetHeight)
            }
        }

        updateDim()
        window.addEventListener('resize', updateDim)
        return () => window.removeEventListener('resize', updateDim)
    }, [])

    const polarToCartesian = (centerX: number, centerY: number, radius: number, angleInDegrees: number) => {
        const angleInRadians = (angleInDegrees - 90) * (Math.PI / 180)
        return {
            x: centerX + radius * Math.cos(angleInRadians),
            y: centerY + radius * Math.sin(angleInRadians),
        }
    }

    const describeArc = (x: number, y: number, r: number, sAng: number, eAng: number) => {
        const start = polarToCartesian(x, y, r, eAng)
        const end = polarToCartesian(x, y, r, sAng)
        const largeArcFlag = eAng - sAng <= 180 ? '0' : '1'

        return [
            `M ${start.x} ${start.y}`,
            `A ${r} ${r} 0 ${largeArcFlag} 0 ${end.x} ${end.y}`,
        ].join(' ')
    }
    //todo: change to be dynamic
    const centerX = 500
    const centerY = 500

    return (
            <path
                d={describeArc(centerX, centerY, r, sAng, eAng)}
                fill="none"
                stroke={color}
                strokeWidth={width}
                preserveAspectRatio="none"
                onMouseOver={omh}
                onMouseOut={oml} />
    )
}

export default Arc