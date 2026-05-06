<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Home, LogIn, LogOut, PenSquare, ShieldCheck, UserRound } from 'lucide-vue-next'
import { clearAuth, useAuthState } from '../api/auth'
import { forumConfig } from '../config/forum'

const router = useRouter()
const route = useRoute()
const { isLoggedIn, user } = useAuthState()

const userLabel = computed(() => user.value?.user_name || user.value?.username || '已登录用户')

function push(path) {
  router.push(path)
}

function logout() {
  clearAuth()
  if (route.path === '/new') {
    router.push('/')
  }
}
</script>

<template>
  <header class="site-header">
    <div class="site-header__inner">
      <button class="brand-button" type="button" @click="push('/')">
        <span class="brand-button__mark">{{ forumConfig.brand.mark }}</span>
        <span class="brand-button__text">
          <strong>{{ forumConfig.brand.name }}</strong>
          <span>{{ forumConfig.brand.fullName }}</span>
        </span>
      </button>

      <nav class="site-nav" aria-label="主导航">
        <RouterLink class="site-nav__link" to="/">
          <Home :size="14" />
          <span>{{ forumConfig.navigation.home }}</span>
        </RouterLink>
        <RouterLink class="site-nav__link" to="/new">
          <PenSquare :size="14" />
          <span>{{ forumConfig.navigation.newPost }}</span>
        </RouterLink>
      </nav>

      <div class="site-header__actions">
        <template v-if="isLoggedIn">
          <div class="user-chip">
            <ShieldCheck :size="15" />
            <span>{{ userLabel }}</span>
          </div>
          <button class="site-action site-action--primary" type="button" @click="push('/new')">
            <PenSquare :size="16" />
            <span>写新主题</span>
          </button>
          <button class="site-action" type="button" @click="logout">
            <LogOut :size="16" />
            <span>退出</span>
          </button>
        </template>

        <template v-else>
          <button class="site-action" type="button" @click="push('/login')">
            <LogIn :size="16" />
            <span>登录</span>
          </button>
          <button class="site-action site-action--primary" type="button" @click="push('/signup')">
            <UserRound :size="16" />
            <span>注册</span>
          </button>
        </template>
      </div>
    </div>
  </header>
</template>
