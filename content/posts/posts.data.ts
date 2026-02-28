import { createContentLoader } from 'vitepress'

interface Post {
    title: string
    url: string
    description?: string
    tags?: string[]
    lastUpdated?: string
    image?: string
}

declare const data: Post[]
export { data }

export default createContentLoader('posts/*/*.md', {
    transform(raw): Post[] {
        return raw.map(({ url, frontmatter }) => ({
            title: frontmatter.title || 'Untitled',
            description: frontmatter.description || '',
            tags: frontmatter.tags || [],
            lastUpdated: frontmatter.lastUpdated || '',
            image: frontmatter.image || '',
            url,
        })).sort((a, b) => {
            const dateA = a.lastUpdated ? new Date(a.lastUpdated).getTime() : 0;
            const dateB = b.lastUpdated ? new Date(b.lastUpdated).getTime() : 0;
            return dateB - dateA;
        })
    }
})
