import axios from 'axios'

const client = axios.create({
  baseURL: '/api',
  timeout: 8000
})

function getMessage(error, fallback) {
  const response = error?.response?.data
  return response?.msg || response?.message || response?.data || error?.message || fallback
}

export function unwrapApiResponse(response, fallback) {
  if (response?.code !== 1000) {
    throw new Error(response?.msg || fallback)
  }
  return response.data
}

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
