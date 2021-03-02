---
date: 2021-02-18
title: The Game of Life
description: Hello blog! Introduction to p5.js and cellular automata with Conway's Game of Life
author: Guillermo Estrada
slug: the-game-of-life
images:
- "https://images.unsplash.com/photo-1530328411047-7063dbd29029?crop=entropy&cs=srgb&fm=jpg&ixid=MXwyMTAzMTZ8MHwxfGFsbHx8fHx8fHx8&ixlib=rb-1.2.1&q=80&fit=crop&w=1200h=627"
tags:
    - p5.js
    - generative art
    - automaton
categories:
    - p5.js
series:
    - Game of Life
---

{{<cover KJMz5Tmbw0k>}}

Finally! I have been trying to start a dev blog for like 10 years now, and I always found a way to procrastinate that 😩.
I always thought I would never have the time or enough content to publish, and with how fast technology becomes outdated, it seemed inappropriate at the time. On the other hand, I always have been an enthusiast of generative artwork or algorithmic art, and I thought: "Hey, that never goes old!", so this time around at least I got to the blog creation part, the first post and a path to learn a lot.

Usually the first thing you do when you start something technology related, has to be a _Hello World_ of some sort. I enjoy writing Go or Javascript code, so I would love to just do something like...

```go
import "fmt"

func main() {
    fmt.Println("Hello Blog!")
}
```
... and be done with it!

But this time, I would like to do something related to generative art, as this blog is going to track all the progress and projects I do regarding that adventure (among any other dev related stuff here and there). For that purpose, an appropriate first project was necessary, and after very little deliberation, what a better place to start than with _Conway's The Game of Life_! It is sort of the _Hello World_ of generative art after all. (isn't it? 😅)

#### Introduction

