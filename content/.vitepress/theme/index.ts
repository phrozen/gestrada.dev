import type { Theme } from 'vitepress'
import DefaultTheme from 'vitepress/theme'
import './style.css'
import Sketch from './components/Sketch.vue'
import P5Embed from './components/P5Embed.vue'

export default {
    extends: DefaultTheme,
    enhanceApp({ app }) {
        app.component('Sketch', Sketch)
        app.component('P5Embed', P5Embed)
    }
} satisfies Theme
