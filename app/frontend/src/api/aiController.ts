import request from '@/request'

export async function submitAI(body: API.AISubmitRequest, options?: { [key: string]: any }) {
  return request<API.BaseResponseBoolean>('/app/ai/submit', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  })
}

export async function pushAI(body: API.AIControlRequest, options?: { [key: string]: any }) {
  return request<API.BaseResponseString>('/app/ai/push', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  })
}

export async function answerAI(body: API.AIControlRequest, options?: { [key: string]: any }) {
  return request<API.BaseResponseBoolean>('/app/ai/answer', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  })
}

export async function cancelAI(body: API.AIControlRequest, options?: { [key: string]: any }) {
  return request<API.BaseResponseString>('/app/ai/cancel', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  })
}

export async function getAIState(params: API.AIStateRequest, options?: { [key: string]: any }) {
  return request<API.BaseResponseAIState>('/app/ai/state', {
    method: 'GET',
    params,
    ...(options || {}),
  })
}

export async function listAIEvents(params: API.AIEventsRequest, options?: { [key: string]: any }) {
  return request<API.BaseResponseAIEvents>('/app/ai/events', {
    method: 'GET',
    params: {
      blockMs: 30000,
      count: 50,
      ...params,
    },
    ...(options || {}),
  })
}
