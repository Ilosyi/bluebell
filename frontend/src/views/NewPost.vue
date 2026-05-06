<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { PenSquare } from 'lucide-vue-next'
import ForumHeader from '../components/ForumHeader.vue'
import ForumSidebar from '../components/ForumSidebar.vue'
import { createPost, fetchCommunities } from '../api/forum'
import { forumConfig } from '../config/forum'

const router = useRouter()
const formRef = ref()
const loading = ref(false)
const communities = ref([])

const form = reactive({
  title: '',
  content: '',
  community_id: ''
})

const rules = {
  community_id: [{ required: true, message: '请选择节点', trigger: 'change' }],
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
  content: [{ required: true, message: '请输入正文', trigger: 'blur' }]
}

async function loadCommunities() {
  communities.value = await fetchCommunities()
}

async function submit() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await createPost({
      title: form.title,
      content: form.content,
      community_id: Number(form.community_id)
    })
    ElMessage.success('主题已发布')
    router.push('/')
  } catch (error) {
    ElMessage.error(error.message)
  } finally {
    loading.value = false
  }
}

onMounted(loadCommunities)
</script>

<template>
  <div class="forum-page">
    <ForumHeader />

    <main class="forum-layout">
      <section class="forum-main">
        <section class="editor-panel">
          <header class="editor-panel__head">
            <div>
              <p class="topic-toolbar__eyebrow">新主题</p>
              <h1>发布新主题</h1>
              <p class="editor-panel__lead">写下你的问题、经验或想法。</p>
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
              <button class="site-action" type="button" @click="$router.push('/')">取消</button>
              <button class="site-action site-action--primary" type="button" :disabled="loading" @click="submit">
                {{ loading ? '发布中...' : '发布主题' }}
              </button>
            </div>
          </el-form>
        </section>
      </section>

      <ForumSidebar :communities="communities" active-community-id="" @change-community="() => {}" />
    </main>
  </div>
</template>
