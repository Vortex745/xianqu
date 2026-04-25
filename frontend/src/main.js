import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import NativeUI from '@/ui/native-components'
import '@/ui/native-ui.css'
import '@/assets/admin-redesign.scss'

// GSAP global registration
import gsap from 'gsap'
import { ScrollTrigger } from 'gsap/ScrollTrigger'

gsap.registerPlugin(ScrollTrigger)

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(NativeUI)

app.mount('#app')
