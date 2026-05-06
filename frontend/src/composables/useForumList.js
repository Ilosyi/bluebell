import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { fetchCommunities, fetchCommunityDetail, fetchPosts } from '../api/forum'

const PAGE_SIZE = 20

export function useForumList() {
  const route = useRoute()
  const router = useRouter()

  const loading = ref(false)
  const communities = ref([])
  const posts = ref([])
  const activeOrder = ref('time')
  const activeCommunityID = ref('')
  const activeCommunity = ref(null)
  const errorMessage = ref('')
  const page = ref(1)
  const total = ref(0)
  const totalPages = ref(1)
  const hasMore = ref(false)
  const initialized = ref(false)

  const sortTabs = [
    { label: '最新', value: 'time' },
    { label: '最热', value: 'score' }
  ]

  const activeCommunityLabel = computed(() => {
    if (!activeCommunityID.value) return '全部主题'
    return communities.value.find((item) => item.id === activeCommunityID.value)?.name || '当前节点'
  })

  const pageSummary = computed(() => `第 ${page.value} 页`)
  const hasPreviousPage = computed(() => page.value > 1)
  const hotTopics = computed(() => posts.value.slice(0, 4))

  const visiblePages = computed(() => {
    const totalValue = totalPages.value || 1
    const current = page.value
    if (totalValue <= 10) {
      return Array.from({ length: totalValue }, (_, index) => index + 1)
    }

    const pages = [1]
    const start = Math.max(2, current - 3)
    const end = Math.min(totalValue - 1, current + 3)

    if (start > 2) pages.push('ellipsis-left')
    for (let value = start; value <= end; value += 1) {
      pages.push(value)
    }
    if (end < totalValue - 1) pages.push('ellipsis-right')
    pages.push(totalValue)

    return pages
  })

  async function loadCommunities() {
    communities.value = await fetchCommunities()
  }

  async function loadActiveCommunity() {
    if (!activeCommunityID.value) {
      activeCommunity.value = null
      return
    }
    activeCommunity.value = await fetchCommunityDetail(activeCommunityID.value)
  }

  async function loadPosts() {
    loading.value = true
    errorMessage.value = ''
    try {
      const result = await fetchPosts({
        order: activeOrder.value,
        communityID: activeCommunityID.value,
        page: page.value,
        size: PAGE_SIZE
      })
      posts.value = result.items
      total.value = result.pagination.total
      totalPages.value = result.pagination.totalPages || 1
      hasMore.value = result.pagination.hasMore
    } catch (error) {
      posts.value = []
      total.value = 0
      totalPages.value = 1
      hasMore.value = false
      errorMessage.value = error.message
    } finally {
      loading.value = false
    }
  }

  function buildQuery() {
    return {
      ...(activeCommunityID.value ? { community: activeCommunityID.value } : {}),
      ...(activeOrder.value !== 'time' ? { order: activeOrder.value } : {}),
      ...(page.value > 1 ? { page: String(page.value) } : {})
    }
  }

  function updateQuery() {
    router.replace({
      path: '/',
      query: buildQuery()
    })
  }

  function changeOrder(order) {
    activeOrder.value = order
    page.value = 1
  }

  function changeCommunity(id) {
    activeCommunityID.value = id
    page.value = 1
  }

  function goPage(nextPage) {
    if (nextPage < 1 || nextPage > totalPages.value || nextPage === page.value) return
    page.value = nextPage
  }

  async function refreshAll() {
    try {
      await Promise.all([loadCommunities(), loadPosts(), loadActiveCommunity()])
    } catch (error) {
      ElMessage.error(error.message || '页面刷新失败')
    }
  }

  watch([activeOrder, activeCommunityID, page], () => {
    if (!initialized.value) return
    updateQuery()
    loadPosts()
  })

  watch(activeCommunityID, () => {
    if (!initialized.value) return
    loadActiveCommunity().catch((error) => {
      activeCommunity.value = null
      ElMessage.error(error.message || '社区详情加载失败')
    })
  })

  async function initialize() {
    if (typeof route.query.community === 'string') {
      activeCommunityID.value = route.query.community
    }
    if (typeof route.query.order === 'string' && ['time', 'score'].includes(route.query.order)) {
      activeOrder.value = route.query.order
    }
    if (typeof route.query.page === 'string') {
      const parsed = Number(route.query.page)
      page.value = parsed > 0 ? parsed : 1
    }
    await refreshAll()
    initialized.value = true
  }

  return {
    PAGE_SIZE,
    loading,
    communities,
    posts,
    activeOrder,
    activeCommunityID,
    activeCommunity,
    errorMessage,
    page,
    total,
    totalPages,
    hasMore,
    sortTabs,
    activeCommunityLabel,
    pageSummary,
    hasPreviousPage,
    hotTopics,
    visiblePages,
    changeOrder,
    changeCommunity,
    goPage,
    refreshAll,
    initialize
  }
}
