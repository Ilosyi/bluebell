<script setup>
import { LogIn, PenLine } from 'lucide-vue-next'
</script>

<template>
  <main class="home-page">
    <el-card class="home-card" shadow="never">
      <div class="brand-mark">Bluebell</div>
      <h1>论坛首页正在搭建</h1>
      <p>当前阶段先完成账号入口。后续可以在这里接入帖子流、分区导航和个人动态。</p>

      <div class="home-actions">
        <el-button type="primary" size="large" round @click="$router.push('/login')">
          <LogIn :size="18" />
          去登录
        </el-button>
        <el-button size="large" round @click="$router.push('/signup')">
          <PenLine :size="18" />
          创建账号
        </el-button>
      </div>
    </el-card>
  </main>
</template>
