const ROWS = 40;
const SMALL = 50;
const MEDIUM = 70;
const LARGE = 90;
const gameBoard = document.getElementById('gameBoard');
const MAX_STEPS = 10000;
let isDragging = false;
let STEP_SPEED = 100;
let savedStates = [];
const LOCAL_STORAGE_KEY = 'gameOfLifePatterns';

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
    document.getElementById('states').innerHTML = '';
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
    STEP_SPEED = document.getElementById('step-speed').value;

    while (step < MAX_STEPS) {

        document.getElementById('step-label').innerText = `Step: ${step} / ${MAX_STEPS}`;
        var cells = Array.from(gameBoard.querySelectorAll('.cell'));
        var totalAlive = cells.filter(cell => cell.style.backgroundColor == 'black').length;
        if (totalAlive == 0) break;


        for (let cell of cells) {
            // let [row, col] = cell.id.split('-').map(Number);
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
        await delay(STEP_SPEED);
    }
}

function saveGameState() {
    const gameState = {
        gridSize: ROWS,
        cells: []
    };

    const cells = gameBoard.querySelectorAll('.cell'); // Select all cells
    cells.forEach(cell => {
        const [row, col] = cell.id.split('-').map(Number); // Extract row and column from ID
        if (cell.style.backgroundColor === 'black') {
            gameState.cells.push({
                row,
                col,
                alive: true
            });
        }
    });

    savedStates.push(createGridFromMinimumState(JSON.stringify(gameState)));
    displaySavedGameStates();
    createGrid();
}

function calculateMinimumSpace(gameStateJSON) {
    const gameState = JSON.parse(gameStateJSON);

    let minRow = Infinity, maxRow = -Infinity;
    let minCol = Infinity, maxCol = -Infinity;

    for (const cell of gameState.cells) {
        if (cell.alive) {
            minRow = Math.min(minRow, cell.row);
            maxRow = Math.max(maxRow, cell.row);
            minCol = Math.min(minCol, cell.col);
            maxCol = Math.max(maxCol, cell.col);
        }
    }

    const rowsNeeded = maxRow - minRow + 1;
    const colsNeeded = maxCol - minCol + 1;

    return [rowsNeeded, colsNeeded];
}

// TODO: remove all json, this can be done with the gameState object
function createGridFromMinimumState(gameStateJSON) {
    console.log(gameStateJSON)
    const [rowsNeeded, colsNeeded] = calculateMinimumSpace(gameStateJSON);
    const mappedGameState = mapCoordinatesToMinimumGrid(JSON.parse(gameStateJSON), [rowsNeeded, colsNeeded]);
    const stateGrid = document.createElement('div');

    stateGrid.style.display = 'grid';
    stateGrid.style.gridTemplateColumns = `repeat(${colsNeeded}, ${GRIDSPACE/2}px)`; // Use calculated columns
    stateGrid.style.cursor = 'pointer';

    // Adjust for offset to maintain original position if needed
    for (let i = 0; i < rowsNeeded; i++) {
        for (let j = 0; j < colsNeeded; j++) {
            const cell = document.createElement('div');
            const cellId = `${i}-${j}`;
            cell.id = cellId;
            const gameStateCell = mappedGameState.cells.find(c => c.row === i && c.col === j)

            // Set the background color based on existence and style of the cell in gameState
            cell.style.backgroundColor = gameStateCell && gameStateCell.alive == true ? 'black' : 'white';

            cell.classList.add('h-2', 'w-2','cell', 'border', 'border-gray-200');
            stateGrid.appendChild(cell);
        }
    }
    return stateGrid;
}

function mapCoordinatesToMinimumGrid(gameState, minimumGrid) {
    // 1. Extract necessary information
    const { cells } = gameState;
    const [minRows, minCols] = minimumGrid;

    // 2. Calculate offsets (shifts)
    let minRow = Infinity, minCol = Infinity;
    for (const cell of cells) {
        minRow = Math.min(minRow, cell.row);
        minCol = Math.min(minCol, cell.col);
    }

    // 3. Create mapping function
    const mapCoordinate = (oldRow, oldCol) => {
        const newRow = oldRow - minRow;
        const newCol = oldCol - minCol;

        if (newRow < 0 || newRow >= minRows || newCol < 0 || newCol >= minCols) {
            throw new Error("Invalid coordinates. Cell outside minimum grid.");
        }

        return { row: newRow, col: newCol };
    };

    // 4. Map all cell coordinates
    const mappedCells = cells.map(cell => ({
        ...cell,
        ...mapCoordinate(cell.row, cell.col)
    }));

    // 5. Return updated game state
    return {
        ...gameState,
        cells: mappedCells
    };
}

function displaySavedGameStates() {
    const statesContainer = document.getElementById('states');
    statesContainer.innerHTML = ''; // Clear previous states

    for (const grid of savedStates) {
        console.log(grid)
        const container = document.createElement('div');
        container.classList.add('state-grid')

        container.addEventListener('click', () => {
            const allStates = statesContainer.querySelectorAll('.state-grid');
            allStates.forEach(state => state.classList.remove('selected'));
            container.classList.add('selected');
        })

        container.style.display = 'grid'
        container.style.gridTemplateColumns = `repeat(var(--grid - cols), 1fr)`;
        container.style.gridTemplateRows = `repeat(var(--grid - rows), 1fr)`;
        container.style.border = '1px solid #ccc';


        container.appendChild(grid);
        statesContainer.appendChild(container);
    }
}

function loadGameState() {
    const selectedState = document.querySelector('.state-grid.selected');
    if (!selectedState) {
        alert('Please select a state to load.');
        return;
    }
    const gameBoard = document.getElementById('gameBoard');
    const selectedCells = selectedState.querySelectorAll('.cell');

    const cellData = Array.from(selectedCells).map(cell => ({
        row: parseInt(cell.id.split('-')[0]), // Assuming you store row/col in data attributes
        col: parseInt(cell.id.split('-')[1]), // Assuming you store row/col in data attributes
        alive: cell.style.backgroundColor == 'black' // Or however you indicate alive cells
    }));

    for (const cell of cellData) {
        const gameCell = gameBoard.querySelector(`#\\3${cell.row}-${cell.col}`);
        gameCell.style.backgroundColor = cell.alive ? 'black' : 'white';
    }

}

function printStateJSON() {
    const selectedState = document.querySelector('.state-grid.selected');
    if (!selectedState) {
        alert('Please select a state to load.');
        return;
    }

    const selectedCells = selectedState.querySelectorAll('.cell');
    const cellData = Array.from(selectedCells).map(cell => ({
        row: parseInt(cell.id.split('-')[0]),
        col: parseInt(cell.id.split('-')[1]),
        alive: cell.style.backgroundColor === 'black' // Assuming this indicates 'alive'
    }));

    const stateJSON = JSON.stringify({ cells: cellData , gridSize: ROWS});
    console.log(stateJSON);
}

document.getElementById('resetbtn').onclick = handleReset;
document.getElementById('save-pattern').onclick = saveGameState;
document.getElementById('load-pattern').onclick = loadGameState;
document.getElementById('step-speed').value = STEP_SPEED;

/* Set view height to screen - header */
const header = document.getElementById("header");
const gridContainer = document.getElementById("container");

function setGridHeight() {
    const headerHeight = header.offsetHeight; // Get header's height
    const viewportHeight = window.innerHeight;

    gridContainer.style.height = `${viewportHeight - headerHeight}px`;
}

// Initial height calculation
setGridHeight();

// Recalculate on window resize
window.addEventListener("resize", setGridHeight);
createGrid();