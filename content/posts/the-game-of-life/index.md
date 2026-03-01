---
title: The Game of Life
description: Introduction to p5.js and cellular automata with Conway's Game of Life.
tags:
  - Generative Art
  - Algorithms
  - JavaScript
  - p5.js
  - Cellular Automata
  - Mathematics
  - Simulation
lastUpdated: 2026-02-27
---

# The Game of Life

<!-- 
Hero Image AI Prompt:
"A retro, pixelized cellular automaton grid coming to life. Glowing neon green and pink cells on a dark background, representing Conway's Game of Life. 8-bit aesthetic, digital art, generative computer science vibe."
-->
![Conway's Game of Life Hero Image](./cover.webp)

## Hello World!

Finally! I have been trying to start a dev blog for like 10 years now, and I always found a way to procrastinate that :weary:.

I always thought I would never have the time or enough content to publish. But I've always been an enthusiast of generative art and algorithmic art, and one day it hit me: "Hey, that never goes old!". So here we are.

Usually the first thing you do when you start something technology-related is a _Hello World_ of some sort. But this time around, I would like to do something related to generative art, as this blog is going to track all the progress and projects I do regarding that adventure. For that purpose, what better place to start than with _Conway's The Game of Life_! It is sort of the _Hello World_ of generative art after all. (isn't it? :sweat_smile:)

## What is it?

