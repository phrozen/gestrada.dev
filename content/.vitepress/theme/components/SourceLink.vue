<script setup>
import { computed } from 'vue'
import { useData } from 'vitepress'
import IconBrandGithub from '~icons/tabler/brand-github'

const { page } = useData()

const githubUrl = computed(() => {
  // Get the directory from the relative path (e.g. "posts/bloom-filter/index.md" → "posts/bloom-filter")
  const dir = page.value.relativePath.replace(/\/[^/]+$/, '')
  return `https://github.com/phrozen/gestrada.dev/tree/main/content/${dir}`
})
</script>

<template>
  <div class="source-link">
    <a :href="githubUrl" target="_blank" rel="noopener noreferrer">
      <IconBrandGithub /> View source on GitHub
    </a>
  </div>
</template>

<style scoped>
.source-link {
  margin-top: 2rem;
  padding-top: 1rem;
  border-top: 1px solid var(--vp-c-divider);
}
.source-link a {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  color: var(--vp-c-text-2);
  text-decoration: none;
  font-size: 0.9rem;
  transition: color 0.2s;
}
.source-link a:hover {
  color: var(--vp-c-brand-1);
}
</style>
