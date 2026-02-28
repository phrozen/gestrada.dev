import { _ as _export_sfc, C as resolveComponent, o as openBlock, c as createElementBlock, ah as createStaticVNode, E as createVNode, w as withCtx, j as createBaseVNode, a as createTextVNode } from "./chunks/framework.Nt3SwOR4.js";
const _imports_0 = "/assets/hero.Bnt-jeEC.webp";
const __pageData = JSON.parse(`{"title":"The Game of Life","description":"Introduction to p5.js and cellular automata with Conway's Game of Life.","frontmatter":{"title":"The Game of Life","description":"Introduction to p5.js and cellular automata with Conway's Game of Life.","tags":["Generative Art","Algorithms","JavaScript","p5.js","Cellular Automata","Mathematics","Simulation"],"lastUpdated":"2026-02-27T00:00:00.000Z"},"headers":[],"relativePath":"posts/the-game-of-life/index.md","filePath":"posts/the-game-of-life/index.md"}`);
const _sfc_main = { name: "posts/the-game-of-life/index.md" };
function _sfc_render(_ctx, _cache, $props, $setup, $data, $options) {
  const _component_P5Embed = resolveComponent("P5Embed");
  const _component_Sketch = resolveComponent("Sketch");
  return openBlock(), createElementBlock("div", null, [
    _cache[0] || (_cache[0] = createStaticVNode("", 30)),
    createVNode(_component_Sketch, {
      title: "Simple Life",
      source: "docs/posts/the-game-of-life/simple-life.js",
      description: "Click anywhere on the canvas to re-seed it!"
    }, {
      default: withCtx(() => [
        createVNode(_component_P5Embed, {
          src: "./simple-life.js",
          pixelated: ""
        })
      ]),
      _: 1
    }),
    _cache[1] || (_cache[1] = createStaticVNode("", 5)),
    createVNode(_component_Sketch, {
      title: "Colorful Life",
      source: "docs/posts/the-game-of-life/colorful-life.js",
      description: "Click and drag your mouse across the canvas to magically paint new life!"
    }, {
      default: withCtx(() => [
        createVNode(_component_P5Embed, { src: "./colorful-life.js" })
      ]),
      _: 1
    }),
    _cache[2] || (_cache[2] = createBaseVNode("p", null, `Now we're talking! Try clicking and dragging your mouse across the paused cells above to "paint" new life structures dynamically and watch them explode.`, -1)),
    _cache[3] || (_cache[3] = createBaseVNode("p", null, [
      createTextVNode("Experimenting with simple rules that result in infinite complexity is incredibly satisfying. I really enjoyed writing "),
      createBaseVNode("em", null, "Life"),
      createTextVNode(" and I will definitively keep doing these visual algorithms! In a future post, I might even try pushing this algorithm into WebAssembly with Go to see how many millions of cells we can crank at 60FPS.")
    ], -1)),
    _cache[4] || (_cache[4] = createBaseVNode("p", null, "¡Hasta la próxima!", -1))
  ]);
}
const index = /* @__PURE__ */ _export_sfc(_sfc_main, [["render", _sfc_render]]);
export {
  __pageData,
  index as default
};
