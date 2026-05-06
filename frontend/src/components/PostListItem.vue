<script setup>
import { computed } from 'vue'
import { ArrowBigUp, Clock3, UserRound } from 'lucide-vue-next'

const props = defineProps({
  post: {
    type: Object,
    required: true
  }
})

const summary = computed(() => {
  const raw = props.post?.content || ''
  if (raw.length <= 96) return raw
  return `${raw.slice(0, 96)}...`
})

const createdAt = computed(() => {
  if (!props.post?.createTime) return '刚刚'
  const date = new Date(props.post.createTime)
  if (Number.isNaN(date.getTime())) return '刚刚'
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(
    date.getDate()
  ).padStart(2, '0')}`
})
</script>

<template>
  <RouterLink class="topic-row" :to="`/post/${post.id}`">
    <div class="topic-row__main">
      <div class="topic-row__meta">
        <span class="topic-chip">{{ post.communityName }}</span>
        <span class="topic-row__author">
          <UserRound :size="12" />
          {{ post.authorName }}
        </span>
        <span class="topic-row__meta-sep">/</span>
        <span class="topic-row__date">
          <Clock3 :size="12" />
          {{ createdAt }}
        </span>
      </div>

      <h3>{{ post.title }}</h3>
      <p>{{ summary }}</p>
    </div>

    <div class="topic-row__score">
      <ArrowBigUp :size="16" />
      <strong>{{ post.voteNum }}</strong>
    </div>
  </RouterLink>
</template>
