const canvas = document.getElementById("gameCanvas");
const ctx = canvas.getContext("2d");

const gridSize = 4;
let numRows = Math.floor(canvas.height / gridSize);
let numCols = Math.floor(canvas.width / gridSize);
let tickRate = 50; // Ticks per second (adjust as needed)
let intervalId;

let gridState = createEmptyGrid();
let newGridState = createEmptyGrid();
let changedCells = [];

let generation = 0;
const generationCounterElement =
    document.getElementById("generationCounter");

const offscreenCanvas = document.createElement("canvas");
offscreenCanvas.width = canvas.width;
offscreenCanvas.height = canvas.height;
const offscreenCtx = offscreenCanvas.getContext("2d");

for (let row = 0; row < numRows; row++) {
    for (let col = 0; col < numCols; col++) {
        changedCells.push([row, col]);
    }
}

function createEmptyGrid() {
    return Array.from({ length: numRows }, () =>
        new Array(numCols).fill(0),
    );
}

function getCellState(row, col) {
    const wrappedRow = (row + numRows) % numRows;
    const wrappedCol = (col + numCols) % numCols;
    return gridState[wrappedRow][wrappedCol];
}
const neighborOffsets = [
    [-1, -1],
    [-1, 0],
    [-1, 1],
    [0, -1],
    [0, 1],
    [1, -1],
    [1, 0],
    [1, 1],
];

function countLiveNeighbors(row, col) {
    let count = 0;
    for (const [dx, dy] of neighborOffsets) {
        if (getCellState(row + dx, col + dy)) {
            count++;
            if (count >= 4) return count; // Early termination
        }
    }
    return count;
}

function updateGrid() {
    const changedCells = []; // Track changed cells

    for (let row = 0; row < numRows; row++) {
        for (let col = 0; col < numCols; col++) {
            let liveNeighbors = countLiveNeighbors(row, col);
            if (liveNeighbors < 2 || liveNeighbors > 3) {
                newGridState[row][col] = 0; // Cell dies
            } else if (liveNeighbors === 3) {
                newGridState[row][col] = 1; // Cell is born
            } else {
                newGridState[row][col] = gridState[row][col];
            }

            // Check if the cell state has changed
            if (newGridState[row][col] !== gridState[row][col]) {
                changedCells.push({ row, col }); // Store the changed cell coordinates
            }
        }
    }

    [gridState, newGridState] = [newGridState, gridState];
    generation++;
    generationCounterElement.textContent = "Generation: " + generation;
    drawGridToOffscreen(changedCells); // Pass changed cells to drawGridToOffscreen
}

function handleResize() {
    const canvas = document.getElementById('gameCanvas');
    const header = document.querySelector('header');
    const bottomButtons = canvas.nextElementSibling;

    // Calculate the available height, taking into account the header and bottom buttons
    const availableHeight = window.innerHeight - header.offsetHeight - bottomButtons.offsetHeight;
    const availableWidth = window.innerWidth;

    // Set the canvas size directly to the available space (no scrollbars)
    canvas.style.height = `${availableHeight}px`;
    canvas.style.width = `${availableWidth}px`;

    // Update the canvas width and height attributes for drawing (must be integers)
    canvas.width = availableWidth;
    canvas.height = availableHeight;

    // If using an offscreen canvas or other drawing logic, update accordingly
    if (offscreenCanvas) {
        offscreenCanvas.width = availableWidth;
        offscreenCanvas.height = availableHeight;
        numRows = Math.floor(availableHeight / gridSize);
        numCols = Math.floor(availableWidth / gridSize);
        gridState = createEmptyGrid();
        newGridState = createEmptyGrid();
        drawGridToOffscreen();
        ctx.drawImage(offscreenCanvas, 0, 0);
    }
}

// Call handleResize when the window resizes
window.addEventListener('resize', handleResize);
document.addEventListener('DOMContentLoaded', handleResize);

function drawGridToOffscreen(changedCells = []) {
    // If no changed cells are provided, redraw the entire grid (e.g., initial draw)
    if (changedCells.length === 0) {
        offscreenCtx.clearRect(
            0,
            0,
            offscreenCanvas.width,
            offscreenCanvas.height,
        );
        for (let row = 0; row < numRows; row++) {
            for (let col = 0; col < numCols; col++) {
                offscreenCtx.fillStyle = gridState[row][col] ? "black" : "white";
                offscreenCtx.fillRect(
                    col * gridSize,
                    row * gridSize,
                    gridSize,
                    gridSize,
                );
            }
        }
    } else {
        // Otherwise, only redraw the changed cells and their neighbors
        for (const { row, col } of changedCells) {
            for (let i = -1; i <= 1; i++) {
                for (let j = -1; j <= 1; j++) {
                    const r = (row + i + numRows) % numRows; // Wrap around for edges
                    const c = (col + j + numCols) % numCols;
                    offscreenCtx.fillStyle = gridState[r][c] ? "black" : "white";
                    offscreenCtx.fillRect(
                        c * gridSize,
                        r * gridSize,
                        gridSize,
                        gridSize,
                    );
                }
            }
        }
    }

    // Copy to the visible canvas
    ctx.drawImage(offscreenCanvas, 0, 0);
}

