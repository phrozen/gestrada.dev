import { defineConfig } from 'vitepress'
import Icons from 'unplugin-icons/vite'

const title = "gestrada.dev"
const description = "Thoughts on Computer Science, Clever Algorithms, System Design and Generative Art."

// https://vitepress.dev/reference/site-config
export default defineConfig({
    title,
    description,
    srcDir: '.',
    outDir: '.vitepress/dist',
    cleanUrls: true,
    sitemap: {
        hostname: 'https://gestrada.dev'
    },
    async transformHead({ pageData }) {
        return [
            ['link', { rel: 'icon', type: 'image/png', href: '/favicon.png' }],
            ['meta', { property: 'og:title', content: pageData.frontmatter.title || title }],
            ['meta', { property: 'og:description', content: pageData.frontmatter.description || description }],
            ['meta', { property: 'og:url', content: `https://gestrada.dev/${pageData.relativePath.replace(/((^|\/)index)?\.md$/, '$2')}` }],
            ['meta', { property: 'og:image', content: `https://gestrada.dev${pageData.frontmatter.image || '/covers/default.webp'}` }]
        ]
    },
    themeConfig: {
        // https://vitepress.dev/reference/default-theme-config
        nav: [
            { text: 'Home', link: '/' },
            { text: 'About', link: '/about' }
        ],

        socialLinks: [
            { icon: 'github', link: 'https://github.com/phrozen' },
            { icon: 'x', link: 'https://x.com/phrzn' }
        ],

        search: {
            provider: 'local'
        },

        footer: {
            //message: 'Released under the MIT License.',
            copyright: '© 2026 Guillermo Estrada'
        }
    },
    vite: {
        plugins: [
            Icons({
                compiler: 'vue3'
            })
        ]
    }
})
