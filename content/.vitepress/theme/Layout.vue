<script setup>
import DefaultTheme from 'vitepress/theme'
import ReadingTime from './components/ReadingTime.vue'
import SourceLink from './components/SourceLink.vue'
import { useData } from 'vitepress'

const { Layout } = DefaultTheme
const { frontmatter, page } = useData()
</script>

<template>
  <Layout>
    <template #doc-before>
      <ReadingTime v-if="frontmatter.readingTime" />
    </template>
    <template #doc-after>
      <SourceLink v-if="page.relativePath.startsWith('posts/')" />
    </template>
  </Layout>
</template>
