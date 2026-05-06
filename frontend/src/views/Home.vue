<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ChevronLeft, ChevronRight, RefreshCcw } from 'lucide-vue-next'
import ForumHeader from '../components/ForumHeader.vue'
import ForumSidebar from '../components/ForumSidebar.vue'
import PostListItem from '../components/PostListItem.vue'
import { fetchCommunities, fetchCommunityDetail, fetchPosts } from '../api/forum'
import { forumConfig } from '../config/forum'

const loading = ref(false)
const communities = ref([])
const posts = ref([])
const activeOrder = ref('time')
const activeCommunityID = ref('')
const activeCommunity = ref(null)
const errorMessage = ref('')
const page = ref(1)
const pageSize = ref(10)
const initialized = ref(false)
const route = useRoute()
const router = useRouter()

const sortTabs = [
  { label: '最新', value: 'time' },
  { label: '最热', value: 'score' }
]

const activeCommunityLabel = computed(() => {
  if (!activeCommunityID.value) return '全部主题'
  return communities.value.find((item) => item.id === activeCommunityID.value)?.name || '当前节点'
})

const pageSummary = computed(() => `第 ${page.value} 页`)
const hasPreviousPage = computed(() => page.value > 1)
const maybeHasNextPage = computed(() => posts.value.length >= pageSize.value)
const hotTopics = computed(() => posts.value.slice(0, 4))

async function loadCommunities() {
  communities.value = await fetchCommunities()
}

async function loadActiveCommunity() {
  if (!activeCommunityID.value) {
    activeCommunity.value = null
    return
  }
  activeCommunity.value = await fetchCommunityDetail(activeCommunityID.value)
}

async function loadPosts() {
  loading.value = true
  errorMessage.value = ''
  try {
    posts.value = await fetchPosts({
      order: activeOrder.value,
      communityID: activeCommunityID.value,
      page: page.value,
      size: pageSize.value
    })
  } catch (error) {
    posts.value = []
    errorMessage.value = error.message
  } finally {
    loading.value = false
  }
}

function changeOrder(order) {
  activeOrder.value = order
  page.value = 1
}

function changeCommunity(id) {
  activeCommunityID.value = id
  page.value = 1
  router.replace({
    path: '/',
    query: {
      ...(id ? { community: id } : {}),
      ...(activeOrder.value !== 'time' ? { order: activeOrder.value } : {}),
      ...(page.value > 1 ? { page: String(page.value) } : {})
    }
  })
}

function updateQuery() {
  router.replace({
    path: '/',
    query: {
      ...(activeCommunityID.value ? { community: activeCommunityID.value } : {}),
      ...(activeOrder.value !== 'time' ? { order: activeOrder.value } : {}),
      ...(page.value > 1 ? { page: String(page.value) } : {})
    }
  })
}

function goPage(nextPage) {
  if (nextPage < 1) return
  page.value = nextPage
}

async function refreshAll() {
  try {
    await Promise.all([loadCommunities(), loadPosts(), loadActiveCommunity()])
  } catch (error) {
    ElMessage.error(error.message || '页面刷新失败')
  }
}

watch([activeOrder, activeCommunityID, page], () => {
  if (!initialized.value) return
  updateQuery()
  loadPosts()
})

watch(activeCommunityID, () => {
  if (!initialized.value) return
  loadActiveCommunity().catch((error) => {
    activeCommunity.value = null
    ElMessage.error(error.message || '社区详情加载失败')
  })
})

onMounted(async () => {
  if (typeof route.query.community === 'string') {
    activeCommunityID.value = route.query.community
  }
  if (typeof route.query.order === 'string' && ['time', 'score'].includes(route.query.order)) {
    activeOrder.value = route.query.order
  }
  if (typeof route.query.page === 'string') {
    const parsed = Number(route.query.page)
    page.value = parsed > 0 ? parsed : 1
  }
  await refreshAll()
  initialized.value = true
})
</script>

<template>
  <div class="forum-page">
    <ForumHeader />

    <main class="forum-layout">
      <section class="forum-main">
        <section class="forum-board">
          <div class="forum-board__primary">
            <div class="forum-board__tabs">
              <button
                class="forum-board__tab"
                :class="{ 'is-active': activeCommunityID === '' }"
                type="button"
                @click="changeCommunity('')"
              >
                {{ forumConfig.home.allTabLabel }}
              </button>
              <button
                v-for="community in communities"
                :key="community.id"
                class="forum-board__tab"
                :class="{ 'is-active': activeCommunityID === community.id }"
                type="button"
                @click="changeCommunity(community.id)"
              >
                {{ community.name }}
              </button>
            </div>

            <div v-if="activeCommunity" class="forum-board__subtabs">
              <span>{{ activeCommunity.introduction || forumConfig.sidebar.missingCommunityIntro }}</span>
            </div>
            <div v-else class="forum-board__subtabs">
              <span>{{ forumConfig.home.intro }}</span>
            </div>
          </div>
        </section>

        <header class="topic-toolbar">
          <div>
            <p class="topic-toolbar__eyebrow">{{ forumConfig.home.streamEyebrow }}</p>
            <h2>{{ activeCommunityLabel }}</h2>
          </div>

          <div class="topic-toolbar__actions">
            <div class="segmented-tabs" role="tablist" aria-label="排序方式">
              <button
                v-for="tab in sortTabs"
                :key="tab.value"
                class="segmented-tabs__item"
                :class="{ 'is-active': activeOrder === tab.value }"
                type="button"
                @click="changeOrder(tab.value)"
              >
                {{ tab.label }}
              </button>
            </div>
            <button class="toolbar-button" type="button" @click="refreshAll">
              <RefreshCcw :size="15" />
              <span>刷新</span>
            </button>
          </div>
        </header>

        <section class="topic-stream">
          <header class="topic-stream__head">
            <div class="topic-stream__title">{{ activeOrder === 'time' ? forumConfig.home.recentTitle : forumConfig.home.hotTitle }}</div>
            <div class="topic-stream__meta">
              <span class="topic-stream__count">{{ posts.length }} 条主题</span>
              <span class="topic-stream__count">{{ pageSummary }}</span>
            </div>
          </header>

          <div v-if="loading" class="topic-empty">正在加载帖子...</div>
          <div v-else-if="errorMessage" class="topic-empty topic-empty--error">{{ errorMessage }}</div>
          <div v-else-if="!posts.length" class="topic-empty">{{ forumConfig.home.emptyMessage }}</div>
          <div v-else class="topic-stream__list">
            <PostListItem v-for="post in posts" :key="post.id" :post="post" />
          </div>

          <footer class="topic-stream__pagination">
            <button class="toolbar-button" type="button" :disabled="!hasPreviousPage || loading" @click="goPage(page - 1)">
              <ChevronLeft :size="15" />
              <span>上一页</span>
            </button>
            <span class="topic-stream__pagination-label">{{ pageSummary }}</span>
            <button class="toolbar-button" type="button" :disabled="!maybeHasNextPage || loading" @click="goPage(page + 1)">
              <span>下一页</span>
              <ChevronRight :size="15" />
            </button>
          </footer>
        </section>
      </section>

      <ForumSidebar
        :communities="communities"
        :active-community-id="activeCommunityID"
        :active-community="activeCommunity"
        :hot-topics="hotTopics"
        @change-community="changeCommunity"
      />
    </main>
  </div>
</template>
