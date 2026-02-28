<script setup>
import { onMounted, onUnmounted, ref } from 'vue'

const props = defineProps({
  src: { type: String, required: true },
  pixelated: { type: Boolean, default: false }
})
const sketchContainer = ref(null)
let p5Instance = null

onMounted(async () => {
  const p5 = (await import('p5')).default

  const response = await fetch(props.src)
  const code = await response.text()

  const sketch = (p) => {
    // 1. Intercept createCanvas to attach it to our Vue container automatically
    const originalCreateCanvas = p.createCanvas;
    p.createCanvas = (...args) => {
      const c = originalCreateCanvas.apply(p, args);
      c.parent(sketchContainer.value);
      return c;
    };

    // 2. Use 'with (p)' inside a dynamic 'new Function' to emulate global mode.
    // 'new Function' executes in the global scope outside of strict mode, 
    // which allows the 'with' statement to work correctly.
    // This dynamically proxies all unresolved variables (like width, height, fill, etc.) to 'p'.
    const wrapper = new Function('p', `
      with (p) {
        ${code}
        
        if (typeof setup === 'function') p.setup = setup;
        if (typeof draw === 'function') p.draw = draw;
        if (typeof windowResized === 'function') p.windowResized = windowResized;
        if (typeof preload === 'function') p.preload = preload;
        if (typeof mousePressed === 'function') p.mousePressed = mousePressed;
        if (typeof mouseReleased === 'function') p.mouseReleased = mouseReleased;
        if (typeof mouseMoved === 'function') p.mouseMoved = mouseMoved;
        if (typeof mouseDragged === 'function') p.mouseDragged = mouseDragged;
        if (typeof keyPressed === 'function') p.keyPressed = keyPressed;
        if (typeof keyReleased === 'function') p.keyReleased = keyReleased;
        if (typeof doubleClicked === 'function') p.doubleClicked = doubleClicked;
      }
    `);

    // 3. Execute the wrapper, giving it the instance 'p'
    try {
      wrapper(p);
    } catch (err) {
      console.error('Error executing p5 sketch:', err);
    }
  }

  p5Instance = new p5(sketch, sketchContainer.value)
})

onUnmounted(() => {
  if (p5Instance) p5Instance.remove()
})
</script>

<template>
  <div ref="sketchContainer" class="p5-container" :class="{ 'pixelated': props.pixelated }"></div>
</template>

<style scoped>
.p5-container {
  width: 100%;
  display: flex;
  justify-content: center;
  margin: 0;
}
/* This ensures the canvas itself is responsive */
.p5-container :deep(canvas) {
  width: 100% !important;
  max-width: 100%;
  height: auto !important;
}
.p5-container.pixelated :deep(canvas) {
  image-rendering: -moz-crisp-edges;
  image-rendering: -webkit-optimize-contrast;
  image-rendering: crisp-edges;
  image-rendering: pixelated;
}
</style>
