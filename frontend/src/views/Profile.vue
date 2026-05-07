<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Edit3, FileText, Send, Trash2, UserRound } from 'lucide-vue-next'
import ForumHeader from '../components/ForumHeader.vue'
import { fetchMe, updateMe, useAuthState } from '../api/auth'
import { deletePost, fetchMyPosts, publishDraft } from '../api/forum'

const router = useRouter()
const { user } = useAuthState()
const profileLoading = ref(false)
const savingProfile = ref(false)
const postLoading = ref(false)
const activeStatus = ref(1)
const page = ref(1)
const PAGE_SIZE = 8
const posts = ref([])
const pagination = ref({ page: 1, size: PAGE_SIZE, total: 0, totalPages: 1, hasMore: false })

const profileForm = reactive({
  nickname: '',
  avatar_url: '',
  bio: ''
})

const displayName = computed(() => {
  return profileForm.nickname || user.value?.nickname || user.value?.user_name || user.value?.username || '风铃草用户'
})

const username = computed(() => user.value?.username || user.value?.user_name || '')
const userID = computed(() => user.value?.user_id || '')
const avatarInitial = computed(() => displayName.value.slice(0, 1).toUpperCase())
const statusLabel = computed(() => (activeStatus.value === 1 ? '我的帖子' : '草稿箱'))

