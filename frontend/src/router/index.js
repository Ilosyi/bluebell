import { createRouter, createWebHistory } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getToken, setAuthRedirectHandler } from '../api/auth'
import Home from '../views/Home.vue'
import Login from '../views/Login.vue'
import NewPost from '../views/NewPost.vue'
import PostDetail from '../views/PostDetail.vue'
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
      path: '/post/:id',
      name: 'post-detail',
      component: PostDetail
    },
    {
      path: '/new',
      name: 'new-post',
      component: NewPost
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
    router.push({
      path: '/login',
      query: router.currentRoute.value.fullPath !== '/' ? { redirect: router.currentRoute.value.fullPath } : {}
    })
  }
})

router.beforeEach((to) => {
  if (to.name === 'new-post' && !getToken()) {
    ElMessage.warning('请先登录后发帖')
    return {
      path: '/login',
      query: { redirect: to.fullPath }
    }
  }
  return true
})

export default router
