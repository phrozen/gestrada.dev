import type { Theme } from 'vitepress'
import DefaultTheme from 'vitepress/theme'
import './style.css'
import Sketch from './components/Sketch.vue'
import P5Embed from './components/P5Embed.vue'
import ReadingTime from './components/ReadingTime.vue'
import Layout from './Layout.vue'

export default {
    extends: DefaultTheme,
    Layout,
    enhanceApp({ app }) {
        app.component('Sketch', Sketch)
        app.component('P5Embed', P5Embed)
        app.component('ReadingTime', ReadingTime)
    }
} satisfies Theme
