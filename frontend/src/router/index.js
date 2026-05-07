import { createRouter, createWebHistory } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getToken, setAuthRedirectHandler } from '../api/auth'
import Home from '../views/Home.vue'
import Login from '../views/Login.vue'
import NewPost from '../views/NewPost.vue'
import PostDetail from '../views/PostDetail.vue'
import Profile from '../views/Profile.vue'
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
      path: '/post/:id/edit',
      name: 'edit-post',
      component: NewPost
    },
    {
      path: '/profile',
      name: 'profile',
      component: Profile
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
  if (['new-post', 'edit-post', 'profile'].includes(to.name) && !getToken()) {
    ElMessage.warning(to.name === 'profile' ? '请先登录后查看账号中心' : '请先登录后发帖')
    return {
      path: '/login',
      query: { redirect: to.fullPath }
    }
  }
  return true
})

export default router
