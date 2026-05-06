<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Leaf, ShieldPlus } from 'lucide-vue-next'
import ForumHeader from '../components/ForumHeader.vue'
import heroUrl from '../assets/hero.png'
import { signup } from '../api/auth'
import { forumConfig } from '../config/forum'

const router = useRouter()
const formRef = ref()
const loading = ref(false)
const errorMessage = ref('')

const form = reactive({
  username: '',
  password: '',
  re_password: ''
})

function confirmPassword(rule, value, callback) {
  if (!value) {
    callback(new Error('请再次输入密码'))
    return
  }
  if (value !== form.password) {
    callback(new Error('两次输入的密码不一致'))
    return
  }
  callback()
}

const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 2, max: 20, message: '用户名长度为 2-20 位', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度为 6-20 位', trigger: 'blur' }
  ],
  re_password: [{ validator: confirmPassword, trigger: 'blur' }]
}

async function submit() {
  if (!formRef.value) return
  errorMessage.value = ''
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await signup({ ...form })
    ElMessage.success('注册成功，请登录')
    router.push('/login')
  } catch (error) {
    errorMessage.value = error.message
    ElMessage.error(error.message)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="forum-page">
    <ForumHeader />

    <main class="auth-shell">
      <section class="auth-panel">
        <div class="auth-panel__head">
          <div>
            <p class="topic-toolbar__eyebrow">{{ forumConfig.auth.signupEyebrow }}</p>
            <h1>{{ forumConfig.auth.signupTitle }}</h1>
            <p>{{ forumConfig.auth.signupCopy }}</p>
          </div>
          <div class="auth-panel__badge">Create Account</div>
        </div>

        <el-alert
          v-if="errorMessage"
          :title="errorMessage"
          type="error"
          show-icon
          :closable="false"
          class="form-alert"
        />

        <el-form ref="formRef" :model="form" :rules="rules" label-position="top" class="auth-form">
          <el-form-item label="用户名" prop="username">
            <el-input v-model.trim="form.username" size="large" clearable />
          </el-form-item>
          <el-form-item label="密码" prop="password">
            <el-input v-model="form.password" size="large" type="password" show-password />
          </el-form-item>
          <el-form-item label="确认密码" prop="re_password">
            <el-input v-model="form.re_password" size="large" type="password" show-password @keyup.enter="submit" />
          </el-form-item>

          <div class="editor-actions">
            <button class="site-action" type="button" @click="$router.push('/login')">{{ forumConfig.auth.signupBackToLogin }}</button>
            <button class="site-action site-action--primary" type="button" :disabled="loading" @click="submit">
              {{ loading ? '注册中...' : '创建账号' }}
            </button>
          </div>
        </el-form>
      </section>

      <aside class="auth-aside">
        <img :src="heroUrl" :alt="forumConfig.brand.fullName" />
        <div class="auth-aside__content">
          <div class="auth-note">
            <ShieldPlus :size="16" />
            <span>{{ forumConfig.auth.signupHintPrimary }}</span>
          </div>
          <div class="auth-note">
            <Leaf :size="16" />
            <span>{{ forumConfig.auth.signupHintSecondary }}</span>
          </div>
        </div>
      </aside>
    </main>
  </div>
</template>