For those of us who thought _The Game of Life_ or just _Life_ for short, was an [old family board game created by Milton Bradley](https://en.wikipedia.org/wiki/The_Game_of_Life), we need a little bit more of an introduction to _Conway's Game of Life_, so directly from it's [Wikipedia article](https://en.wikipedia.org/wiki/Conway%27s_Game_of_Life):

> The Game of Life, also known simply as Life, is a cellular automaton devised by the British mathematician John Horton Conway in 1970. It is a zero-player game, meaning that its evolution is determined by its initial state, requiring no further input. One interacts with the Game of Life by creating an initial configuration and observing how it evolves. It is Turing complete and can simulate a universal constructor or any other Turing machine.

This is really interesting, and it seems easy to implement and a lot of fun to play with once you have a little more information about the algorithm. It also produces very good looking results and you can tweak the implementation in a lot of ways to produce more artistic animations, all through code! 😁 So, where to start? Let's keep reading:

> The universe of the Game of Life is an infinite, two-dimensional orthogonal grid of square cells, each of which is in one of two possible states, live or dead, (or populated and unpopulated, respectively). Every cell interacts with its eight neighbours, which are the cells that are horizontally, vertically, or diagonally adjacent.

So, once we start talking about two-dimension orthogonal grids, it immediately comes to mind a way to implement it using arrays, and although we cannot make an infinite grid using arrays, we can simulate the _infinity_ by wrapping the borders toroidally (just like Pac-Man ⍩⃝ would go out on one side of the screen and appear on the opposite side). Or we can _clamp_ our grid to a finite size, just as if we are looking at a sample of this infinite world. Each of these discrete cells can only be in 2 states (alive | dead), so making a `boolean` array makes the most sense, but the question is: How do we play? How do we decide who lives and who dies?

In _Conway's Game of Life_, we determine the initial state of the grid, either by manually setting some cells to be alive arbitrarily or we can do so randomly. Doing it randomly seems to be the better approach, that way we can _"play"_ with the initial state and we will get a different _"game"_ every time. It is worth noting that there is no such thing as *random* in computers, we only have *pseudo-random number generators*, and that could potentially be good, because if we seed our generator, we can reproduce the exact same game by just using the same seed as an input. Once the initial state is set, we have to iterate through our grid and decide who lives and who dies into the next generation. We do so by following the rules set by Conway:

> At each step in time, the following transitions occur:
> 
> 1. Any live cell with fewer than two live neighbours dies, as if by under-population.
> 2. Any live cell with two or three live neighbours lives on to the next generation.
> 3. Any live cell with more than three live neighbours dies, as if by over-population.
> 4. Any dead cell with exactly three live neighbours becomes a live cell, as if by reproduction.

That seems simple enough, to implement this we have to maintain two states, the current generation and the next generation, and we can reuse both states by swapping them in each time step. To calculate the next generation state we have to iterate over the current one, and for each cell, check all it's neighbours and get a count of how many are alive, then proceed to seed the next generation based on these rules. After that, the next generation becomes the current generation and we start all over again. Let's do this!

There are hundreds of implementations of _Conway's Game of Life_ all over the internet, there is even one for the console in the [Go documentation](https://golang.org/doc/play/life.go) and of course there is another one in the [p5.js examples](https://p5js.org/examples/simulate-game-of-life.html), so let's try something different. This time let's work with 1D arrays! One dimensional arrays (or just arrays as we know them) are usually used in computer graphic to represent images, frame buffers and all kind of two dimensional matrix. The reason is, there is no concept of 2D arrays for computers, those are just mathematical constructs on top of arbitrary data on any length. Usually the most resource intensive part of any graphics sketch is the rendering, and 2D arrays are not very optimal because we have to iterate over them and set each pixel 1 by 1 (or 1 row at a time), but with 1D arrays we could set the whole pixel data in a single assignment if we plan for it!


#### Writing some code

Let's not get ahead of ourselves and start coding something. We're going to be using [p5.js](https://p5js.org/) for this very simple example of life. From their site:

> p5.js is a JavaScript library for creative coding, with a focus on making coding accessible and inclusive for artists, designers, educators, beginners, and anyone else! p5.js is free and open-source because we believe software, and the tools to learn it, should be accessible to everyone.

Every sketch starts the same, with 2 required functions `setup` and `draw`, aside from that, everything else is just plain old javascript and everything is fair game, let's start by doing our `setup` and have some arrays to store the state of our cells.

```js
let size;             // simplify code
let state, next;      // current and next generation
let live = 255;       // color for live cells
let dead =   0;       // color for dead cells

function setup() {
  createCanvas(180, 120);
  size = width * height;
  state = Array(size).fill(dead);
  next = Array(size).fill(dead);
}
```

Instead of using boolean arrays already, let's start by using integers and define the actual value of the color of dead or live cells, in this case `255` or white for live ones, and `0` or black for dead ones. This also has the nice side effect that 0 evaluates to false in javascript, but declaring them in a variable will make the code a lot more readable and will let us change colors later on if we want to. We also declared `size` as a shortcut for `width * height`, it will save a lot of calculations in the long run, but it's mainly for cleaner code too. We create our state arrays of length `size` and fill them with `dead` cells. We will be going back and forth between 1D and 2D coordinates, so let's review the conversions:

```go
// From 2D(x, y) to 1D(i)
i = y * width + x

// From 1D(i) to 2D(x, y)
x = i % width
y = i / width
```

In a 1D array, we store each row(y) of pixels one after the other, so we need to always use the width to know when each of those rows end and the next one begins. Conversions are very straight forward, but it's worth noting that depending on the language and platform, one has to be careful that `y` might be a floating point number, and we can only index arrays with integer numbers so let's be careful move on. One thing to note about 1D arrays, is that they are already toroidally bound (of sorts 😅) on the x axis! That means that for `x == width` if we do `x + 1`, what usually would be an `out of bounds` error on 2D arrays, in a 1D array, we just get the first item of the next row (on the other side of the screen). Not perfect but it will be good enough for this use case, we will bound the y axis later on too.

The next step would be seeding our initial state with random live cells. In a 2D array we could simply iterate over it and randomly set the cell state, but this will always give us approximately 50% of live/dead cells each time (just like flipping a coin). With 1D arrays on the other hand, we can define a percentage of live cells to seed our initial state, then generate that percentage of random numbers between `0` and `size` and use them as index to set those cells to a live state. Definitely faster, simpler and more flexible. Declare our `percentage` at the top and create our seed function.

```js
// Randomly seeds the state with live cells
function seed() {
  state.fill(dead);
  const living = floor(size * percentage / 100);
  for (let i = 0; i <living; i++) {
    state[floor(random(size))] = live;
  }
}
```

We make sure the state is filled with dead cells before seeding just in case we want to re-seed mid game (restart). As we can see, seeding a `%` of the state with live cells is way easier in 1D arrays. Quick note, the `random()` function in p5.js has many modes. When called with a single parameter, it will produce a number between `0` (inclusive) and that number (exclusive), but this number is a `float`, so we need to round it down to use as an index.

Next, just add a `seed()` call at the end of our setup, and we are ready to draw our cells into the canvas to see that we seeded correctly. Let's write our draw function next, in will also be very simple thanks to our decisions so far.

```js
// Main rendering loop
function draw() {
  for (let i = 0; i < size; i++) {
    set(i % width, i / width, color(state[i]));
  }
  updatePixels();
  // step(); // Here we will create the next generation
}
```

That's it! Single loop and single call to set the color of the cell based on our current state. The `set(x, y, color)` function is very easy to use. We just transform our 1D coordinates into 2D inline and we don't even have to fix `y` because the the canvas does not care, it can take floating numbers as coordinates and will just interpolate the best possible way. One thing to note here is that we must call `updatePixels()` once we are done setting our pixel data so that it renders the whole screen again. Finally we are using the `color()` function, which also has a lot of modes, so make sure to [check it out](https://p5js.org/reference/#/p5/color) so we can just change our colors later on.

Last but not least, our `step()` function, here is where the magic happens, we iterate through our current state, apply Conway's rules and create a new generation of cells. We will do this over and over, and this will be _*The Game of Life!*_ Before we proceed, let's review how do we access neighbours in a 1D array, as it might not be that intuitive after all. We need to check the status of all 8 adjacent cells to find out what to do with our current cell.

```js
// These are really simple and straight forward
state[i - 1]            // west
state[i + 1]            // east
// If we go one width back, we are actually moving 'up'
// a row, same goes for down, we add a width...
state[i - width]        // north 
state[i + width]        // south
// The rest is just a simple combination of all of them
state[i - width - 1]    // northwest
state[i - width + 1]    // northeast
state[i + width - 1]    // southwest
state[i + width + 1]    // southeast
```

From the code above we can see that for each `offset` only `i` is a variable, and we can treat the rest as a constant, let's optimize here and calculate the constant part and put them in an array so we can iterate over those faster later on. Declare `offsets` at the top and initialize it inside `setup` once the `width` is known.

```js
offset = [      // offsets for neighbours in 1D array
    -width - 1, // nw
    -width,     // n
    -width + 1, // ne
    1,          // e
    width + 1,  // se
    width,      // s
    width - 1,  // sw
    -1,         // w
];
```

One last thing before we proceed with `step()`, let's do a helper function `at(i)` that will tell us if the cell is *dead or alive* for any given position in the current state, and while we are at it, let's bound the `y` axis toroidally altogether.

```js
// Gets cell 'status' at a given index (1D)
function at(i) {
  if (i <0) i += size;
  if (i> size) i -= size;
  return state[i] == live ? 1 : 0;
}
```

What is happening here? As we only care about two bounds, if `index < 0` we add `size` to it essentially moving the index to the last row `y` in the same `x` position. Comparable to that, if `index > size` we move it to the first row `y` by subtracting `size` from it. We then return `0` if it's dead or `1` if it's live, as we do not care about the actual color value of the state, but just the status for calculating the number of neighbours. Now, for the last part...

```js
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
```

The code is weirdly formatted on purpose for readability. As said before, we iterate over each cell and for each one, we count the number of live neighbours it has (that is why `at(i)` returns 0 or 1) by iterating over our `offset` array defined above. After that we follow Conway's rules to determine what the status of that cell will be towards the next generation. This could be simplified a bit further between the first and second checks (over and under population), but I think it's a bit more explicit what is happening this way. Finally we swap our current generation (`state`) with the next generation. Let's give it a try, shall we?

#### Simple Life
{{<sketch simple-life>}}

Great! 😁 The initial state is randomly seeded with 25% live cells, and then we let the game follow it's course. But there is no fun in just watching it a single time, so you can click on {{<fa refresh accent>}} or anywhere on the canvas to re-seed the state and restart the game. This usually runs much faster, specially for the canvas size we picked as it's very small, but I used `frameRate(10)` in our setup so that each step in time could be appreciated. You can click on {{<fa code accent>}} on the sketch toolbar to check the code in Github. If you try this code in the [p5.js web editor](https://editor.p5js.org/), you will notice that the sketch is really small, exactly the way we defined it in the `setup` with `createCanvas(180, 120)`, as we are actually using pixels! Your first impulse would be to change the size and make the canvas bigger, but that will not scale the pixels, you are just going to have a lot more cells to deal with. That is why most people implement _The Game of Life_ by drawing squares or so, but the trick here, is that the browser already has the ability to scale the `canvas` element regardless of it's size! So while embedding sketches in my blog I'm using...

```css
main, .p5Canvas {
    width: 100% !important;
    height: auto !important;
    max-width: 860px;
    display: block;
    image-rendering: pixelated;
}
```

... essentially overriding p5.js style and making the canvas scale to the content, it also makes it responsive if we also scale the `iframe` (try resizing the browser). Neat! One attribute is apparently not supported in {{<fa firefox accent>}} Firefox (at the time of writing) and that is `image-rendering: pixelated`. What it does, is essentially disabling interpolation while scaling our canvas, so our little pixels don't blur out and look like perfect squares when scaled up! You can read all about it in the [MDN Web Docs](https://developer.mozilla.org/es/docs/Web/CSS/image-rendering).

But, what if we wanted to implement it in a large canvas with colors and figures like everyone else? This is about generative art after all, right? There is not much art in diminutive black and white pixels dancing around a canvas (or... is there? 🤔). Thankfully the way we implemented this version makes it incredibly easy to play and create different versions of the game. Let's start by doing some changes in our setup and variables. I will only show the meaningful changes, as always the full code is available on Github.

```js
let live = true;      // Use booleans as we don't need to store
let dead = false;     // the color, it's calculated on draw
let resolution = 10;  // To create a virtual grid on top the canvas

function setup() {
  createCanvas(1280, 720);
  noStroke(); // disables the stroke when figure drawing
  width = width / resolution;
  height = height / resolution;
  size = width * height;
  ...
}
```

So first we grow the canvas to a comfortable viewing size, and we define a `resolution` to "scale down" our grid, in this example our `state` of cells will be of `128 * 72`, which is a good enough grid to play with. Then as we will not be using the `state` to store the color anymore, we just change dead and live to `booleans` as originally intended (yes, the rest of the code will work without a hitch). `noStroke()` disables the stroke when figure drawing, as we will be using those now. After that, we redefine our `width` and `height` values to those scaled by our resolution, this will not change tha canvas already created, but now our whole sketch will use these. Finally let's modify our `draw` function and have a little fun with it.

```js
// Main rendering loop
function draw() {
  clear();
  background(0);
  for (let i = 0; i < size; i++) {
    if (state[i] == live) {
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
```

That is all, let's quickly review what is happening, first we `clear()` the canvas and fill it with black `background(0)` as we are only going to be drawing live cells (way more efficient). Next we proceed as usual, but check for live cells only `state[i] == live` (for readability as `state[i]` is already a boolean). Now the magic happens, we need to scale both `x` and `y` back to screen coordinates by multiplying them by the resolution, and in `cell(i, x, y)` we set the `fill(color)` with a neat trick by interpolating `i` into the hue [0, 360) of an `HSL` color! And finally we draw our cell as a colored `circle(x, y, radius)` with an offset to center it. Let's see it in action! 😎

#### Colorful Life
{{<sketch colorful-life>}}

Bravo! I separated `cell` into a new function to add a bit more fun to *The Game of Life*, now instead of re-seeding randomly by clicking on the canvas, it actually seeds new cells manually! Try it, just press the mouse button and draw some live cells directly to the current generation while it remains `paused`. Now it's way more fun to play with it. To achieve this, we simply modify `draw()` like this:

```js
// Main rendering loop
function draw() {
  if (mouseIsPressed === true) {
    let i = floor(mouseY / resolution) * width + floor(mouseX / resolution);
    state[i] = live;
    cell(i, mouseX, mouseY);
    return;
  }
  // ...rest of draw function  
}
```

#### Closing thoughts

And that's all for today. You can keep learning [p5.js](https://p5js.org/) using it's [web editor](https://editor.p5js.org/) for your own sketches! Play with *The Game of Life* on the sketch above or [learn more about it](https://en.wikipedia.org/wiki/Conway%27s_Game_of_Life)! I really enjoyed writing _Life_ as it's something I have never done before, and I will definitely keep on doing it! The main purpose of this blog is to document everything I learn about generative art, as well as to share everything I already know. In a later episode, let's do a similar implementation using Go (or Rust if I can manage to learn it before then), compile it into Web Assembly and make some benchmarks!

One thing that took some time was hacking p5.js so that I could embed sketches where the canvas scale with the content, so that they could be viewed within the post naturally without worrying about specific sizing of the canvas. This led me to write some [Hugo shortcodes](https://gohugo.io/templates/shortcode-templates/) in order to generate minimal HTML files I could embed within my site, where I had full control over the style (you can check them out on Github). I would like to keep using p5.js in my posts on a regular basis, as it's awesome for drafting some code and producing amazing results, so it is most likely I will write my own minimal editor that I can embed into the blog with JS. I will be using [Svelte](https://svelte.dev/) and going the same route as the official editor by hacking away with the [iframe's srcdoc attribute](https://developer.mozilla.org/en-US/docs/Web/HTML/Element/iframe), that way I can provide fine controls (play, stop, refresh, edit, fullscreen, etc...) and I'll make sure it's open source. If it sounds like something you might be interested in let me know in the comments!

¡Hasta la próxima!

{{<github>}}