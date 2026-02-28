// #region setup
let size;               // width * height (readability)
let state, next;        // current and nex generation
let percentage = 25;    // ~% of live cells to seed
let live = true;        // Use booleans as we don't need to store
let dead = false;       // the color, it's calculated on draw
let offset;             // hold the adjacent offsets of neighbours
let resolution = 10;    // to create a virtual grid on top the canvas
let half;               // readability (resolution / 2)
let w, h;

function setup() {
    createCanvas(1280, 720);
    frameRate(15);
    noStroke();
    w = width / resolution;
    h = height / resolution;
    size = w * h;
    half = resolution / 2;
    state = Array(size).fill(dead);
    next = Array(size).fill(dead);
    offset = [ // offsets for neighbours in 1D array
        -w - 1, // nw
        -w, // n
        -w + 1, // ne
        1, // e
        w + 1, // se
        w, // s
        w - 1, // sw
        -1, // w
    ];
    seed();
}
// #endregion setup

// #region draw
// Main rendering loop
function draw() {
    if (mouseIsPressed === true && mouseX >= 0 && mouseX <= width && mouseY >= 0 && mouseY <= height) {
        let i = floor(mouseY / resolution) * w + floor(mouseX / resolution);
        if (i >= 0 && i < size) {
            state[i] = live;
            cell(i, mouseX, mouseY);
            return;
        }
    }

    clear();
    background(0);
    for (let i = 0; i < size; i++) {
        if (state[i]) {
            let x = (i % w) * resolution;
            let y = floor(i / w) * resolution;
            cell(i, x, y);
        }
    }
    step();
}

// Draws a colored circle with interpolated hue
function cell(i, x, y) {
    fill(`hsl(${floor(map(i, 0, size, 0, 360))},100%,50%)`);
    circle(floor(x) + half, floor(y) + half, resolution);
}
// #endregion draw

function seed() {
    state.fill(dead);
    const living = floor(size * percentage / 100);
    for (let i = 0; i < living; i++) {
        state[floor(random(size))] = live;
    }
}

function step() {
    for (let i = 0; i < size; i++) {
        let neighbours = 0;
        for (let j of offset) {
            neighbours += at(i + j);
        }
        if ((state[i] == live) && (neighbours < 2 || neighbours > 3)) next[i] = dead;
        else if ((state[i] == dead) && (neighbours == 3)) next[i] = live;
        else next[i] = state[i];
    }
    let tmp = state;
    state = next;
    next = tmp;
}

function at(i) {
    if (i < 0) i += size;
    if (i >= size) i -= size;
    return state[i] == live ? 1 : 0;
}
