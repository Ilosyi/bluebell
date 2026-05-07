<script setup>
import { onMounted, ref } from 'vue'
import { RefreshCcw } from 'lucide-vue-next'
import ForumHeader from '../components/ForumHeader.vue'
import ForumSidebar from '../components/ForumSidebar.vue'
import PostListItem from '../components/PostListItem.vue'
import { useForumList } from '../composables/useForumList'
import { forumConfig } from '../config/forum'

const jumpPageInput = ref('')
const {
  PAGE_SIZE,
  loading,
  communities,
  posts,
  activeOrder,
  activeCommunityID,
  activeCommunity,
  activeKeyword,
  errorMessage,
  page,
  total,
  totalPages,
  sortTabs,
  activeCommunityLabel,
  pageSummary,
  hasPreviousPage,
  hotTopics,
  visiblePages,
  changeOrder,
  changeCommunity,
  changeKeyword,
  goPage,
  refreshAll,
  initialize
} = useForumList()

function goInputPage() {
  const value = Number(jumpPageInput.value)
  if (!Number.isInteger(value)) return
  goPage(value)
}

onMounted(async () => {
  await initialize()
})
</script>

<template>
  <div class="forum-page">
    <ForumHeader />

    <main class="forum-layout forum-layout--home">
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
              <span class="topic-stream__count">每页 {{ PAGE_SIZE }} 条</span>
              <span class="topic-stream__count">共 {{ total }} 条</span>
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
            <div class="pagination-bar">
              <button
                v-for="pageItem in visiblePages"
                :key="pageItem"
                class="pagination-bar__item"
                :class="{ 'is-active': pageItem === page }"
                :disabled="pageItem === 'ellipsis-left' || pageItem === 'ellipsis-right' || loading"
                type="button"
                @click="typeof pageItem === 'number' && goPage(pageItem)"
              >
                {{ pageItem === 'ellipsis-left' || pageItem === 'ellipsis-right' ? '…' : pageItem }}
              </button>
            </div>

            <div class="pagination-jump">
              <span class="topic-stream__pagination-label">共 {{ totalPages }} 页</span>
              <el-input
                v-model="jumpPageInput"
                class="pagination-jump__input"
                placeholder="页码"
                @keyup.enter="goInputPage"
              />
              <button class="toolbar-button" type="button" :disabled="loading" @click="goInputPage">跳转</button>
            </div>
          </footer>
        </section>
      </section>

      <ForumSidebar
        class="forum-sidebar--compact"
        :communities="communities"
        :active-community-id="activeCommunityID"
        :active-community="activeCommunity"
        :search-keyword="activeKeyword"
        :hot-topics="hotTopics"
        @change-community="changeCommunity"
        @search="changeKeyword"
      />
    </main>
  </div>
</template>