For those of us who thought _The Game of Life_ was an [old family board game created by Milton Bradley](https://en.wikipedia.org/wiki/The_Game_of_Life), we need a little bit more of an introduction to Conway's version. Directly from its [Wikipedia article](https://en.wikipedia.org/wiki/Conway%27s_Game_of_Life):

> The Game of Life is a cellular automaton devised by the British mathematician John Horton Conway in 1970. It is a zero-player game, meaning its evolution is determined entirely by its initial state. You interact with it by creating an initial configuration and observing how it evolves. It is Turing complete and can simulate a universal constructor or any other Turing machine.

That last part is really interesting — **Turing complete**. That means this tiny set of rules, applied to a grid of cells, can in theory compute *anything* a regular computer can. From nothing but a grid and four rules. And it produces incredible looking results that you can tweak endlessly through code :sparkles:

The universe of the Game of Life is an infinite, two-dimensional grid of square cells. Each cell can either be **alive** or **dead**. Every cell interacts with its eight neighbors (horizontally, vertically, and diagonally adjacent). At each step in time, we evaluate the following rules:

1. **Under-population**: A live cell with fewer than two live neighbors dies.
2. **Survival**: A live cell with two or three live neighbors lives on.
3. **Over-population**: A live cell with more than three live neighbors dies.
4. **Reproduction**: A dead cell with exactly three live neighbors becomes alive.

Seems simple enough! But the magic lies in how these simple rules create massive, infinite complexity. We determine the initial state either manually or randomly, and then we just... watch it evolve. No player input. No decisions. Just pure emergent behavior from four lines of logic.

So... how do we actually code this? Let's do this!

## The World is Flat

Once we start talking about 2D grids, our minds immediately jump to 2D arrays (e.g., `state[x][y]`). But here is the gotcha: **there is no such thing as a 2D array for a computer**. They are just mathematical constructs on top of arbitrary length data! Usually, the most resource-intensive part of graphics sketches is rendering, and traversing nested arrays pixel-by-pixel is slow.

Instead, let's use **1D arrays**. This is how frame buffers and images are actually represented natively in computer graphics — one long flat buffer of pixels, row after row. Converting between 1D and 2D coordinates is straightforward:

```js
// From 2D(x, y) to 1D(i)
i = y * width + x

// From 1D(i) to 2D(x, y)
x = i % width
y = floor(i / width)
```

We store each row of pixels one after the other, so we need the `width` to know where each row ends. One thing to note: that `y` could be a floating point number when dividing, and we can only index arrays with integers, so don't forget `floor()`! (I learned that one the hard way :sweat_smile:)

And here is the coolest part: **1D arrays are already toroidally bound on the X-axis!** :astonished: What does that mean? It means if `x == width` and we do `x + 1`, instead of throwing an `Out of Bounds` error like a 2D array would, it seamlessly wraps to the first item of the next row. Just like Pac-Man going out one side of the screen and coming out the other! Not perfect (you end up a row below), but good enough for our grid — and way simpler than explicit boundary checking.

## The Setup

We are going to use [p5.js](https://p5js.org/) for this. From their site:

> p5.js is a JavaScript library for creative coding, with a focus on making coding accessible and inclusive for artists, designers, educators, beginners, and anyone else!

Every sketch starts with two functions: `setup` and `draw`. Beyond that, everything else is just plain old JavaScript. Let's start by declaring our state arrays and the neighbor offsets.

Instead of using booleans right away, let's store the actual color values of our cells — `255` (white) for live, `0` (black) for dead. This has a sneaky benefit: `0` evaluates to `false` in JavaScript, so our conditionals still work, but we can also use the value directly for rendering.

<<< @/posts/the-game-of-life/simple-life.js#setup

Let's talk about that `offset` array — it's one of my favorite tricks in this whole implementation. Accessing neighbors in a 2D array means checking 8 different `[x+dx][y+dy]` combinations. But in a 1D array, each neighbor is just a simple arithmetic offset from the current index:

```js
state[i - 1]            // west
state[i + 1]            // east
state[i - width]        // north (one full row back)
state[i + width]        // south (one full row forward)
state[i - width - 1]    // northwest
state[i - width + 1]    // northeast
state[i + width - 1]    // southwest
state[i + width + 1]    // southeast
```

Since only `i` changes and the offsets are constants (once `width` is known), we precompute them into that array and just iterate over it. No nested loops, no boundary math, just `i + offset[j]`.

## Initialization

In a 2D array, randomly seeding live cells means looping over `X` and `Y` and rolling a dice for each cell (like flipping a coin). This usually nets you exactly ~50% alive/dead population.

But with 1D arrays? We can just calculate the literal *number* of cells we want alive based on a percentage, pick `N` random index spots, and turn them on! Faster, simpler, and way more flexible.

*(A quick gotcha: `p5.js`'s `random()` function returns floating point numbers! So remember to `floor` it before you try indexing your array, or JavaScript will be very unhappy).*

<<< @/posts/the-game-of-life/simple-life.js#seed

## The Rules of Life

Every step we must iterate over each cell, count its live neighbors, and apply Conway's rules. Using our `offset` array, this becomes beautifully simple.

But first, we need a little helper. Remember how the 1D array only natively wraps on the X-axis? For the Y-axis, we need to handle it manually. Enter the `at(i)` function:

<<< @/posts/the-game-of-life/simple-life.js#helper

What's happening here? If the index goes negative (above the top row), we add `size` to wrap it to the bottom. If it exceeds `size` (below the bottom row), we subtract `size` to wrap it to the top. Then we return `1` for alive, `0` for dead — perfect for summing up neighbors.

Then during each step, we apply the rules to generate the next state. The formatting looks a little weird, but I did it on purpose for readability — each of Conway's four rules maps to exactly one line:

<<< @/posts/the-game-of-life/simple-life.js#step

Notice the swap at the end: instead of creating a new array every frame (expensive!), we maintain two arrays and just swap references. The "next" generation becomes the "current," and the old "current" gets recycled as scratch space.

## Drawing

Because our logic is blazing fast, the rendering is a single-loop assignment. We convert each 1D index back to 2D coordinates inline and set the pixel color directly from the state value:

<<< @/posts/the-game-of-life/simple-life.js#draw

The `set(x, y, color)` function writes directly to the pixel buffer, and `updatePixels()` flushes the whole frame at once. Because we stored the color value (255/0) directly in the state array, we don't need any conditionals — just `color(state[i])`.

Now check it out!

<Sketch 
  title="Simple Life" 
  source="content/posts/the-game-of-life/simple-life.js"
  description="Click anywhere on the canvas to re-seed it!"
>
  <P5Embed src="./simple-life.js" pixelated />
</Sketch>

The initial state is randomly seeded with 25% live cells, and then we let the game follow its course. I used `frameRate(10)` so that each generation could be appreciated.

If you try this code in the [p5.js web editor](https://editor.p5js.org/), you'll notice the sketch is really small — exactly 180×120 pixels as defined in `setup`. Your first instinct would be to bump up `createCanvas(1280, 720)`, but that doesn't scale the pixels — it just gives you way more cells (and way more computation). The trick is that the browser can scale the `<canvas>` element independently of its resolution! I'm using CSS to stretch it to full width:

```css
canvas {
    width: 100% !important;
    height: 100% !important;
    image-rendering: pixelated;
}
```

That `image-rendering: pixelated` is the secret sauce — it disables interpolation so our tiny pixels don't blur into each other when scaled up. They stay as perfect, crisp squares. (Some browsers call this `crisp-edges`, but `pixelated` works in all modern ones now :tada:)

## Making it Colorful (and Interactive!)

Okay, there is not much "art" in diminutive black and white pixels dancing around a tiny canvas (or... is there? :thinking:). Thankfully, because we implemented this grid so cleanly, manipulating the visuals is a breeze.

First, some structural changes. We grow the canvas to a comfortable 1280×720, and introduce a `resolution` variable that creates a "virtual grid" on top of the canvas. With `resolution = 10`, our actual cell grid becomes 128×72 — a nice size to play with. And since we're drawing shapes now instead of raw pixels, we switch `live`/`dead` from integers to booleans. (The rest of the code works without a hitch because `false` is falsy in JavaScript — same behavior, cleaner semantics.)

The real magic happens in the rendering. Instead of setting individual pixels, we only draw the live cells as colored circles. We skip dead cells completely (way more efficient than clearing and drawing everything). And for the color? We map the cell's 1D index to the HSL hue spectrum — `map(i, 0, size, 0, 360)`. That's it. Instant rainbow :rainbow:

<<< @/posts/the-game-of-life/colorful-life.js#draw

The `cell(i, x, y)` function interpolates `i` into a hue value `[0°, 360°)`, creating a smooth gradient across the entire grid. And notice the `mouseIsPressed` check at the top of `draw` — when you click and drag, it converts mouse coordinates back to grid coordinates and seeds live cells directly into the current generation. You can literally paint life onto the canvas while it evolves.

The result? A clear winner :trophy:

<Sketch 
  title="Colorful Life" 
  source="content/posts/the-game-of-life/colorful-life.js"
  description="Click and drag your mouse across the canvas to magically paint new life!"
>
  <P5Embed src="./colorful-life.js" />
</Sketch>

Now we're talking! Try clicking and dragging your mouse across the paused cells above to "paint" new life structures dynamically and watch them explode.

## Closing Thoughts

Experimenting with simple rules that result in infinite complexity is incredibly satisfying. Four rules. That's it. No physics engine, no AI, no complex state machines. Just four conditionals applied to a grid, and the system produces gliders, oscillators, spaceships, and patterns that mathematicians are *still* discovering 55 years later.

I really enjoyed writing _Life_ and I will definitely keep doing these visual algorithms! There are hundreds of implementations all over the internet — even one in the [Go documentation](https://golang.org/doc/play/life.go) and another in the [p5.js examples](https://p5js.org/examples/simulate-game-of-life.html). But there's something deeply satisfying about writing your own from scratch, making design decisions like 1D arrays over 2D, and discovering tricks like the offset array along the way.

In a future post, I might push this into WebAssembly with Go to see how many millions of cells we can crank at 60FPS. Or maybe explore other cellular automata like [Langton's Ant](https://en.wikipedia.org/wiki/Langton%27s_ant) or [Rule 110](https://en.wikipedia.org/wiki/Rule_110) — another Turing complete system, but this time in just one dimension. (How wild is that? :astonished:)

¡Hasta la próxima!

