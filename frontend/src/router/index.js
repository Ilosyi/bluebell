import { createRouter, createWebHistory } from 'vue-router'
import { ElMessage } from 'element-plus'
import { setAuthRedirectHandler } from '../api/auth'
import Home from '../views/Home.vue'
import Login from '../views/Login.vue'
import Signup from '../views/Signup.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: Home
    },
    {
      path: '/login',
      name: 'login',
      component: Login
    },
    {
      path: '/signup',
      name: 'signup',
      component: Signup
    }
  ]
})

setAuthRedirectHandler(() => {
  if (router.currentRoute.value.path !== '/login') {
    ElMessage.warning('登录状态已失效，请重新登录')
    router.push('/login')
  }
})

export default router
