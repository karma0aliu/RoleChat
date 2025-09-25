import { createRouter, createWebHistory } from 'vue-router'
import LoginVue from '../components/Login.vue'
import MainLayout from '../layouts/MainLayout.vue'
import NewChat from '../views/NewChat.vue'
import Chats from '../views/Chats.vue'

const routes = [
  {
    path: '/',
    redirect: '/login',
  },
  {
      path: '/login',
      name: 'Login',
      component: LoginVue
  },
  {
    path: '/api',
    component: MainLayout,
    children: [
      { path: '', redirect: '/api/new' },
      { path: 'new', name: 'NewChat', component: NewChat },
      { path: 'chats', name: 'Chats', component: Chats },
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router