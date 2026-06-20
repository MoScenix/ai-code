import axios from 'axios'

export const getResponseErrorMessage = (data: unknown, fallback = '请求失败') => {
  if (typeof data === 'string' && data.trim()) {
    return data.trim()
  }
  if (data && typeof data === 'object') {
    const body = data as Record<string, unknown>
    for (const key of ['message', 'error', 'msg']) {
      const value = body[key]
      if (typeof value === 'string' && value.trim()) {
        return value.trim()
      }
    }
  }
  return fallback
}

export const getRequestErrorMessage = (error: unknown, fallback = '请求失败') => {
  if (axios.isAxiosError(error)) {
    const responseMessage = getResponseErrorMessage(error.response?.data, '')
    if (responseMessage) {
      return responseMessage
    }
    if (error.message) {
      return error.message
    }
  }
  if (error instanceof Error && error.message) {
    return error.message
  }
  if (typeof error === 'string' && error.trim()) {
    return error.trim()
  }
  return fallback
}
