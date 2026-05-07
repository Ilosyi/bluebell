import axios from 'axios'
import { computed, readonly, ref } from 'vue'

export const AUTH_ERROR_CODES = [1006, 1007]
const TOKEN_KEY = 'bluebell_token'
const USER_KEY = 'bluebell_user'

export const client = axios.create({
  baseURL: '/api/v1',
  timeout: 8000
})

let authRedirectHandler = null

function getStorage() {
  return typeof localStorage === 'undefined' ? null : localStorage
}

function readTokenFromStorage() {
  return getStorage()?.getItem(TOKEN_KEY) || ''
}

function readUserFromStorage() {
  const raw = getStorage()?.getItem(USER_KEY)
  if (!raw) return null
  try {
    return JSON.parse(raw)
  } catch {
    return null
  }
}

function getMessage(error, fallback) {
  const response = error?.response?.data
  return response?.msg || response?.message || response?.data || error?.message || fallback
}

const authToken = ref(readTokenFromStorage())
const authUser = ref(readUserFromStorage())
const authLoggedIn = computed(() => Boolean(authToken.value))

function syncAuthStateFromStorage() {
  authToken.value = readTokenFromStorage()
  authUser.value = readUserFromStorage()
}

if (typeof window !== 'undefined') {
  window.addEventListener('storage', syncAuthStateFromStorage)
}

export function useAuthState() {
  return {
    token: readonly(authToken),
    user: readonly(authUser),
    isLoggedIn: readonly(authLoggedIn)
  }
}

export function getToken() {
  return authToken.value || ''
}

export function getUser() {
  return authUser.value
}

export function saveAuth(payload) {
  const storage = getStorage()
  const currentToken = authToken.value || readTokenFromStorage()
  const nextPayload = payload ? { ...authUser.value, ...payload } : null
  const nextToken = nextPayload ? nextPayload.token || currentToken : ''
  if (nextPayload && nextToken) {
    nextPayload.token = nextToken
  }
  authToken.value = nextToken || ''
  authUser.value = nextPayload

  if (!storage) return

  if (nextPayload?.token) {
    storage.setItem(TOKEN_KEY, nextPayload.token)
  } else {
    storage.removeItem(TOKEN_KEY)
  }

  if (nextPayload) {
    storage.setItem(USER_KEY, JSON.stringify(nextPayload))
  } else {
    storage.removeItem(USER_KEY)
  }
}

export function clearAuth() {
  authToken.value = ''
  authUser.value = null
  const storage = getStorage()
  if (!storage) return
  storage.removeItem(TOKEN_KEY)
  storage.removeItem(USER_KEY)
}

export function buildAuthHeader(token) {
  return token ? `Bearer ${token}` : ''
}

export function isAuthErrorCode(code) {
  return AUTH_ERROR_CODES.includes(code)
}

export function createAuthError(message, code) {
  const error = new Error(message)
  error.code = code
  error.isAuthError = true
  return error
}

export function setAuthRedirectHandler(handler) {
  authRedirectHandler = handler
}

function handleAuthFailure(payload) {
  clearAuth()
  if (authRedirectHandler) {
    authRedirectHandler(payload)
  }
  throw createAuthError(payload?.msg || '登录状态已失效，请重新登录', payload?.code)
}

export function unwrapApiResponse(response, fallback) {
  if (isAuthErrorCode(response?.code)) {
    throw createAuthError(response?.msg || '登录状态已失效，请重新登录', response?.code)
  }
  if (response?.code !== 1000) {
    throw new Error(response?.msg || fallback)
  }
  return response.data
}

client.interceptors.request.use((config) => {
  const authHeader = buildAuthHeader(getToken())
  config.headers = config.headers || {}
  if (authHeader) {
    config.headers.Authorization = authHeader
  } else {
    delete config.headers.Authorization
  }
  return config
})

client.interceptors.response.use(
  (response) => {
    const { data } = response
    if (isAuthErrorCode(data?.code)) {
      handleAuthFailure(data)
    }
    return response
  },
  (error) => {
    const response = error?.response?.data
    if (isAuthErrorCode(response?.code)) {
      handleAuthFailure(response)
    }
    throw error
  }
)

export async function login(payload) {
  try {
    const { data } = await client.post('/login', payload)
    return unwrapApiResponse(data, '登录失败，请稍后重试')
  } catch (error) {
    if (error?.isAuthError) throw error
    throw new Error(getMessage(error, '登录失败，请稍后重试'))
  }
}

export async function signup(payload) {
  try {
    const { data } = await client.post('/signup', payload)
    return unwrapApiResponse(data, '注册失败，请稍后重试')
  } catch (error) {
    if (error?.isAuthError) throw error
    throw new Error(getMessage(error, '注册失败，请稍后重试'))
  }
}

export async function fetchMe() {
  try {
    const { data } = await client.get('/me')
    const profile = unwrapApiResponse(data, '用户资料加载失败')
    saveAuth({
      ...profile,
      user_name: profile.username
    })
    return profile
  } catch (error) {
    if (error?.isAuthError) throw error
    throw new Error(getMessage(error, '用户资料加载失败'))
  }
}

export async function updateMe(payload) {
  try {
    const { data } = await client.put('/me', payload)
    const profile = unwrapApiResponse(data, '用户资料保存失败')
    saveAuth({
      ...profile,
      user_name: profile.username
    })
    return profile
  } catch (error) {
    if (error?.isAuthError) throw error
    throw new Error(getMessage(error, '用户资料保存失败'))
  }
}
