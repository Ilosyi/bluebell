<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowBigDown, ArrowBigUp, CornerDownLeft, MessagesSquare } from 'lucide-vue-next'
import ForumHeader from '../components/ForumHeader.vue'
import ForumSidebar from '../components/ForumSidebar.vue'
import { getToken } from '../api/auth'
import { fetchCommunities, fetchPostDetail, votePost } from '../api/forum'

const route = useRoute()
const router = useRouter()

const post = ref(null)
const communities = ref([])
const activeCommunity = ref(null)
const loading = ref(false)
const voteLoading = ref(false)
const errorMessage = ref('')

const formattedTime = computed(() => {
  if (!post.value?.createTime) return '刚刚发布'
  const date = new Date(post.value.createTime)
  if (Number.isNaN(date.getTime())) return '刚刚发布'
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(
    date.getDate()
  ).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
})

async function loadPage() {
  loading.value = true
  errorMessage.value = ''
  try {
    const [postDetail, communityItems] = await Promise.all([
      fetchPostDetail(route.params.id),
      fetchCommunities()
    ])
    post.value = postDetail
    communities.value = communityItems
    activeCommunity.value = communityItems.find((item) => item.id === postDetail.communityID) || null
  } catch (error) {
    errorMessage.value = error.message
  } finally {
    loading.value = false
  }
}

async function submitVote(direction) {
  if (!getToken()) {
    ElMessage.warning('请先登录后投票')
    router.push('/login')
    return
  }
  if (!post.value?.id) return

  voteLoading.value = true
  try {
    await votePost(post.value.id, direction)
    ElMessage.success(direction === 0 ? '已取消投票' : '投票成功')
    post.value = await fetchPostDetail(post.value.id)
  } catch (error) {
    ElMessage.error(error.message)
  } finally {
    voteLoading.value = false
  }
}

onMounted(loadPage)
</script>

<template>
  <div class="forum-page">
    <ForumHeader />

    <main class="forum-layout">
      <section class="forum-main">
        <div class="detail-shell">
          <button class="back-link" type="button" @click="$router.push('/')">
            <CornerDownLeft :size="15" />
            <span>返回主题流</span>
          </button>

          <section v-if="loading" class="topic-empty">正在加载帖子详情...</section>
          <section v-else-if="errorMessage" class="topic-empty topic-empty--error">{{ errorMessage }}</section>
          <section v-else-if="post" class="detail-panel">
            <header class="detail-panel__head">
              <div class="detail-panel__meta">
                <span class="topic-chip">{{ post.communityName }}</span>
                <span>{{ post.authorName }}</span>
                <span>/</span>
                <span>{{ formattedTime }}</span>
              </div>
              <h1>{{ post.title }}</h1>
            </header>

            <div class="detail-panel__body">
              <p>{{ post.content }}</p>
            </div>

            <footer class="detail-panel__foot">
              <div class="detail-vote">
                <button class="toolbar-button" type="button" :disabled="voteLoading" @click="submitVote(1)">
                  <ArrowBigUp :size="16" />
                  <span>赞成</span>
                </button>
                <strong>{{ post.voteNum }}</strong>
                <button class="toolbar-button" type="button" :disabled="voteLoading" @click="submitVote(-1)">
                  <ArrowBigDown :size="16" />
                  <span>反对</span>
                </button>
                <button class="toolbar-button" type="button" :disabled="voteLoading" @click="submitVote(0)">
                  取消
                </button>
              </div>

              <div class="detail-reply-hint">
                <MessagesSquare :size="16" />
                <span>当前后端还没有评论接口，这里先保留帖子阅读和投票工作流。</span>
              </div>
            </footer>
          </section>
        </div>
      </section>

      <ForumSidebar
        :communities="communities"
        :active-community-id="post?.communityID || ''"
        :active-community="activeCommunity"
        @change-community="(id) => $router.push(id ? { path: '/', query: { community: id } } : '/')"
      />
    </main>
  </div>
</template>
