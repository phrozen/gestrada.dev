---
layout: home
---

<script setup>
import { computed } from 'vue'
import { data as posts } from './posts/posts.data.ts'
</script>

<div class="post-grid">
  <a v-for="post in posts" :key="post.url" :href="post.url" class="post-card">
    <div class="post-image-container">
      <img :src="post.url.replace(/\.html$/, '') + 'cover.webp'" :alt="post.title" class="post-image" />
    </div>
    <div class="post-content">
      <h3 class="post-title">{{ post.title }}</h3>
      <p v-if="post.description" class="post-desc">{{ post.description }}</p>
      <div v-if="post.tags && post.tags.length" class="post-tags">
        <span v-for="tag in post.tags" :key="tag" class="tag">{{ tag }}</span>
      </div>
    </div>
  </a>
</div>

<style>
.post-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 2rem;
  margin-top: 2rem;
  padding: 0 1rem;
}

.post-card {
  display: flex;
  flex-direction: column;
  background-color: var(--vp-c-bg-soft);
  border: 1px solid var(--vp-c-divider);
  border-radius: 12px;
  overflow: hidden;
  text-decoration: none !important;
  color: inherit !important;
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
  height: 100%;
}

.post-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.1);
  border-color: var(--vp-c-brand-1);
}

.post-image-container {
  width: 100%;
  aspect-ratio: 16 / 9;
  background-color: var(--vp-c-bg-mute);
  border-bottom: 1px solid var(--vp-c-divider);
  overflow: hidden;
}

.post-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s ease;
}

.post-card:hover .post-image {
  transform: scale(1.05);
}

.post-content {
  padding: 1rem;
  display: flex;
  flex-direction: column;
  flex-grow: 1;
}

.post-title {
  margin: 0 0 0.5rem 0 !important;
  font-size: 1.25rem;
  font-weight: 600;
  line-height: 1.4;
  color: var(--vp-c-text-1);
}

.post-card:hover .post-title {
  color: var(--vp-c-brand-1);
}

.post-desc {
  margin: 0 0 1.5rem 0 !important;
  color: var(--vp-c-text-2);
  font-size: 0.95rem;
  line-height: 1.5;
  flex-grow: 1;
}

.post-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: auto;
}

.tag {
  font-size: 0.75rem;
  padding: 0.2rem 0.6rem;
  border-radius: 6px;
  background-color: var(--vp-c-bg-mute);
  color: var(--vp-c-text-2);
  border: 1px solid var(--vp-c-divider);
  transition: color 0.2s, border-color 0.2s;
}

.post-card:hover .tag {
  border-color: var(--vp-c-brand-soft);
}
</style>
