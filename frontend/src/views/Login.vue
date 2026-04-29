<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { MessageCircle, Sparkles } from 'lucide-vue-next'
import { login } from '../api/auth'

const router = useRouter()
const formRef = ref()
const loading = ref(false)
const errorMessage = ref('')

const form = reactive({
  username: '',
  password: ''
})

const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 2, max: 20, message: '用户名长度为 2-20 位', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度为 6-20 位', trigger: 'blur' }
  ]
}

async function submit() {
  if (!formRef.value) return
  errorMessage.value = ''
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    const payload = await login({ ...form })
    if (payload?.token) {
      localStorage.setItem('bluebell_token', payload.token)
      localStorage.setItem('bluebell_user', JSON.stringify(payload))
      router.push('/')
      return
    }
    throw new Error('登录响应缺少 token')
  } catch (error) {
    errorMessage.value = error.message
    ElMessage.error(error.message)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main class="auth-page">
    <section class="auth-hero">
      <div class="sky-orb sky-orb-primary"></div>
      <div class="sky-orb sky-orb-soft"></div>
      <p class="eyebrow">Bluebell Forum</p>
      <h1>回到你的天蓝色讨论场</h1>
      <p class="hero-copy">登录后继续参与话题、收藏帖子，并和同好一起沉淀有价值的讨论。</p>
      <div class="hero-note">
        <MessageCircle :size="18" />
        <span>轻量、清爽、专注内容的社区前端入口。</span>
      </div>
    </section>

    <el-card class="auth-card" shadow="never">
      <template #header>
        <div class="card-header">
          <div>
            <p class="form-kicker">欢迎回来</p>
            <h2>登录账号</h2>
          </div>
          <Sparkles :size="24" />
        </div>
      </template>

      <el-alert
        v-if="errorMessage"
        :title="errorMessage"
        type="error"
        show-icon
        :closable="false"
        class="form-alert"
      />

      <el-form ref="formRef" :model="form" :rules="rules" label-position="top" size="large">
        <el-form-item label="用户名" prop="username">
          <el-input v-model.trim="form.username" placeholder="请输入用户名" clearable />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="form.password"
            placeholder="请输入密码"
            show-password
            type="password"
            @keyup.enter="submit"
          />
        </el-form-item>
        <el-button class="submit-button" type="primary" size="large" round :loading="loading" @click="submit">
          登录
        </el-button>
      </el-form>

      <p class="switch-copy">
        还没有账号？
        <RouterLink to="/signup">立即注册</RouterLink>
      </p>
    </el-card>
  </main>
</template>