function formatDate(value) {
  if (!value) return '刚刚'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '刚刚'
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

async function loadProfile() {
  profileLoading.value = true
  try {
    const profile = await fetchMe()
    profileForm.nickname = profile.nickname || profile.username || ''
    profileForm.avatar_url = profile.avatar_url || ''
    profileForm.bio = profile.bio || ''
  } catch (error) {
    ElMessage.error(error.message)
  } finally {
    profileLoading.value = false
  }
}

async function saveProfile() {
  savingProfile.value = true
  try {
    await updateMe({
      nickname: profileForm.nickname.trim(),
      avatar_url: profileForm.avatar_url.trim(),
      bio: profileForm.bio.trim()
    })
    ElMessage.success('资料已保存')
  } catch (error) {
    ElMessage.error(error.message)
  } finally {
    savingProfile.value = false
  }
}

async function loadPosts(nextPage = page.value) {
  postLoading.value = true
  page.value = nextPage
  try {
    const result = await fetchMyPosts({
      page: page.value,
      size: PAGE_SIZE,
      status: activeStatus.value
    })
    posts.value = result.items
    pagination.value = {
      page: result.pagination.page || page.value,
      size: result.pagination.size || PAGE_SIZE,
      total: result.pagination.total || 0,
      totalPages: result.pagination.total_pages || result.pagination.totalPages || 1,
      hasMore: Boolean(result.pagination.has_more ?? result.pagination.hasMore)
    }
  } catch (error) {
    posts.value = []
    ElMessage.error(error.message)
  } finally {
    postLoading.value = false
  }
}

function changeStatus(status) {
  activeStatus.value = status
  loadPosts(1)
}

function editPost(post) {
  router.push(`/post/${post.id}/edit`)
}

async function removePost(post) {
  try {
    await ElMessageBox.confirm(`确认删除「${post.title || '未命名草稿'}」吗？删除后不可恢复。`, '删除确认', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消'
    })
    await deletePost(post.id)
    ElMessage.success('已删除')
    await loadPosts(posts.value.length === 1 && page.value > 1 ? page.value - 1 : page.value)
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

async function publish(post) {
  try {
    await publishDraft(post.id)
    ElMessage.success('草稿已发布')
    await loadPosts(page.value)
  } catch (error) {
    ElMessage.error(error.message)
  }
}

onMounted(async () => {
  await Promise.all([loadProfile(), loadPosts(1)])
})
</script>

<template>
  <div class="forum-page">
    <ForumHeader />

    <main class="profile-layout">
      <section class="profile-hero">
        <div class="profile-hero__identity">
          <div class="profile-avatar">
            <img v-if="profileForm.avatar_url" :src="profileForm.avatar_url" :alt="displayName" />
            <span v-else>{{ avatarInitial }}</span>
          </div>
          <div>
            <p class="topic-toolbar__eyebrow">账号中心</p>
            <h1>{{ displayName }}</h1>
            <p>@{{ username }} · UID {{ userID }}</p>
          </div>
        </div>
        <button class="site-action site-action--primary" type="button" @click="router.push('/new')">
          <Send :size="16" />
          <span>发布主题</span>
        </button>
      </section>

      <section class="profile-grid">
        <section class="profile-card" v-loading="profileLoading">
          <header class="profile-card__head">
            <div>
              <p class="topic-toolbar__eyebrow">Profile</p>
              <h2>用户资料</h2>
            </div>
            <UserRound :size="20" />
          </header>

          <el-form label-position="top" class="editor-form">
            <el-form-item label="昵称">
              <el-input v-model.trim="profileForm.nickname" maxlength="32" show-word-limit placeholder="给自己起一个论坛昵称" />
            </el-form-item>
            <el-form-item label="头像 URL">
              <el-input v-model.trim="profileForm.avatar_url" maxlength="512" placeholder="https://example.com/avatar.png" />
            </el-form-item>
            <el-form-item label="个人简介">
              <el-input
                v-model="profileForm.bio"
                type="textarea"
                :rows="5"
                maxlength="160"
                show-word-limit
                resize="none"
                placeholder="写一句你的技术兴趣、当前方向或签名。"
              />
            </el-form-item>
            <button class="site-action site-action--primary profile-save" type="button" :disabled="savingProfile" @click="saveProfile">
              {{ savingProfile ? '保存中...' : '保存资料' }}
            </button>
          </el-form>
        </section>

        <section class="profile-card profile-card--wide">
          <header class="profile-card__head">
            <div>
              <p class="topic-toolbar__eyebrow">Posts</p>
              <h2>{{ statusLabel }}</h2>
            </div>
            <div class="segmented-tabs" role="tablist" aria-label="帖子管理">
              <button class="segmented-tabs__item" :class="{ 'is-active': activeStatus === 1 }" type="button" @click="changeStatus(1)">
                我的帖子
              </button>
              <button class="segmented-tabs__item" :class="{ 'is-active': activeStatus === 0 }" type="button" @click="changeStatus(0)">
                草稿箱
              </button>
            </div>
          </header>

          <div class="profile-posts">
            <div v-if="postLoading" class="topic-empty">正在加载{{ statusLabel }}...</div>
            <div v-else-if="!posts.length" class="topic-empty">这里还没有内容。</div>
            <template v-else>
              <article v-for="post in posts" :key="post.id" class="profile-post-row">
                <div class="profile-post-row__main">
                  <div class="topic-row__meta">
                    <span class="topic-chip">{{ post.communityName }}</span>
                    <span>{{ formatDate(post.createTime) }}</span>
                    <span v-if="activeStatus === 1">{{ post.voteNum }} 赞成</span>
                  </div>
                  <h3>{{ post.title || '未命名草稿' }}</h3>
                  <p>{{ post.content || '草稿正文还没有填写。' }}</p>
                </div>

                <div class="profile-post-row__actions">
                  <button v-if="activeStatus === 1" class="toolbar-button" type="button" @click="router.push(`/post/${post.id}`)">
                    <FileText :size="15" />
                    <span>查看</span>
                  </button>
                  <button class="toolbar-button" type="button" @click="editPost(post)">
                    <Edit3 :size="15" />
                    <span>编辑</span>
                  </button>
                  <button v-if="activeStatus === 0" class="toolbar-button" type="button" @click="publish(post)">
                    <Send :size="15" />
                    <span>发布</span>
                  </button>
                  <button class="toolbar-button toolbar-button--danger" type="button" @click="removePost(post)">
                    <Trash2 :size="15" />
                    <span>删除</span>
                  </button>
                </div>
              </article>
            </template>
          </div>

          <footer class="topic-stream__pagination profile-pagination">
            <span class="topic-stream__pagination-label">共 {{ pagination.total }} 条，第 {{ page }} / {{ pagination.totalPages || 1 }} 页</span>
            <div class="pagination-bar">
              <button class="toolbar-button" type="button" :disabled="page <= 1 || postLoading" @click="loadPosts(page - 1)">上一页</button>
              <button class="toolbar-button" type="button" :disabled="!pagination.hasMore || postLoading" @click="loadPosts(page + 1)">下一页</button>
            </div>
          </footer>
        </section>
      </section>
    </main>
  </div>
</template>
