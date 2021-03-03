let size;               // width * height (DRY)
let state, next;        // current and next generation
let percentage = 25;    // ~% of live cells to seed
let live = 255;         // color for live cells ('limegreen', '#FF6600')
let dead = 0;           // color for dead cells ('pink', 'rgba(123,45,67)')
let offset;             // hold the adjacent offsets of neighbours

function setup() {
    createCanvas(180, 120);
    frameRate(10);
    size = width * height;
    state = Array(size).fill(dead);
    next = Array(size).fill(dead);
    offset = [ // offsets for neighbours in 1D array
        -width - 1, // nw
        -width,     // n
        -width + 1, // ne
        1,          // e
        width + 1,  // se
        width,      // s
        width - 1,  // sw
        -1,         // w
    ];
    seed();
}

// Main rendering loop
function draw() {
    for (let i = 0; i < size; i++) {
        set(i % width, i / width, color(state[i]));
    }
    updatePixels();
    step();
}

// Randomly seeds the state with live cells
function seed() {
    state.fill(dead);
    const living = floor(size * percentage / 100);
    for (let i = 0; i < living; i++) {
        state[floor(random(size))] = live;
    }
}

// Creates the next generation of cells
function step() {
    for (let i = 0; i < size; i++) {
        let neighbours = 0;
        for (let j of offset) {
            neighbours += at(i + j);
        }
        if ((state[i] == live)      && (neighbours < 2))  next[i] = dead;     // under-population
        else if ((state[i] == live) && (neighbours > 3))  next[i] = dead;     // over-population
        else if ((state[i] == dead) && (neighbours == 3)) next[i] = live;     // reproduction
        else                                              next[i] = state[i]; // stasis
    }
    let tmp = state;
    state = next;
    next = tmp;
}

// Gets cell 'status' at a given index (1D)
function at(i) {
    if (i < 0)      i += size;
    if (i >= size)  i -= size;
    return state[i] == live ? 1 : 0;
}

function mousePressed() {
    seed();
}