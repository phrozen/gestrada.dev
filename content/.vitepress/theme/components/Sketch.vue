<script setup>
import { computed } from 'vue'

import IconBrandGithub from '~icons/tabler/brand-github'

const props = defineProps({
  title: String,
  description: String,
  source: String // e.g. "docs/posts/the-game-of-life/colorful-life.js" 
})

const githubUrl = computed(() => {
  if (!props.source) return ''
  return `https://github.com/phrozen/gestrada.dev/blob/main/${props.source.replace(/^\/+/, '')}`
})
</script>

<template>
  <div class="sketch-wrapper">
    <div class="sketch-header">
      <span class="sketch-title">{{ title }}</span>
      <a v-if="source" :href="githubUrl" target="_blank" rel="noopener noreferrer" class="sketch-link">
        <IconBrandGithub />
        View on GitHub
      </a>
    </div>
    
    <slot></slot>
    
    <div class="sketch-footer">
      {{ description }}
    </div>
  </div>
</template>

<style scoped>
.sketch-wrapper {
  margin: 1.5rem 0;
  border: 1px solid var(--vp-c-divider);
  border-radius: 8px;
  overflow: hidden;
  background-color: var(--vp-c-bg-soft);
}
.sketch-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 1rem;
  background-color: var(--vp-c-bg-mute);
  border-bottom: 1px solid var(--vp-c-divider);
  font-family: var(--vp-font-family-mono);
  font-size: 0.9em;
}
.sketch-title {
  font-weight: bold;
}
.sketch-link {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  color: var(--vp-c-text-2);
  text-decoration: none;
  font-size: 0.85em;
  transition: color 0.2s;
}
.sketch-link:hover {
  color: var(--vp-c-brand-1);
}
.github-icon {
  opacity: 0.8;
}
.sketch-footer {
  padding: 0.75rem 1rem;
  text-align: center;
  font-size: 0.9em;
  font-style: italic;
  color: var(--vp-c-text-2);
  background-color: var(--vp-c-bg-mute);
  border-top: 1px solid var(--vp-c-divider);
}
</style>
