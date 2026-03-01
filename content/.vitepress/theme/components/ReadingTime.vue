<script setup>
import { useData } from 'vitepress'
const { frontmatter } = useData()
</script>

<template>
  <span v-if="frontmatter.readingTime" class="reading-time">
    {{ frontmatter.readingTime }}
  </span>
</template>

<style scoped>
.reading-time {
  display: inline-block;
  font-size: 0.9rem;
  color: var(--vp-c-text-2);
  font-style: italic;
  margin-bottom: 1rem;
}
</style>
