<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { PenSquare } from 'lucide-vue-next'
import ForumHeader from '../components/ForumHeader.vue'
import ForumSidebar from '../components/ForumSidebar.vue'
import {
  createDraft,
  createPost,
  fetchCommunities,
  fetchMyPostDetail,
  publishDraft,
  updateDraft,
  updatePost
} from '../api/forum'
import { forumConfig } from '../config/forum'

const router = useRouter()
const route = useRoute()
const formRef = ref()
const loading = ref(false)
const draftLoading = ref(false)
const pageLoading = ref(false)
const communities = ref([])
const editingStatus = ref(1)

const form = reactive({
  title: '',
  content: '',
  community_id: ''
})

const editingID = computed(() => (route.params.id == null ? '' : String(route.params.id)))
const isEditMode = computed(() => Boolean(editingID.value))
const isDraftEdit = computed(() => isEditMode.value && editingStatus.value === 0)
const pageTitle = computed(() => {
  if (!isEditMode.value) return '发布新主题'
  return isDraftEdit.value ? '编辑草稿' : '编辑主题'
})
const pageLead = computed(() => {
  if (isDraftEdit.value) return '完善草稿内容，确认后可以直接发布。'
  if (isEditMode.value) return '更新标题、正文或所属节点。'
  return '写下你的问题、经验或想法。'
})

const rules = {
  community_id: [{ required: true, message: '请选择节点', trigger: 'change' }],
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
  content: [{ required: true, message: '请输入正文', trigger: 'blur' }]
}

async function loadCommunities() {
  communities.value = await fetchCommunities()
}

async function loadEditingPost() {
  if (!editingID.value) return
  const post = await fetchMyPostDetail(editingID.value)
  editingStatus.value = Number(post.status ?? 1)
  form.title = post.title || ''
  form.content = post.content || ''
  form.community_id = post.communityID ? String(post.communityID) : ''
}

function buildPayload() {
  return {
    title: form.title,
    content: form.content,
    community_id: Number(form.community_id)
  }
}

async function submit() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    const payload = buildPayload()
    if (isEditMode.value && editingStatus.value === 1) {
      const post = await updatePost(editingID.value, payload)
      ElMessage.success('主题已更新')
      router.push(`/post/${post.id}`)
      return
    }
    if (isEditMode.value && editingStatus.value === 0) {
      await updateDraft(editingID.value, payload)
      const post = await publishDraft(editingID.value)
      ElMessage.success('草稿已发布')
      router.push(`/post/${post.id}`)
      return
    }
    await createPost(payload)
    ElMessage.success('主题已发布')
    router.push('/')
  } catch (error) {
    ElMessage.error(error.message)
  } finally {
    loading.value = false
  }
}

async function saveDraft() {
  if (!form.community_id) {
    ElMessage.warning('请先选择节点')
    return
  }
  draftLoading.value = true
  try {
    const payload = buildPayload()
    if (isEditMode.value && editingStatus.value === 0) {
      await updateDraft(editingID.value, payload)
      ElMessage.success('草稿已保存')
      return
    }
    const draft = await createDraft(payload)
    editingStatus.value = 0
    ElMessage.success('草稿已保存')
    router.replace(`/post/${draft.id}/edit`)
  } catch (error) {
    ElMessage.error(error.message)
  } finally {
    draftLoading.value = false
  }
}

onMounted(async () => {
  pageLoading.value = true
  try {
    await loadCommunities()
    await loadEditingPost()
  } catch (error) {
    ElMessage.error(error.message)
  } finally {
    pageLoading.value = false
  }
})
</script>

<template>
  <div class="forum-page">
    <ForumHeader />

    <main class="forum-layout">
      <section class="forum-main">
        <section class="editor-panel" v-loading="pageLoading">
          <header class="editor-panel__head">
            <div>
              <p class="topic-toolbar__eyebrow">{{ isEditMode ? '主题管理' : '新主题' }}</p>
              <h1>{{ pageTitle }}</h1>
              <p class="editor-panel__lead">{{ pageLead }}</p>
            </div>
            <div class="editor-panel__badge">
              <PenSquare :size="16" />
              <span>{{ forumConfig.editor.badge }}</span>
            </div>
          </header>

          <el-form ref="formRef" :model="form" :rules="rules" label-position="top" class="editor-form">
            <el-form-item label="节点" prop="community_id">
              <el-select v-model="form.community_id" placeholder="选择一个社区节点" size="large">
                <el-option
                  v-for="community in communities"
                  :key="community.id"
                  :label="community.name"
                  :value="community.id"
                />
              </el-select>
            </el-form-item>

            <el-form-item label="标题" prop="title">
              <el-input v-model.trim="form.title" maxlength="120" show-word-limit size="large" />
            </el-form-item>

            <el-form-item label="正文" prop="content">
              <el-input
                v-model="form.content"
                type="textarea"
                :rows="14"
                resize="none"
                placeholder="写下问题背景、上下文、日志、结论或经验。"
              />
            </el-form-item>

            <div class="editor-actions">
              <button class="site-action" type="button" @click="$router.push(isEditMode ? '/profile' : '/')">取消</button>
              <button
                v-if="!isEditMode || isDraftEdit"
                class="site-action"
                type="button"
                :disabled="draftLoading || loading"
                @click="saveDraft"
              >
                {{ draftLoading ? '保存中...' : '保存草稿' }}
              </button>
              <button class="site-action site-action--primary" type="button" :disabled="loading" @click="submit">
                {{ loading ? '处理中...' : isEditMode && !isDraftEdit ? '保存修改' : '发布主题' }}
              </button>
            </div>
          </el-form>
        </section>
      </section>

      <ForumSidebar :communities="communities" active-community-id="" @change-community="() => {}" />
    </main>
  </div>
</template>
