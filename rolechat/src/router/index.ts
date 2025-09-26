import { createRouter, createWebHistory } from 'vue-router'
import { isLoggedIn } from '../auth'
import LoginVue from '../components/Login.vue'
import MainLayout from '../layouts/Mainlayout.vue'
import NewChat from '../views/NewChat.vue'
import Chats from '../views/Chats.vue'

const routes = [
  {
    path: '/',
    component: MainLayout,
    children: [
      { path: '', name: 'NewChatPublic', component: NewChat }
    ]
  },
  { path: '/login', name: 'Login', component: LoginVue },
  {
    path: '/app',
    component: MainLayout,
    children: [
      { path: '', redirect: '/app/new' },
      { path: 'new', name: 'NewChat', component: NewChat },
      { path: 'chats', name: 'Chats', component: Chats },
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, _from, next) => {
  if (to.path.startsWith('/app')) {
    if (!isLoggedIn.value) return next({ path: '/login', query: { redirect: to.fullPath } })
  }
  if (to.path === '/login' && isLoggedIn.value) return next('/app/new')
  next()
})

export default router