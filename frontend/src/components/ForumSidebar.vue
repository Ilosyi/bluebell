<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { Compass, Leaf, PenSquare, Sparkles } from 'lucide-vue-next'
import heroUrl from '../assets/hero.png'
import { useAuthState } from '../api/auth'
import { forumConfig } from '../config/forum'

defineProps({
  communities: {
    type: Array,
    default: () => []
  },
  hotTopics: {
    type: Array,
    default: () => []
  },
  activeCommunityID: {
    type: String,
    default: ''
  },
  activeCommunity: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['change-community'])

const router = useRouter()
const { user, isLoggedIn } = useAuthState()

const userLabel = computed(() => user.value?.user_name || user.value?.username || '已登录用户')

function selectCommunity(id) {
  emit('change-community', id)
}

function openNewPost() {
  if (isLoggedIn.value) {
    router.push('/new')
    return
  }
  router.push('/login')
}
</script>

<template>
  <aside class="forum-sidebar">
    <section class="sidebar-panel sidebar-panel--hero">
      <div class="sidebar-panel__topline">
        <span>{{ forumConfig.sidebar.topline }}</span>
        <Leaf :size="14" />
      </div>
      <div class="sidebar-hero">
        <div>
          <h2>{{ forumConfig.sidebar.heroTitle }}</h2>
          <p>{{ forumConfig.sidebar.heroDescription }}</p>
        </div>
        <img :src="heroUrl" :alt="forumConfig.brand.fullName" />
      </div>
    </section>

    <section class="sidebar-panel">
      <div class="sidebar-panel__head">
        <h3>{{ forumConfig.sidebar.statusTitle }}</h3>
        <Compass :size="15" />
      </div>

      <div v-if="isLoggedIn" class="sidebar-user">
        <strong>{{ userLabel }}</strong>
        <p>{{ forumConfig.sidebar.loggedInDescription }}</p>
      </div>
      <div v-else class="sidebar-user">
        <strong>{{ forumConfig.sidebar.guestTitle }}</strong>
        <p>{{ forumConfig.sidebar.guestDescription }}</p>
      </div>

      <button class="sidebar-cta" type="button" @click="openNewPost">
        <PenSquare :size="16" />
        <span>{{ isLoggedIn ? forumConfig.sidebar.newPostCta : forumConfig.sidebar.loginPostCta }}</span>
      </button>
    </section>

    <section v-if="activeCommunity" class="sidebar-panel">
      <div class="sidebar-panel__head">
        <h3>{{ forumConfig.sidebar.currentCommunityTitle }}</h3>
        <Sparkles :size="15" />
      </div>
      <div class="sidebar-user">
        <strong>{{ activeCommunity.name }}</strong>
        <p>{{ activeCommunity.introduction || forumConfig.sidebar.missingCommunityIntro }}</p>
      </div>
    </section>

    <section class="sidebar-panel">
      <div class="sidebar-panel__head">
        <h3>{{ forumConfig.sidebar.hotTopicsTitle }}</h3>
        <Sparkles :size="15" />
      </div>
      <div class="sidebar-links">
        <RouterLink v-for="topic in hotTopics" :key="topic.id" class="sidebar-link" :to="`/post/${topic.id}`">
          <span class="sidebar-link__title">{{ topic.title }}</span>
          <span class="sidebar-link__meta">{{ topic.communityName }}</span>
        </RouterLink>
      </div>
    </section>
  </aside>
</template>
