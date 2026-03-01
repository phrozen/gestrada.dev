import { createContentLoader } from 'vitepress'
import readingTime from 'reading-time'

interface Post {
    title: string
    url: string
    description?: string
    tags?: string[]
    lastUpdated?: string
    readingTime?: string
}

declare const data: Post[]
export { data }

export default createContentLoader('posts/*/*.md', {
    includeSrc: true,
    transform(raw): Post[] {
        return raw.map(({ url, frontmatter, src }) => ({
            title: frontmatter.title || 'Untitled',
            description: frontmatter.description || '',
            tags: frontmatter.tags || [],
            lastUpdated: frontmatter.lastUpdated || '',
            readingTime: src ? readingTime(src).text : '',
            url,
        })).sort((a, b) => {
            const dateA = a.lastUpdated ? new Date(a.lastUpdated).getTime() : 0;
            const dateB = b.lastUpdated ? new Date(b.lastUpdated).getTime() : 0;
            return dateB - dateA;
        })
    }
})
