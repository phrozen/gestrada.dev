import { useSSRContext, unref } from "vue";
import { ssrRenderAttrs, ssrRenderList, ssrRenderAttr, ssrInterpolate } from "vue/server-renderer";
const data = JSON.parse(`[{"title":"The Game of Life","description":"Introduction to p5.js and cellular automata with Conway's Game of Life.","tags":["Generative Art","Algorithms","JavaScript","p5.js","Cellular Automata","Mathematics","Simulation"],"lastUpdated":"2026-02-27T00:00:00.000Z","image":"","url":"/posts/the-game-of-life/"}]`);
const __pageData = JSON.parse('{"title":"","description":"","frontmatter":{"layout":"home"},"headers":[],"relativePath":"index.md","filePath":"index.md"}');
const __default__ = { name: "index.md" };
const _sfc_main = /* @__PURE__ */ Object.assign(__default__, {
  __ssrInlineRender: true,
  setup(__props) {
    return (_ctx, _push, _parent, _attrs) => {
      _push(`<div${ssrRenderAttrs(_attrs)}><div class="post-grid"><!--[-->`);
      ssrRenderList(unref(data), (post) => {
        _push(`<a${ssrRenderAttr("href", post.url)} class="post-card"><div class="post-image-container"><img${ssrRenderAttr("src", post.url.replace(/\.html$/, "") + "/hero.webp")}${ssrRenderAttr("alt", post.title)} class="post-image"></div><div class="post-content"><h3 class="post-title">${ssrInterpolate(post.title)}</h3>`);
        if (post.description) {
          _push(`<p class="post-desc">${ssrInterpolate(post.description)}</p>`);
        } else {
          _push(`<!---->`);
        }
        if (post.tags && post.tags.length) {
          _push(`<div class="post-tags"><!--[-->`);
          ssrRenderList(post.tags, (tag) => {
            _push(`<span class="tag">${ssrInterpolate(tag)}</span>`);
          });
          _push(`<!--]--></div>`);
        } else {
          _push(`<!---->`);
        }
        _push(`</div></a>`);
      });
      _push(`<!--]--></div></div>`);
    };
  }
});
const _sfc_setup = _sfc_main.setup;
_sfc_main.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("index.md");
  return _sfc_setup ? _sfc_setup(props, ctx) : void 0;
};
export {
  __pageData,
  _sfc_main as default
};
