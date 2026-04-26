import { createApp } from 'vue'
import { createHead } from '@unhead/vue/client'
import ui from '@nuxt/ui/vue-plugin'
import App from './App.vue'
import router from './router'
import { installNoZoom } from './no-zoom'
import './assets/main.css'

installNoZoom()

const app = createApp(App)

app.use(createHead())
app.use(ui)
app.use(router)

app.mount('#app')
