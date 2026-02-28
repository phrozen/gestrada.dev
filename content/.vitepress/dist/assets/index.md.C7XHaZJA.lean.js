import { o as openBlock, c as createElementBlock, j as createBaseVNode, F as Fragment, B as renderList, t as toDisplayString, e as createCommentVNode, k as unref } from "./chunks/framework.Nt3SwOR4.js";
const data = JSON.parse(`[{"title":"The Game of Life","description":"Introduction to p5.js and cellular automata with Conway's Game of Life.","tags":["Generative Art","Algorithms","JavaScript","p5.js","Cellular Automata","Mathematics","Simulation"],"lastUpdated":"2026-02-27T00:00:00.000Z","image":"","url":"/posts/the-game-of-life/"}]`);
const _hoisted_1 = { class: "post-grid" };
const _hoisted_2 = ["href"];
const _hoisted_3 = { class: "post-image-container" };
const _hoisted_4 = ["src", "alt"];
const _hoisted_5 = { class: "post-content" };
const _hoisted_6 = { class: "post-title" };
const _hoisted_7 = {
  key: 0,
  class: "post-desc"
};
const _hoisted_8 = {
  key: 1,
  class: "post-tags"
};
const __pageData = JSON.parse('{"title":"","description":"","frontmatter":{"layout":"home"},"headers":[],"relativePath":"index.md","filePath":"index.md"}');
const __default__ = { name: "index.md" };
const _sfc_main = /* @__PURE__ */ Object.assign(__default__, {
  setup(__props) {
    return (_ctx, _cache) => {
      return openBlock(), createElementBlock("div", null, [
        createBaseVNode("div", _hoisted_1, [
          (openBlock(true), createElementBlock(Fragment, null, renderList(unref(data), (post) => {
            return openBlock(), createElementBlock("a", {
              key: post.url,
              href: post.url,
              class: "post-card"
            }, [
              createBaseVNode("div", _hoisted_3, [
                createBaseVNode("img", {
                  src: post.url.replace(/\.html$/, "") + "/hero.webp",
                  alt: post.title,
                  class: "post-image"
                }, null, 8, _hoisted_4)
              ]),
              createBaseVNode("div", _hoisted_5, [
                createBaseVNode("h3", _hoisted_6, toDisplayString(post.title), 1),
                post.description ? (openBlock(), createElementBlock("p", _hoisted_7, toDisplayString(post.description), 1)) : createCommentVNode("", true),
                post.tags && post.tags.length ? (openBlock(), createElementBlock("div", _hoisted_8, [
                  (openBlock(true), createElementBlock(Fragment, null, renderList(post.tags, (tag) => {
                    return openBlock(), createElementBlock("span", {
                      key: tag,
                      class: "tag"
                    }, toDisplayString(tag), 1);
                  }), 128))
                ])) : createCommentVNode("", true)
              ])
            ], 8, _hoisted_2);
          }), 128))
        ])
      ]);
    };
  }
});
export {
  __pageData,
  _sfc_main as default
};
