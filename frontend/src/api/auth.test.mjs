import assert from 'node:assert/strict'
import {
  AUTH_ERROR_CODES,
  buildAuthHeader,
  createAuthError,
  isAuthErrorCode,
  unwrapApiResponse
} from './auth.js'

assert.deepEqual(AUTH_ERROR_CODES, [1006, 1007])
assert.equal(buildAuthHeader('jwt-token'), 'Bearer jwt-token')
assert.equal(buildAuthHeader(''), '')
assert.equal(isAuthErrorCode(1006), true)
assert.equal(isAuthErrorCode(1007), true)
assert.equal(isAuthErrorCode(1005), false)

assert.deepEqual(
  unwrapApiResponse({ code: 1000, msg: 'success', data: { token: 'token-1' } }, 'fallback'),
  { token: 'token-1' }
)

assert.throws(
  () => unwrapApiResponse({ code: 1004, msg: '用户名或密码错误', data: null }, 'fallback'),
  /用户名或密码错误/
)

assert.throws(
  () => unwrapApiResponse({ code: 1005, msg: null, data: null }, '服务繁忙'),
  /服务繁忙/
)

assert.throws(
  () => unwrapApiResponse({ code: 1006, msg: '需要登录', data: null }, 'fallback'),
  (error) => error.message === '需要登录' && error.isAuthError === true
)

const authError = createAuthError('无效的token', 1007)
assert.equal(authError.message, '无效的token')
assert.equal(authError.code, 1007)
assert.equal(authError.isAuthError, true)
