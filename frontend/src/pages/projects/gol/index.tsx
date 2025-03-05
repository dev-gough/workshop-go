import React, { useState, useRef, useEffect } from 'react'
import Header from '../../../components/Header'

const GRID_SIZE = 4

type GridCell = 0 | 1
type Grid = GridCell[][]
type ChangedCell = { row: number, col: number }

const GameOfLife: React.FC = () => {
    // Refs for canvas and offscreen canvas
    const canvasRef = useRef<HTMLCanvasElement>(null)
    const offscreenCanvasRef = useRef<HTMLCanvasElement | null>(null)
    const gridStateRef = useRef<Grid | null>(null)

    // State for UI elements
    const [generation, setGeneration] = useState(0)
    const [isRunning, setIsRunning] = useState(false)
    const [numRows, setNumRows] = useState(0)
    const [numCols, setNumCols] = useState(0)
    const [showModal, setShowModal] = useState(false)
    const [patterns, setPatterns] = useState<string[]>([])
    const [selectedPattern, setSelectedPattern] = useState<string | null>(null)
    const [tickRate, setTickRate] = useState(50)

    // Initialize canvas and handle resize
    useEffect(() => {
        const canvas = canvasRef.current
        if (!canvas) return

        const offscreenCanvas = document.createElement('canvas')
        offscreenCanvasRef.current = offscreenCanvas

        const handleResize = () => {
            const header = document.querySelector('header')
            const bottomButtons = canvas.nextElementSibling
            if (!header || !bottomButtons) return

            const availableHeight = window.innerHeight - header.offsetHeight - (bottomButtons as HTMLElement).offsetHeight
            const availableWidth = window.innerWidth

            canvas.width = availableWidth
            canvas.height = availableHeight
            offscreenCanvas.width = availableWidth
            offscreenCanvas.height = availableHeight

            const newNumRows = Math.floor(availableHeight / GRID_SIZE)
            const newNumCols = Math.floor(availableWidth / GRID_SIZE)
            setNumRows(newNumRows)
            setNumCols(newNumCols)
            gridStateRef.current = createEmptyGrid(newNumRows, newNumCols)
            drawGridToOffscreen()
        }

        window.addEventListener('resize', handleResize)
        handleResize() // Initial call
        return () => window.removeEventListener('resize', handleResize)
    }, [])

    // Game loop
    useEffect(() => {
        let intervalId: NodeJS.Timeout | undefined
        if (isRunning && tickRate > 0) {
            intervalId = setInterval(updateGrid, 1000 / tickRate)
        }
        return () => clearInterval(intervalId)
    }, [isRunning, tickRate, numRows, numCols])

    // Utility functions
    const createEmptyGrid = (rows: number, cols: number): Grid => {
        return Array.from({ length: rows }, () => new Array(cols).fill(0))
    }

    const getCellState = (row: number, col: number): GridCell => {
        const wrappedRow = (row + numRows) % numRows
        const wrappedCol = (col + numCols) % numCols
        return gridStateRef.current?.[wrappedRow]?.[wrappedCol] ?? 0
    }

    const neighborOffsets: [number, number][] = [
        [-1, -1], [-1, 0], [-1, 1],
        [0, -1], [0, 1],
        [1, -1], [1, 0], [1, 1],
    ]

    const countLiveNeighbors = (row: number, col: number): number => {
        let count = 0
        for (const [dx, dy] of neighborOffsets) {
            if (getCellState(row + dx, col + dy)) {
                count++
                if (count >= 4) return count // Early termination
            }
        }
        return count
    }

    const updateGrid = () => {
        if (!gridStateRef.current) return

        const currentGrid = gridStateRef.current
        const newGrid = createEmptyGrid(numRows, numCols)
        const changedCells: ChangedCell[] = []

        for (let row = 0; row < numRows; row++) {
            for (let col = 0; col < numCols; col++) {
                const liveNeighbors = countLiveNeighbors(row, col)
                if (liveNeighbors < 2 || liveNeighbors > 3) {
                    newGrid[row][col] = 0
                } else if (liveNeighbors === 3) {
                    newGrid[row][col] = 1
                } else {
                    newGrid[row][col] = currentGrid[row][col]
                }
                if (newGrid[row][col] !== currentGrid[row][col]) {
                    changedCells.push({ row, col })
                }
            }
        }

        gridStateRef.current = newGrid
        setGeneration(g => g + 1)
        drawGridToOffscreen(changedCells)
    }

    const drawGridToOffscreen = (changedCells: ChangedCell[] = []) => {
        const offscreenCanvas = offscreenCanvasRef.current
        const canvas = canvasRef.current
        if (!offscreenCanvas || !canvas) return

        const offscreenCtx = offscreenCanvas.getContext('2d')
        const ctx = canvas.getContext('2d')
        if (!offscreenCtx || !ctx) return

        if (changedCells.length === 0) {
            offscreenCtx.clearRect(0, 0, canvas.width, canvas.height)
            for (let row = 0; row < numRows; row++) {
                for (let col = 0; col < numCols; col++) {
                    offscreenCtx.fillStyle = gridStateRef.current?.[row]?.[col] ? 'black' : 'white'
                    offscreenCtx.fillRect(col * GRID_SIZE, row * GRID_SIZE, GRID_SIZE, GRID_SIZE)
                }
            }
        } else {
            for (const { row, col } of changedCells) {
                for (let i = -1; i <= 1; i++) {
                    for (let j = -1; j <= 1; j++) {
                        const r = (row + i + numRows) % numRows
                        const c = (col + j + numCols) % numCols
                        offscreenCtx.fillStyle = gridStateRef.current?.[r]?.[c] ? 'black' : 'white'
                        offscreenCtx.fillRect(c * GRID_SIZE, r * GRID_SIZE, GRID_SIZE, GRID_SIZE)
                    }
                }
            }
        }
        ctx.drawImage(offscreenCanvas, 0, 0)
    }

    // Event handlers
    const handleCanvasClick = (event: React.MouseEvent<HTMLCanvasElement>) => {
        const canvas = canvasRef.current
        if (!canvas || !gridStateRef.current) return

        const rect = canvas.getBoundingClientRect()
        const x = event.clientX - rect.left
        const y = event.clientY - rect.top
        const row = Math.floor(y / GRID_SIZE)
        const col = Math.floor(x / GRID_SIZE)

        gridStateRef.current[row][col] = (1 - gridStateRef.current[row][col]) as GridCell
        drawGridToOffscreen([{ row, col }])
    }

    const startGame = () => setIsRunning(true)
    const stopGame = () => setIsRunning(false)
    const resetGame = () => {
        gridStateRef.current = createEmptyGrid(numRows, numCols)
        setGeneration(0)
        setIsRunning(false)
        setSelectedPattern(null)
        setTickRate(50)
        drawGridToOffscreen()
    }

    const handleLoadPatterns = async () => {
        try {
            const res = await fetch('/api/gol/patterns')
            if (!res.ok) throw new Error('Failed to fetch patterns')
            const files: string[] = await res.json()
            setPatterns(files)
            setShowModal(true)
        } catch (e) {
            console.error('Error fetching patterns:', e)
        }
    }

    const handlePatternLoad = async () => {
        if (!selectedPattern) {
            alert('Please select a pattern')
            return
        }
        try {
            const res = await fetch(`/api/gol/patterns/${selectedPattern}`)
            if (!res.ok) throw new Error('Failed to fetch pattern')
            const data = await res.json()
            const content = data.contents

            const patternBlocks = parseLifeFile(content)
            const minGrid = calculateMinGrid(patternBlocks, 5)
            gridStateRef.current = createGameState(minGrid, numRows, numCols)
            setGeneration(0)
            drawGridToOffscreen()
            setShowModal(false)
        } catch (e) {
            console.error('Error loading pattern:', e)
        }
    }

    // Pattern parsing functions
    const parseLifeFile = (content: string) => {
        const normContent = content.replace(/\r\n|\r/g, "\n")
        const lines = normContent.split("\n")
        const patternBlocks: { x: number, y: number, pattern: string[] }[] = []
        let currentBlock: { x: number, y: number, pattern: string[] } | null = null

        lines.forEach((line) => {
            line = line.trim()
            if (line.startsWith("#")) {
                if (line.startsWith("#P")) {
                    if (currentBlock) patternBlocks.push(currentBlock)
                    const [, x, y] = line.split(" ").map(Number)
                    currentBlock = { x, y, pattern: [] }
                }
            } else if (line && currentBlock) {
                currentBlock.pattern.push(line)
            }
        })
        if (currentBlock) patternBlocks.push(currentBlock)
        return patternBlocks
    }

    const calculateMinGrid = (patternBlocks: { x: number, y: number, pattern: string[] }[], padding: number) => {
        let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity
        patternBlocks.forEach((block) => {
            const blockStartX = block.x
            const blockStartY = block.y
            block.pattern.forEach((line, yIndex) => {
                line.split("").forEach((char, xIndex) => {
                    if (char === "*") {
                        const x = blockStartX + xIndex
                        const y = blockStartY + yIndex
                        minX = Math.min(minX, x)
                        minY = Math.min(minY, y)
                        maxX = Math.max(maxX, x)
                        maxY = Math.max(maxY, y)
                    }
                })
            })
        })
        return {
            minX: minX - padding,
            minY: minY - padding,
            maxX: maxX + padding,
            maxY: maxY + padding,
            patternBlocks,
        }
    }

    const createGameState = (minGrid: any, rows: number, cols: number): Grid => {
        const grid = createEmptyGrid(rows, cols)
        const patternWidth = minGrid.maxX - minGrid.minX + 1
        const patternHeight = minGrid.maxY - minGrid.minY + 1

        // Center the pattern by calculating offsets, ensuring negative coords are shifted into grid
        const offsetX = Math.floor((cols - patternWidth) / 2) - minGrid.minX
        const offsetY = Math.floor((rows - patternHeight) / 2) - minGrid.minY

        minGrid.patternBlocks.forEach((block: { x: number, y: number, pattern: string[] }) => {
            const startX = block.x + offsetX
            const startY = block.y + offsetY
            block.pattern.forEach((line, yIndex) => {
                line.split('').forEach((char, xIndex) => {
                    if (char === '*') {
                        const x = startX + xIndex
                        const y = startY + yIndex
                        // Only set cells within grid boundaries
                        if (x >= 0 && x < cols && y >= 0 && y < rows) {
                            grid[y][x] = 1
                        }
                    }
                })
            })
        })
        return grid
    }

    return (
        <div className="h-screen w-screen flex flex-col overflow-hidden">
            <Header projectName='Game of Life' />
            <div className="flex-grow flex flex-col items-center justify-center">
                <canvas
                    ref={canvasRef}
                    className="border border-black m-0 p-0 w-full h-full"
                    onClick={handleCanvasClick}
                />
                <div className="flex w-full p-4 flex-shrink-0 bg-gray-200">
                    <div className="w-1/3 flex justify-start">
                        <div className="flex space-x-2">
                            <button className="px-4 py-2 hover:bg-opacity-100 border border-gray-700 bg-opacity-50 bg-green-400 rounded-md" onClick={startGame}>Start</button>
                            <button className="px-4 py-2 hover:bg-opacity-100 border border-gray-700 bg-opacity-50 bg-red-600 rounded-md" onClick={stopGame}>Stop</button>
                            <button className="px-4 py-2 hover:bg-opacity-100 border border-gray-700 bg-opacity-50 bg-orange-400 rounded-md" onClick={resetGame}>Reset</button>
                        </div>
                    </div>
                    <div className="flex items-center space-x-4">
                        <div className="w-16 text-right text-md text-gray-600">{tickRate} TPS</div>
                        <input
                            type="range"
                            min={0}
                            max={60}
                            value={tickRate}
                            onChange={(e) => setTickRate(Number(e.target.value))}
                            className="w-32"
                        />
                        <div className="relative inline-block">
                            <button
                                className="px-4 py-2 border border-gray-700 hover:bg-gray-400 bg-gray-300 rounded-md"
                                onClick={handleLoadPatterns}
                            >
                                Load Patterns
                            </button>
                            {selectedPattern && (
                                <span className="absolute top-1/2 left-full ml-2 transform -translate-y-1/2 text-sm text-gray-600 whitespace-nowrap">
                                    {selectedPattern}
                                </span>
                            )}
                        </div>
                    </div>
                    <div className="w-1/3 flex justify-end">
                        <div className="text-xl w-40 text-right whitespace-nowrap">Generation: {generation}</div>
                    </div>
                </div>
                {showModal && (
                    <div className="absolute inset-0 bg-gray-800 bg-opacity-75 flex items-center justify-center">
                        <div className="bg-white p-6 rounded-lg max-w-4xl w-full">
                            <h2 className="text-2xl mb-4">Select a Pattern</h2>
                            <div className="grid sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
                                {patterns.map((pattern) => (
                                    <div
                                        key={pattern}
                                        className={`p-4 border border-gray-300 rounded text-center cursor-pointer ${selectedPattern === pattern ? 'bg-gray-200' : ''
                                            }`}
                                        onClick={() => setSelectedPattern(pattern)}
                                    >
                                        {pattern}
                                    </div>
                                ))}
                            </div>
                            <button
                                className="mt-4 px-4 py-2 bg-green-500 text-white rounded"
                                onClick={handlePatternLoad}
                            >
                                Load
                            </button>
                            <button
                                className="mt-4 ml-2 px-4 py-2 bg-red-500 text-white rounded"
                                onClick={() => setShowModal(false)}
                            >
                                Close
                            </button>
                        </div>
                    </div>
                )}
            </div>
        </div>
    )
}

export default GameOfLife