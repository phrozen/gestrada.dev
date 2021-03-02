---
date: 2021-02-18
title: Colorful Life
author: Guillermo Estrada
description: 
---
<script>
let size;             // width * height (DRY)
let state, next;      // current and nex generation
let percentage = 25;  // ~% of live cells to seed
let live = true;      // Use booleans as we don't need to store
let dead = false;     // the color, it's calculated on draw
let offset;           // hold the adjacent offsets of neighbours
let resolution = 10;  // to create a virtual grid on top the canvas
let half;             // readability

function setup() {
  createCanvas(1280, 720);
  frameRate(15);
  noStroke();
  width = width / resolution;
  height = height / resolution;
  size = width * height;
  half = resolution / 2;
  state = Array(size).fill(dead);
  next = Array(size).fill(dead);
  offset = [    // offsets for neighbours in 1D array
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
  if (mouseIsPressed === true) {
    let i = floor(mouseY / resolution) * width + floor(mouseX / resolution);
    state[i] = live;
    cell(i, mouseX, mouseY);
    return;
  }
  
  clear();
  background(0);
  for (let i = 0; i < size; i++) {
    if (state[i]) {
      let x = (i % width) * resolution;
      let y = (i / width) * resolution;
      cell(i, x, y);
    }
  }
  step();  
}

// Draws a colored circle with interpolated hue
function cell(i, x, y) {
  fill(`hsl(${floor(map(i,0,size,0,360))},100%,50%)`);
  circle(floor(x) + half, floor(y) + half, resolution);
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
    if      ((state[i] == live) && (neighbours < 2))  next[i] = dead;      // under-population
    else if ((state[i] == live) && (neighbours > 3))  next[i] = dead;      // over-population
    else if ((state[i] == dead) && (neighbours == 3)) next[i] = live;      // reproduction
    else                                              next[i] = state[i];  // stasis
  }
  let tmp = state; state = next; next = tmp;
}

// Gets cell 'status' at a given index (1D)
function at(i) {
  if (i < 0) i += size;
  if (i > size) i -= size;
  return state[i] == live ? 1 : 0;
}
</script>