function startGame() {
    intervalId = setInterval(updateGrid, 1000 / tickRate);
}

function stopGame() {
    clearInterval(intervalId);
}

// (Optional) Add buttons or UI elements to start/stop the game and change the tick rate
canvas.addEventListener("click", (event) => {
    const rect = canvas.getBoundingClientRect();
    const x = event.clientX - rect.left;
    const y = event.clientY - rect.top;
    const row = Math.floor(y / gridSize);
    const col = Math.floor(x / gridSize);

    // Toggle the cell state and redraw
    gridState[row][col] = 1 - gridState[row][col];
    offscreenCtx.fillStyle = gridState[row][col] ? "black" : "white";
    offscreenCtx.fillRect(
        col * gridSize,
        row * gridSize,
        gridSize,
        gridSize,
    );

    // Copy the updated offscreen canvas to the visible canvas
    ctx.drawImage(offscreenCanvas, 0, 0);
});

drawGridToOffscreen();
ctx.drawImage(offscreenCanvas, 0, 0);

document
    .getElementById("fileInput")
    .addEventListener("change", handleFile, false);


function handleFile(event) {
    const file = event.target.files[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = function (e) {
        gridState = createEmptyGrid();
        newGridState = createEmptyGrid();
        const content = e.target.result;
        const patternBlocks = parseLifeFile(content);
        const minGrid = calculateMinGrid(patternBlocks, 5);

        gridState = createGameState(minGrid, numRows, numCols); // Define the game board size
        drawGridToOffscreen();
        ctx.drawImage(offscreenCanvas, 0, 0);
    };
    reader.readAsText(file);
}

function parseLifeFile(content) {
    const lines = content.split("\n");
    const patternBlocks = [];
    let currentBlock = null;

    lines.forEach((line) => {
        line = line.trim();
        if (line.startsWith("#")) {
            if (line.startsWith("#P")) {
                if (currentBlock) {
                    patternBlocks.push(currentBlock);
                }
                const [, x, y] = line.split(" ").map(Number);
                currentBlock = { x, y, pattern: [] };
            }
        } else if (line) {
            if (currentBlock) {
                currentBlock.pattern.push(line);
            }
        }
    });

    if (currentBlock) {
        patternBlocks.push(currentBlock);
    }

    return patternBlocks;
}

function calculateMinGrid(patternBlocks, padding) {
    let minX = Infinity,
        minY = Infinity,
        maxX = -Infinity,
        maxY = -Infinity;

    patternBlocks.forEach((block) => {
        const blockStartX = block.x;
        const blockStartY = block.y;

        block.pattern.forEach((line, yIndex) => {
            line.split("").forEach((char, xIndex) => {
                if (char === "*") {
                    const x = blockStartX + xIndex;
                    const y = blockStartY + yIndex;
                    if (x < minX) minX = x;
                    if (y < minY) minY = y;
                    if (x > maxX) maxX = x;
                    if (y > maxY) maxY = y;
                }
            });
        });
    });

    // Add padding
    minX -= padding;
    minY -= padding;
    maxX += padding;
    maxY += padding;

    return { minX, minY, maxX, maxY, patternBlocks };
}

function createGameState(minGrid, width, height) {
    drawGridToOffscreen();
    ctx.drawImage(offscreenCanvas, 0, 0);
    const offsetX = -minGrid.minX;
    const offsetY = -minGrid.minY;

    minGrid.patternBlocks.forEach((block) => {
        const startX = block.x + offsetX;
        const startY = block.y + offsetY;

        block.pattern.forEach((line, yIndex) => {
            line.split("").forEach((char, xIndex) => {
                if (char === "*") {
                    const x = startX + xIndex;
                    const y = startY + yIndex;
                    if (x >= 0 && x < width && y >= 0 && y < height) {
                        if (!gridState[y]) gridState[y] = []; // Ensure the row exists
                        gridState[y][x] = 1;
                    }
                }
            });
        });
    });

    return gridState;
}