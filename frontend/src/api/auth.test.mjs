import assert from 'node:assert/strict'
import { unwrapApiResponse } from './auth.js'

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
