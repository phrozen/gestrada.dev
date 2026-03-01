<script setup>
import DefaultTheme from 'vitepress/theme'
import ReadingTime from './components/ReadingTime.vue'
import { useData } from 'vitepress'

const { Layout } = DefaultTheme
const { frontmatter } = useData()
</script>

<template>
  <Layout>
    <template #doc-before>
      <ReadingTime v-if="frontmatter.readingTime" />
    </template>
  </Layout>
</template>
