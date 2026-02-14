import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import Login from './views/Login.vue'
import ScreeningList from './views/ScreeningList.vue'
import SeatMap from './views/SeatMap.vue'
import Admin from './views/Admin.vue'

const routes = [
  { path: '/', redirect: '/screenings' },
  { path: '/login', component: Login },
  { path: '/screenings', component: ScreeningList, meta: { auth: true } },
  { path: '/screenings/:id', component: SeatMap, meta: { auth: true } },
  { path: '/admin', component: Admin, meta: { auth: true, admin: true } },
]

const router = createRouter({ history: createWebHistory(), routes })

router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('token')
  const role = localStorage.getItem('role')
  if (to.meta.auth && !token) {
    next('/login')
    return
  }
  if (to.meta.admin && role !== 'ADMIN') {
    next('/screenings')
    return
  }
  next()
})

createApp(App).use(router).mount('#app')
