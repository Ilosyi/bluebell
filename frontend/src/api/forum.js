import { client, unwrapApiResponse } from './auth'

function unwrap(data, fallback) {
  return unwrapApiResponse(data, fallback)
}

export function normalizePost(item) {
  const post = item?.post || item || {}
  const community = item?.community || item?.CommunityDetail || post.community || {}
  const rawID = post.id ?? post.post_id ?? item?.id
  const id = rawID == null ? '' : String(rawID)

  return {
    id,
    title: post.title || '未命名主题',
    content: post.content || '',
    authorID: post.author_id ?? post.AuthorID ?? '',
    authorName: item?.author_name || item?.AuthorName || post.author_name || '匿名用户',
    communityID: post.community_id ?? community.id ?? '',
    communityName: community.name || post.community_name || '未分类',
    communityIntro: community.introduction || '',
    voteNum: Number(item?.vote_num ?? item?.VoteNum ?? post.vote_num ?? 0),
    status: post.status ?? 0,
    createTime: post.create_time || item?.create_time || '',
    raw: item
  }
}

export function normalizeCommunity(item) {
  return {
    id: item?.id == null ? '' : String(item.id),
    name: item?.name || '未命名节点',
    introduction: item?.introduction || '',
    createTime: item?.create_time || ''
  }
}

export function formatPageParams(params = {}) {
  return {
    page: Number(params.page) > 0 ? Number(params.page) : 1,
    size: Number(params.size) > 0 ? Number(params.size) : 10,
    order: params.order || 'time',
    community_id: params.communityID || undefined
  }
}

export async function fetchCommunities() {
  const { data } = await client.get('/community')
  return unwrap(data, '社区列表加载失败').map(normalizeCommunity)
}

export async function fetchCommunityDetail(id) {
  const { data } = await client.get(`/community/${id}`)
  return normalizeCommunity(unwrap(data, '社区详情加载失败'))
}

export async function fetchPosts(params = {}) {
  const { data } = await client.get('/posts2', {
    params: formatPageParams(params)
  })
  const payload = unwrap(data, '帖子列表加载失败')
  return {
    items: (payload.items || []).map(normalizePost),
    pagination: payload.pagination || {
      page: 1,
      size: params.size || 20,
      total: 0,
      totalPages: 1,
      hasMore: false
    }
  }
}

export async function fetchPostDetail(id) {
  const { data } = await client.get(`/post/${id}`)
  return normalizePost(unwrap(data, '帖子详情加载失败'))
}

export async function createPost(payload) {
  const { data } = await client.post('/post', payload)
  return unwrap(data, '发帖失败')
}

export async function votePost(postID, direction) {
  const { data } = await client.post('/vote', {
    post_id: String(postID),
    direction: String(direction)
  })
  return unwrap(data, '投票失败')
}
