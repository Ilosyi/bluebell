<script setup>
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { Leaf, PenSquare, Search, Sparkles, X } from 'lucide-vue-next'
import heroUrl from '../assets/hero.png'
import { useAuthState } from '../api/auth'
import { forumConfig } from '../config/forum'

const props = defineProps({
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
  },
  searchKeyword: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['change-community', 'search'])

const router = useRouter()
const { isLoggedIn } = useAuthState()
const searchInput = ref(props.searchKeyword)

watch(
  () => props.searchKeyword,
  (value) => {
    searchInput.value = value
  }
)

function selectCommunity(id) {
  emit('change-community', id)
}

function submitSearch() {
  emit('search', searchInput.value)
}

function clearSearch() {
  searchInput.value = ''
  emit('search', '')
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

    <section class="sidebar-panel sidebar-panel--search">
      <div class="sidebar-panel__head">
        <h3>{{ forumConfig.sidebar.searchTitle }}</h3>
        <Search :size="15" />
      </div>

      <div class="sidebar-search">
        <div class="sidebar-search__input">
          <Search :size="15" />
          <input
            v-model.trim="searchInput"
            type="search"
            :placeholder="forumConfig.sidebar.searchPlaceholder"
            @keyup.enter="submitSearch"
          />
          <button v-if="searchInput" type="button" aria-label="清空搜索" @click="clearSearch">
            <X :size="14" />
          </button>
        </div>
        <button class="sidebar-cta" type="button" @click="submitSearch">
          <Search :size="16" />
          <span>{{ searchInput ? '搜索帖子' : '查看全部' }}</span>
        </button>
      </div>

      <button class="sidebar-search__publish" type="button" @click="openNewPost">
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
