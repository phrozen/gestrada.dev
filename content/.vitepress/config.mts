import { defineConfig } from 'vitepress'
import Icons from 'unplugin-icons/vite'
import { viteStaticCopy } from 'vite-plugin-static-copy'

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
        const routePath = pageData.relativePath.replace(/((^|\/)index)?\.md$/, '$2')
        return [
            ['link', { rel: 'icon', type: 'image/png', href: '/favicon.png' }],
            ['meta', { property: 'og:title', content: pageData.frontmatter.title || title }],
            ['meta', { property: 'og:description', content: pageData.frontmatter.description || description }],
            ['meta', { property: 'og:url', content: `https://gestrada.dev/${routePath}` }],
            ['meta', { property: 'og:image', content: `https://gestrada.dev/${routePath}cover.webp` }]
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
            }),
            viteStaticCopy({
                structured: true,
                targets: [
                    {
                        src: 'posts/**/*.{webp,js,go,wasm}',
                        dest: './'
                    }
                ]
            })
        ]
    }
})
