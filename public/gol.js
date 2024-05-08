const ROWS = 40;
const SMALL = 50;
const MEDIUM = 70;
const LARGE = 90;
const gameBoard = document.getElementById('gameBoard');
const MAX_STEPS = 10000;
let isDragging = false;

function timeFunctionExecution(func) {
    return function (...args) {
        const start = performance.now(); // Start timing before executing the function
        const result = func.apply(this, args); // Execute the function with its arguments
        const end = performance.now(); // End timing after executing the function
        if (end - start > 1) console.log(`Execution time for ${func}: ${end - start} ms`); // Log the execution time
        return result; // Return the original function's result
    };
}


if (ROWS < SMALL) GRIDSPACE = 20;
else if (ROWS >= MEDIUM && ROWS < LARGE) GRIDSPACE = 10;
else GRIDSPACE = 5;

function handleClick(e) {
    tar = document.getElementById(e.target.id);
    tar.style.backgroundColor = tar.style.backgroundColor === 'white' ? 'black' : 'white';
}

const cells = document.querySelectorAll('.cell');
const container = document.getElementById('gameBoard');

container.addEventListener('mousedown', (e) => {
    // Prevent the default behavior of selecting text
    e.preventDefault();
    e.stopPropagation();
    isDragging = true;
});

container.addEventListener('mousemove', (e) => {
    // Get the cell that the mouse is currently over
    const cell = e.target.closest('.cell');

    // If the cell exists and has not already been highlighted, highlight it
    if (isDragging && cell && !cell.classList.contains('highlighted')) {
        cell.style.backgroundColor = 'black'; // or any other color you prefer
    }
});

// Set up the mouseup event to clear the highlighting when the drag ends
container.addEventListener('mouseup', () => {
    isDragging = false;
});

function createGrid() {
    gameBoard.innerHTML = ''; // Clear the previous grid
    gameBoard.style.display = 'grid';
    gameBoard.style.gridTemplateColumns = `repeat(${ROWS}, ${GRIDSPACE}px)`; // Set the number of columns
    gameBoard.style.cursor = 'pointer';
    document.getElementById('step-label').innerText = `Step 0 / ${MAX_STEPS}`;

    // Rows, then cols.  Sets up id in form {row}-{col} - e.g. 0-24 is top right square
    for (let i = 0; i < ROWS; i++) {
        for (let j = 0; j < ROWS; j++) {
            let cell = document.createElement('div');
            cell.id = `${i}-${j}`;
            cell.classList.add(ROWS < MEDIUM ? 'grid-item-s' : (ROWS < LARGE ? 'grid-item-m' : 'grid-item-l'));
            cell.classList.add('cell', 'border', 'border-gray-200');
            cell.style.backgroundColor = 'white'; // Default color
            gameBoard.appendChild(cell);
            cell.onclick = handleClick;
        }
    }
}

function boundCoords(x, y) {
    return [
        Math.min(Math.max(x, 0), ROWS - 1),
        Math.min(Math.max(y, 0), ROWS - 1)
    ];
}

function handleReset() {
    createGrid();
    setTimeout(() => {
        document.getElementById('step-label').innerText = `Step 0 / ${MAX_STEPS}`;
    }, 10);
}

function delay(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

function getNeighbours(cell) {
    let [row, col] = cell.id.split('-').map(Number);
    let neighbours = new Set();

    // Get the 1-neighbourhood of the cell
    for (let i = -1; i <= 1; i++) {
        for (let j = -1; j <= 1; j++) {
            if (i == 0 && j == 0) continue;
            let [nRow, nCol] = boundCoords(row + i, col + j);
            if (nRow == row && nCol == col) continue;
            neighbours.add(`${nRow}-${nCol}`);
            cell = document.getElementById(`${nRow}-${nCol}`);
        }
    }
    return neighbours;
}

function getNeighbourhood(cellId, n) {
    const [row, col] = cellId.split('-').map(Number);
    const neighbourhood = new Set();

    for (let i = -n; i <= n; i++) {
        for (let j = -n; j <= n; j++) {
            if (i === 0 && j === 0) continue; // skip the center cell
            const coords = boundCoords(row + i, col + j);
            if (coords) {
                neighbourhood.add(`${coords[0]}-${coords[1]}`);
            }
        }
    }

    return neighbourhood;
}

async function gameLoop() {
    let step = 0;

    while (step < MAX_STEPS) {

        document.getElementById('step-label').innerText = `Step: ${step} / ${MAX_STEPS}`;
        var totalAlive = Array.from(document.getElementsByClassName('cell')).filter(cell => cell.style.backgroundColor == 'black').length;
        if (totalAlive == 0) break;

        var cells = Array.from(document.getElementsByClassName('cell'));
        for (let cell of cells) {
            let [row, col] = cell.id.split('-').map(Number);
            let neighbours = getNeighbours(cell);
            let alive = Array.from(neighbours).filter(n => document.getElementById(n).style.backgroundColor == 'black').length;
            let markedForDeath = (alive >= 4 || alive <= 1);
            let markerForBirth = alive == 3;


            if (markedForDeath) cell.classList.add('marked-dead');
            if (markerForBirth) cell.classList.add('marked-birth');
        }

        var marked_dead = Array.from(document.getElementsByClassName('marked-dead'));
        var marked_birth = Array.from(document.getElementsByClassName('marked-birth'));

        for (let cell of marked_dead) {
            cell.style.backgroundColor = 'white';
            cell.classList.remove('marked-dead');
        }

        for (let cell of marked_birth) {
            cell.style.backgroundColor = 'black';
            cell.classList.remove('marked-birth');
        }
        step++;
        await delay(20);
    }
}
document.getElementById('resetbtn').onclick = handleReset;
createGrid();