import axios from 'axios'

export const AUTH_ERROR_CODES = [1006, 1007]
const TOKEN_KEY = 'bluebell_token'
const USER_KEY = 'bluebell_user'

export const client = axios.create({
  baseURL: '/api/v1',
  timeout: 8000
})

let authRedirectHandler = null

function getMessage(error, fallback) {
  const response = error?.response?.data
  return response?.msg || response?.message || response?.data || error?.message || fallback
}

export function getToken() {
  return localStorage.getItem(TOKEN_KEY) || ''
}

export function getUser() {
  const raw = localStorage.getItem(USER_KEY)
  if (!raw) return null
  try {
    return JSON.parse(raw)
  } catch {
    return null
  }
}

export function saveAuth(payload) {
  if (payload?.token) {
    localStorage.setItem(TOKEN_KEY, payload.token)
  }
  localStorage.setItem(USER_KEY, JSON.stringify(payload))
}

export function clearAuth() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
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
  const token = getToken()
  const authHeader = buildAuthHeader(token)
  if (authHeader) {
    config.headers.Authorization = authHeader
  }
  return config
})

client.interceptors.response.use(
  (response) => {
    const { data } = response
    if (isAuthErrorCode(data?.code)) {
      clearAuth()
      if (authRedirectHandler) {
        authRedirectHandler(data)
      }
      throw createAuthError(data?.msg || '登录状态已失效，请重新登录', data?.code)
    }
    return response
  },
  (error) => {
    const response = error?.response?.data
    if (isAuthErrorCode(response?.code)) {
      clearAuth()
      if (authRedirectHandler) {
        authRedirectHandler(response)
      }
      throw createAuthError(response?.msg || '登录状态已失效，请重新登录', response?.code)
    }
    throw error
  }
)

export async function login(payload) {
  try {
    const { data } = await client.post('/login', payload)
    return unwrapApiResponse(data, '登录失败，请稍后重试')
  } catch (error) {
    throw new Error(getMessage(error, '登录失败，请稍后重试'))
  }
}

export async function signup(payload) {
  try {
    const { data } = await client.post('/signup', payload)
    return unwrapApiResponse(data, '注册失败，请稍后重试')
  } catch (error) {
    throw new Error(getMessage(error, '注册失败，请稍后重试'))
  }
}
