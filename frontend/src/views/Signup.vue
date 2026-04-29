<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Leaf, ShieldCheck } from 'lucide-vue-next'
import { signup } from '../api/auth'

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
  <main class="auth-page signup-page">
    <section class="auth-hero">
      <div class="sky-orb sky-orb-primary"></div>
      <div class="sky-orb sky-orb-soft"></div>
      <p class="eyebrow">Join Bluebell</p>
      <h1>创建一个清爽的论坛身份</h1>
      <p class="hero-copy">只保留必要信息，先让账号体系跑起来，之后再逐步扩展头像、简介和权限。</p>
      <div class="hero-note">
        <ShieldCheck :size="18" />
        <span>用户名 2-20 位，密码 6-20 位，与后端校验规则保持一致。</span>
      </div>
    </section>

    <el-card class="auth-card" shadow="never">
      <template #header>
        <div class="card-header">
          <div>
            <p class="form-kicker">新用户入口</p>
            <h2>注册账号</h2>
          </div>
          <Leaf :size="24" />
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
          <el-input v-model.trim="form.username" placeholder="2-20 位用户名" clearable />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="form.password" placeholder="6-20 位密码" show-password type="password" />
        </el-form-item>
        <el-form-item label="确认密码" prop="re_password">
          <el-input
            v-model="form.re_password"
            placeholder="再次输入密码"
            show-password
            type="password"
            @keyup.enter="submit"
          />
        </el-form-item>
        <el-button class="submit-button" type="primary" size="large" round :loading="loading" @click="submit">
          创建账号
        </el-button>
      </el-form>

      <p class="switch-copy">
        已有账号？
        <RouterLink to="/login">返回登录</RouterLink>
      </p>
    </el-card>
  </main>
</template>
