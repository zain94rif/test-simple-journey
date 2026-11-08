import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { setUnauthorizedHandler, clearToken } from './lib/http'
import './style.css'

const app = createApp(App)
const pinia = createPinia()

// One-shot 401 handling: clear the session and send the user to the login
// page, remembering where they were.
setUnauthorizedHandler(() => {
  clearToken()
  const current = router.currentRoute.value
  if (current.name === 'login') return
  void router.replace({
    name: 'login',
    query: { redirect: current.fullPath },
  })
})

app.use(pinia)
app.use(router)
app.mount('#app')
