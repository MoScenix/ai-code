<template>
  <div class="h-screen w-full flex flex-col bg-slate-50 text-slate-900 overflow-hidden">
    <header
      class="h-16 flex items-center justify-end px-6 bg-white/80 backdrop-blur-md border-b border-slate-200 z-10">
      <div class="flex items-center gap-2">
        <a-button type="text" @click="showAppDetail" class="!flex !items-center hover:!bg-slate-100 !rounded-full">
          <template #icon>
            <InfoCircleOutlined />
          </template>
          应用详情
        </a-button>
        <div class="h-4 w-[1px] bg-slate-200 mx-2"></div>
        <a-button type="default" @click="downloadCode" :loading="downloading" :disabled="!isOwner"
          class="!rounded-full !border-slate-200 hover:!border-blue-500 hover:!text-blue-500">
          <template #icon>
            <DownloadOutlined />
          </template>
          下载代码
        </a-button>
        <a-button type="primary" @click="deployApp" :loading="deploying"
          class="!rounded-full !bg-blue-600 shadow-md shadow-blue-100 hover:!scale-105 transition-transform">
          <template #icon>
            <CloudUploadOutlined />
          </template>
          部署发布
        </a-button>
      </div>
    </header>

    <main class="flex-1 flex overflow-hidden p-4 gap-4">

      <section
        class="chat-shell flex-1 min-w-[400px] flex flex-col bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden">
        <div ref="messagesContainer" class="chat-list flex-1 overflow-y-auto scroll-smooth custom-scrollbar">
          <div v-if="hasMoreHistory" class="flex justify-center">
            <a-button type="link" @click="loadMoreHistory" :loading="loadingHistory" size="small"
              class="text-slate-400 font-normal">
              查看更早的历史消息
            </a-button>
          </div>

          <div v-if="!messages.length" class="empty-chat">
            <div class="empty-title">开始一次修改</div>
            <div class="empty-subtitle">描述你想改的内容，AI 会继续处理当前应用。</div>
          </div>

          <template v-for="(item, index) in messages" :key="item.id || index">
            <div v-if="item.type === 'system'" class="system-row">
              <span>{{ item.content }}</span>
            </div>

            <div v-else-if="item.type === 'user'" class="message-row user-row">
              <div class="user-message">
                <div v-if="item.isPush" class="push-label">push</div>
                <div
                  v-if="item.isFile && item.fileMeta"
                  class="file-message"
                  @dblclick="openFileMessage(item.fileMeta)"
                >
                  <FileTextOutlined class="file-icon" />
                  <div class="file-main">
                    <div class="file-name">{{ item.fileMeta.filename || '未命名文件' }}</div>
                    <div class="file-meta">
                      <span>{{ formatFileSize(item.fileMeta.size) }}</span>
                      <span>{{ item.fileMeta.isBig ? '已分块' : '已解析' }}</span>
                      <span v-if="item.fileMeta.parentCount">{{ item.fileMeta.parentCount }} 个父块</span>
                    </div>
                  </div>
                </div>
                <MarkdownRenderer v-else class="message-markdown user-markdown" :content="item.content.trimEnd()" />
              </div>
            </div>

            <div v-else class="message-row assistant-row">
              <div class="assistant-message">
                <div v-if="item.agent" class="agent-label">{{ item.agent }}</div>
                <template v-if="item.parts?.length">
                  <template v-for="part in item.parts" :key="part.id">
                    <MarkdownRenderer
                      v-if="part.type === 'text' && part.content"
                      class="message-markdown assistant-markdown"
                      :content="part.content.trimEnd()"
                    />
                    <div v-else-if="part.type === 'tool'" class="tool-list">
                      <details class="tool-item">
                        <summary>
                          <span class="tool-corner"></span>
                          <span class="tool-status">{{ part.tool.status === 'running' ? 'running' : part.tool.status === 'error' ? 'failed' : 'ran' }}</span>
                          <span class="tool-name">{{ part.tool.name }}</span>
                        </summary>
                        <pre v-if="part.tool.args" class="tool-code">{{ formatJSON(part.tool.args) }}</pre>
                        <pre v-if="part.tool.result" class="tool-code">{{ part.tool.result }}</pre>
                      </details>
                    </div>
                  </template>
                </template>
                <template v-else>
                  <MarkdownRenderer v-if="item.content" class="message-markdown assistant-markdown" :content="item.content.trimEnd()" />
                  <div v-if="item.toolCalls?.length" class="tool-list">
                    <details v-for="tool in item.toolCalls" :key="tool.id" class="tool-item">
                      <summary>
                        <span class="tool-corner"></span>
                        <span class="tool-status">{{ tool.status === 'running' ? 'running' : tool.status === 'error' ? 'failed' : 'ran' }}</span>
                        <span class="tool-name">{{ tool.name }}</span>
                      </summary>
                      <pre v-if="tool.args" class="tool-code">{{ formatJSON(tool.args) }}</pre>
                      <pre v-if="tool.result" class="tool-code">{{ tool.result }}</pre>
                    </details>
                  </div>
                </template>
                <div v-if="item.loading" class="assistant-loading">
                  <span class="loading-dot"></span>
                  <span>{{ aiState?.status || 'unknown' }}</span>
                </div>
              </div>
            </div>
          </template>
        </div>

        <div class="composer-wrap">
          <div v-if="currentQuestion && currentQuestionItem" class="question-panel">
            <div class="question-title">
              <span>{{ currentQuestion.agent || 'AI' }} 需要确认</span>
              <span v-if="currentQuestionItems.length > 1" class="question-count">
                {{ currentQuestionIndex + 1 }} / {{ currentQuestionItems.length }}
              </span>
            </div>
            <div class="question-item">
              <div class="question-content">{{ currentQuestionItem.question }}</div>
              <div v-if="currentQuestionItem.options.length" class="question-options">
                <button
                  v-for="option in currentQuestionItem.options"
                  :key="option"
                  type="button"
                  class="question-option"
                  :class="{ 'question-option-active': currentAnswerSelection === option }"
                  @click="selectAnswerOption(option)"
                >
                  {{ option }}
                </button>
              </div>
            </div>
            <a-textarea
              v-model:value="answerInput"
              :rows="2"
              placeholder="输入其他回答，Enter 继续"
              class="!rounded-xl !text-sm !border-slate-200"
              @keydown.enter.prevent="submitQuestionStep"
            />
            <div class="flex justify-end gap-2 mt-2">
              <a-button size="small" @click="currentQuestion = null">稍后</a-button>
              <a-button
                type="primary"
                size="small"
                :disabled="!canSubmitAnswer || answeringQuestion"
                :loading="answeringQuestion"
                @click="submitQuestionStep"
              >
                {{ isLastQuestion ? '发送' : '继续' }}
              </a-button>
            </div>
          </div>

          <div v-if="!currentQuestion && selectedElementInfo" class="mb-3">
            <div
              class="flex items-center justify-between bg-amber-50 border border-amber-100 rounded-lg px-3 py-2 animate-in zoom-in-95">
              <div class="flex items-center gap-2 overflow-hidden">
                <span class="px-1.5 py-0.5 bg-amber-200 text-amber-800 rounded text-[10px] font-bold">选中元素</span>
                <span class="text-xs text-amber-900 truncate font-mono">{{ selectedElementInfo.tagName.toLowerCase()
                }}{{
                    selectedElementInfo.id ? '#' + selectedElementInfo.id : '' }}</span>
              </div>
              <button @click="clearSelectedElement" class="text-amber-400 hover:text-amber-600 transition-colors">
                <CloseCircleFilled />
              </button>
            </div>
          </div>

          <PromptInputBox
            v-if="!currentQuestion"
            v-model="userInput"
            :is-loading="isGenerating"
            :is-submitting="sendingMessage"
            :disabled="!isOwner && !isAdmin"
            :placeholder="getInputPlaceholder()"
            @send="handleSendMessage"
            @cancel="cancelCurrentTask"
          />
          <div v-if="!currentQuestion" class="composer-hint">
            <span v-if="isOwner">Enter 发送，Shift+Enter 换行；运行中输入内容会追加给当前任务</span>
            <span v-else class="text-amber-500">
              <LockOutlined /> 访客模式不可编辑
            </span>
          </div>
        </div>
      </section>

      <section
        class="flex-[1.5] flex flex-col bg-white rounded-2xl shadow-xl shadow-slate-200/50 border border-slate-200 overflow-hidden relative group">
        <div class="h-11 bg-slate-50 flex items-center px-4 justify-between border-b border-slate-200">
          <div class="flex items-center gap-2 w-1/4">
            <div class="w-3 h-3 rounded-full bg-slate-200 group-hover:bg-[#FF5F57] transition-colors"></div>
            <div class="w-3 h-3 rounded-full bg-slate-200 group-hover:bg-[#FFBD2E] transition-colors"></div>
            <div class="w-3 h-3 rounded-full bg-slate-200 group-hover:bg-[#28C840] transition-colors"></div>
          </div>

          <div class="flex-1 flex items-center justify-center">
            <span class="px-3 py-1 rounded-full text-[12px] font-semibold border shadow-sm" :class="previewUrl
              ? 'bg-emerald-50 text-emerald-700 border-emerald-200'
              : 'bg-rose-50 text-rose-700 border-rose-200'">
              {{ previewUrl ? '预览成功' : '预览失败' }}
            </span>
          </div>

          <ReloadOutlined class="text-[10px] text-slate-400 cursor-pointer hover:text-blue-500 transition-colors"
            @click="updatePreview(true)" />



          <div class="w-1/4 flex justify-end gap-2">
            <a-button v-if="isOwner && previewUrl" type="text" size="small" @click="toggleEditMode"
              class="!text-xs !flex !items-center !gap-1 !rounded-md"
              :class="isEditMode ? '!text-blue-600 !bg-blue-50' : '!text-slate-500 hover:!bg-slate-100'">
              <template #icon>
                <EditOutlined />
              </template>
              {{ isEditMode ? '停止选择' : '点击选择' }}
            </a-button>

            <div class="w-[1px] h-4 bg-slate-200 mx-1 self-center"></div>

            <a-button type="text" size="small" @click="openInNewTab"
              class="!text-slate-500 hover:!bg-slate-100 !flex !items-center !justify-center">
              <template #icon>
                <ExportOutlined />
              </template>
            </a-button>
          </div>
        </div>

        <div class="flex-1 bg-white relative">
          <div v-if="!previewUrl && !isGenerating"
            class="absolute inset-0 flex flex-col items-center justify-center text-slate-300 gap-4">
            <div
              class="w-16 h-16 rounded-3xl bg-slate-50 flex items-center justify-center border border-slate-100 shadow-inner">
              <GlobalOutlined class="text-3xl" />
            </div>
            <div class="text-center">
              <p class="text-sm font-semibold text-slate-400">等待页面生成...</p>
              <p class="text-xs text-slate-300 mt-1">在左侧描述你的需求，AI 将实时渲染页面</p>
            </div>
          </div>

          <div v-else-if="isGenerating && !previewUrl"
            class="absolute inset-0 flex flex-col items-center justify-center bg-white/80 backdrop-blur-sm z-20">
            <a-spin size="large" />
            <p class="mt-4 text-slate-500 text-sm font-medium animate-pulse">正在构建页面代码...</p>
          </div>

          <iframe ref="previewIframe" v-show="previewUrl" :src="previewUrl"
            class="w-full h-full border-none shadow-inner" @load="onIframeLoad" />


          <div v-if="isEditMode"
            class="absolute bottom-6 left-1/2 -translate-x-1/2 px-5 py-2.5 bg-blue-600 text-white rounded-full text-xs font-bold shadow-2xl shadow-blue-500/40 animate-bounce pointer-events-none z-30 flex items-center gap-2">
            <span class="w-2 h-2 bg-white rounded-full animate-ping"></span>
            请在上方点击你想修改的网页元素
          </div>
        </div>
      </section>
    </main>

    <AppDetailModal v-model:open="appDetailVisible" :app="appInfo" :show-actions="isOwner || isAdmin" @edit="editApp"
      @delete="deleteApp" />
    <DeploySuccessModal v-model:open="deployModalVisible" :deploy-url="deployUrl" @open-site="openDeployedSite" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { useLoginUserStore } from '@/stores/loginUser'
import { getAppVoById, deployApp as deployAppApi, deleteApp as deleteAppApi } from '@/api/appController'
import { listAppChatHistory } from '@/api/chatHistoryController'
import request from '@/request'

import MarkdownRenderer from '@/components/MarkdownRenderer.vue'
import PromptInputBox from '@/components/PromptInputBox.vue'
import AppDetailModal from '@/components/AppDetailModal.vue'
import DeploySuccessModal from '@/components/DeploySuccessModal.vue'
import { STATIC_BASE_URL, getStaticPreviewUrl } from '@/config/env'
import { VisualEditor, type ElementInfo } from '@/utils/visualEditor'
import { getRequestErrorMessage, getResponseErrorMessage } from '@/utils/requestError'
import { useAIEvents, type AIFileMeta, type AIMessage } from '@/composables/useAIEvents'

import {
  CloudUploadOutlined,
  CloseCircleFilled,
  ExportOutlined,
  InfoCircleOutlined,
  DownloadOutlined,
  EditOutlined,
  FileTextOutlined,
  GlobalOutlined,
  LockOutlined,
  ReloadOutlined,
} from '@ant-design/icons-vue'

const route = useRoute()
const router = useRouter()
const loginUserStore = useLoginUserStore()

// 应用信息
const appInfo = ref<API.AppVO>()
const appId = ref<any>()

const userInput = ref('')
const answerInput = ref('')
const answerSelections = ref<Record<number, string>>({})
const currentQuestionIndex = ref(0)
const answeringQuestion = ref(false)
const sendingMessage = ref(false)
const messagesContainer = ref<HTMLElement>()

const {
  messages,
  aiState,
  isGenerating,
  currentQuestion,
  sendMessage,
  pushMessage,
  answerQuestion,
  cancelCurrentTask,
  loadInitialState,
  stop: stopAI,
} = useAIEvents(appId)

// 对话历史相关
const loadingHistory = ref(false)
const hasMoreHistory = ref(false)
const lastCreateTime = ref<string>()
const historyLoaded = ref(false)

// 预览相关
const previewUrl = ref('')
const previewReady = ref(false)

// 部署相关
const deploying = ref(false)
const deployModalVisible = ref(false)
const deployUrl = ref('')

// 下载相关
const downloading = ref(false)

// 可视化编辑相关
const isEditMode = ref(false)
const selectedElementInfo = ref<ElementInfo | null>(null)
const visualEditor = new VisualEditor({
  onElementSelected: (elementInfo: ElementInfo) => {
    selectedElementInfo.value = elementInfo
  },
})

// 权限相关
const isOwner = computed(() => {
  return appInfo.value?.userId === loginUserStore.loginUser.id
})

const isAdmin = computed(() => {
  return loginUserStore.loginUser.userRole === 'admin'
})

const currentQuestionItems = computed(() => {
  if (!currentQuestion.value) return []
  return currentQuestion.value.questions.length
    ? currentQuestion.value.questions
    : [{ question: currentQuestion.value.content, options: [] }]
})

const currentQuestionItem = computed(() => currentQuestionItems.value[currentQuestionIndex.value])

const currentAnswerSelection = computed(() => {
  return answerSelections.value[currentQuestionIndex.value] || ''
})

const isLastQuestion = computed(() => {
  return currentQuestionIndex.value >= currentQuestionItems.value.length - 1
})

const canSubmitAnswer = computed(() => {
  return Boolean(answerInput.value.trim() || currentAnswerSelection.value.trim())
})

// 应用详情相关
const appDetailVisible = ref(false)

// 显示应用详情
const showAppDetail = () => {
  appDetailVisible.value = true
}

const parseFileMeta = (content?: string): AIFileMeta | undefined => {
  if (!content) return undefined
  try {
    const meta = JSON.parse(content) as AIFileMeta
    return meta && typeof meta === 'object' ? meta : undefined
  } catch {
    return undefined
  }
}

const buildHistoryMessage = (chat: API.ChatHistory): AIMessage => {
  const fileMeta = chat.isFile ? parseFileMeta(chat.message) : undefined
  return {
    id: chat.id?.toString() || `${chat.messageType}-${chat.createTime || Math.random()}`,
    type: (chat.messageType === 'user' ? 'user' : 'ai') as 'user' | 'ai',
    content: fileMeta ? fileMeta.filename || '' : chat.message || '',
    createTime: chat.createTime,
    isFile: Boolean(chat.isFile && fileMeta),
    fileMeta,
  }
}

const openFileMessage = (fileMeta: AIFileMeta) => {
  if (!appId.value || !fileMeta.fileId || !fileMeta.filename) {
    message.warning('文件地址不存在')
    return
  }
  const base = STATIC_BASE_URL.replace(/\/$/, '')
  const filename = fileMeta.filename.split('/').map(encodeURIComponent).join('/')
  window.open(`${base}/document/${appId.value}/${fileMeta.fileId}/${filename}`, '_blank')
}

const formatFileSize = (size?: number) => {
  if (!size || size <= 0) return '未知大小'
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / 1024 / 1024).toFixed(1)} MB`
}

// 加载对话历史
const loadChatHistory = async (isLoadMore = false) => {
  if (!appId.value || loadingHistory.value) return
  loadingHistory.value = true
  try {
    const params: API.listAppChatHistoryParams = {
      appId: appId.value,
      pageSize: 10,
    }
    // 如果是加载更多，传递最后一条消息的创建时间作为游标
    if (isLoadMore && lastCreateTime.value) {
      params.lastCreateTime = lastCreateTime.value
    }
    const res = await listAppChatHistory(params)
    console.log('listAppChatHistory raw res =', res)
    if (res.data.code === 0 && res.data.data) {
      const chatHistories = res.data.data.records || []
      if (chatHistories.length > 0) {
        // 将对话历史转换为消息格式，并按时间正序排列（老消息在前）
        const historyMessages: AIMessage[] = chatHistories
          .map(buildHistoryMessage)
          .reverse() // 反转数组，让老消息在前
        if (isLoadMore) {
          // 加载更多时，将历史消息添加到开头
          messages.value.unshift(...historyMessages)
        } else {
          // 初始加载，直接设置消息列表
          messages.value = historyMessages
        }
        // 更新游标
        lastCreateTime.value = chatHistories[chatHistories.length - 1]?.createTime
        // 检查是否还有更多历史
        hasMoreHistory.value = chatHistories.length === 10
      } else {
        hasMoreHistory.value = false
      }
      historyLoaded.value = true
    }
  } catch (error) {
    console.error('加载对话历史失败：', error)
    message.error('加载对话历史失败')
  } finally {
    loadingHistory.value = false
  }
}

// 加载更多历史消息
const loadMoreHistory = async () => {
  await loadChatHistory(true)
}

// 获取应用信息
const fetchAppInfo = async () => {
  const id = route.params.id as string
  if (!id) {
    message.error('应用ID不存在')
    router.push('/')
    return
  }

  appId.value = id

  try {
    const res = await getAppVoById({ id: id as unknown as number })
    if (res.data.code === 0 && res.data.data) {
      appInfo.value = res.data.data

      // 先加载对话历史
      await loadChatHistory()
      // 如果有至少2条对话记录，展示对应的网站
      if (messages.value.length >= 2) {
        updatePreview()
      }
      // 检查是否需要自动发送初始提示词
      // 只有在是自己的应用且没有对话历史时才自动发送
      const hasNonFileHistory = messages.value.some((item) => !item.isFile)
      if (appInfo.value.initPrompt && isOwner.value && !hasNonFileHistory && historyLoaded.value) {
        await sendInitialMessage(appInfo.value.initPrompt)
      }
    } else {
      message.error('获取应用信息失败')
      router.push('/')
    }
  } catch (error) {
    console.error('获取应用信息失败：', error)
    message.error('获取应用信息失败')
    router.push('/')
  }
}

const sendInitialMessage = async (prompt: string) => {
  await handleSendMessage(prompt)
}

const uploadProjectFile = async (file: File) => {
  if (!appId.value) throw new Error('应用ID不存在')
  const formData = new FormData()
  formData.append('appId', String(appId.value))
  formData.append('file', file)
  const res = await request.post('/app/file/add', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
    timeout: 10 * 60 * 1000,
  })
  const data = res.data as API.BaseResponseString | string
  if (typeof data === 'string') {
    throw new Error(data)
  }
  if (data.code !== 0) {
    throw new Error(getResponseErrorMessage(data, '文件上传失败'))
  }
}

const handleSendMessage = async (rawMessage: string, files: File[] = []) => {
  if (!rawMessage.trim() && files.length === 0) return
  if (isGenerating.value && files.length > 0) {
    message.warning('AI 工作中不能追加文件，请等待当前任务结束')
    return
  }
  if (sendingMessage.value) return
  sendingMessage.value = true
  try {
    for (const file of files) {
      await uploadProjectFile(file)
    }
    if (files.length > 0) {
      await loadChatHistory()
    }
    if (!rawMessage.trim()) {
      userInput.value = ''
      return
    }
    let msg = rawMessage.trim()
    if (selectedElementInfo.value) {
      let elementContext = `\n\n选中元素信息：`
      if (selectedElementInfo.value.pagePath) {
        elementContext += `\n- 页面路径: ${selectedElementInfo.value.pagePath}`
      }
      elementContext += `\n- 标签: ${selectedElementInfo.value.tagName.toLowerCase()}\n- 选择器: ${selectedElementInfo.value.selector}`
      if (selectedElementInfo.value.textContent) {
        elementContext += `\n- 当前内容: ${selectedElementInfo.value.textContent.substring(0, 100)}`
      }
      msg += elementContext
      clearSelectedElement()
      if (isEditMode.value) toggleEditMode()
    }

    const ok = isGenerating.value ? await pushMessage(msg) : await sendMessage(msg)
    if (!ok) {
      message.error(isGenerating.value ? '追加失败' : '提交失败')
      return
    }
    userInput.value = ''
    updatePreview()
  } catch (error) {
    console.error(files.length > 0 ? '文件上传失败：' : '消息提交失败：', error)
    message.error(getRequestErrorMessage(error, files.length > 0 ? '文件上传失败' : '提交失败'))
  } finally {
    sendingMessage.value = false
  }
}

const selectAnswerOption = (option: string) => {
  answerSelections.value = {
    ...answerSelections.value,
    [currentQuestionIndex.value]: option,
  }
}

const currentStepAnswer = () => {
  return (answerInput.value.trim() || currentAnswerSelection.value.trim()).trim()
}

const buildAnswerContent = (answers = answerSelections.value) => {
  const selected = currentQuestionItems.value
    .map((question, index) => {
      const value = answers[index]?.trim()
      if (!value) return ''
      if (currentQuestionItems.value.length === 1) return value
      return `问题：${question.question}\n回答：${value}`
    })
    .filter(Boolean)

  return selected.join('\n\n').trim()
}

const submitQuestionStep = async () => {
  const stepAnswer = currentStepAnswer()
  if (!stepAnswer) return
  const nextSelections = {
    ...answerSelections.value,
    [currentQuestionIndex.value]: stepAnswer,
  }
  answerSelections.value = nextSelections
  answerInput.value = ''

  if (!isLastQuestion.value) {
    currentQuestionIndex.value += 1
    return
  }

  const answer = buildAnswerContent(nextSelections)
  if (!answer) return
  answeringQuestion.value = true
  try {
    const ok = await answerQuestion(answer)
    if (!ok) {
      message.error('回答提交失败')
      return
    }
    currentQuestion.value = null
    answerSelections.value = {}
  } finally {
    answeringQuestion.value = false
  }
}

// 更新预览（已统一生成类型：不再依赖 codeGenType）
const updatePreview = (cacheBust = false) => {
  if (appId.value) {
    // 统一由后端/环境配置决定预览入口
    const baseUrl = getStaticPreviewUrl(appId.value)
    const newPreviewUrl = cacheBust ? `${baseUrl}?t=${Date.now()}` : baseUrl
    previewUrl.value = newPreviewUrl
    previewReady.value = true
  }
}

watch(
  () => aiState.value?.status,
  (status, previous) => {
    if (status === 'done' && previous !== 'done') {
      updatePreview(true)
    }
  },
)

watch(
  () => currentQuestion.value?.id,
  () => {
    currentQuestionIndex.value = 0
    answerInput.value = ''
    answerSelections.value = {}
    answeringQuestion.value = false
  },
)

// 滚动到底部
const scrollToBottom = () => {
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

watch(
  () => {
    const last = messages.value[messages.value.length - 1]
    return `${messages.value.length}:${last?.id || ''}:${last?.content?.length || 0}:${last?.toolCalls?.length || 0}`
  },
  () => nextTick(scrollToBottom),
)

// 下载代码
const downloadCode = async () => {
  if (!appId.value) {
    message.error('应用ID不存在')
    return
  }
  downloading.value = true
  try {
    const baseUrl = request.defaults.baseURL || ''
    const url = `${baseUrl}/app/download/${appId.value}`
    const response = await fetch(url, {
      method: 'GET',
      credentials: 'include',
    })
    if (!response.ok) {
      throw new Error(`下载失败: ${response.status}`)
    }
    // 获取文件名
    const contentDisposition = response.headers.get('Content-Disposition')
    const fileName =
      contentDisposition?.match(/filename="(.+)"/)?.[1] || `app-${appId.value}.zip`
    // 下载文件
    const blob = await response.blob()
    const downloadUrl = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = downloadUrl
    link.download = fileName
    link.click()
    // 清理
    URL.revokeObjectURL(downloadUrl)
    message.success('代码下载成功')
  } catch (error) {
    console.error('下载失败：', error)
    message.error('下载失败，请重试')
  } finally {
    downloading.value = false
  }
}

// 部署应用
const deployApp = async () => {
  if (!appId.value) {
    message.error('应用ID不存在')
    return
  }

  deploying.value = true
  try {
    const res = await deployAppApi({
      appId: Number(appId.value),
    })
    if (res.data.code === 0 && res.data.data) {
      const deployPath = res.data.data
      deployUrl.value = new URL(deployPath, window.location.origin).toString()
      await fetchAppInfo()
      deployModalVisible.value = true
      message.success('部署成功')
    } else {
      message.error('部署失败：' + res.data.message)
    }
  } catch (error) {
    console.error('部署失败：', error)
    message.error('部署失败，请重试')
  } finally {
    deploying.value = false
  }
}

// 在新窗口打开预览
const openInNewTab = () => {
  if (previewUrl.value) {
    window.open(previewUrl.value, '_blank')
  }
}

// 打开部署的网站
const openDeployedSite = () => {
  if (deployUrl.value) {
    window.open(deployUrl.value, '_blank')
  }
}

// iframe加载完成
const previewIframe = ref<HTMLIFrameElement | null>(null)
const onIframeLoad = () => {
  console.log('[iframe] loaded:', previewUrl.value)
  previewReady.value = true

  const iframe = previewIframe.value
  if (!iframe) {
    console.warn('[iframe] ref is null')
    return
  }

  try {
    // 关键：这里如果跨域会直接报错
    void iframe.contentDocument // 触发一下访问，能快速发现跨域问题
    visualEditor.init(iframe)
    visualEditor.onIframeLoad()
    console.log('[visualEditor] init OK')
  } catch (e) {
    console.error('[visualEditor] init FAILED (maybe cross-origin):', e)
    message.error('预览页跨域，无法开启点击选择')
  }
}


// 编辑应用
const editApp = () => {
  if (appInfo.value?.id) {
    router.push(`/app/edit/${appInfo.value.id}`)
  }
}

// 删除应用
const deleteApp = async () => {
  if (!appInfo.value?.id) return

  try {
    const res = await deleteAppApi({ id: appInfo.value.id })
    if (res.data.code === 0) {
      message.success('删除成功')
      appDetailVisible.value = false
      router.push('/')
    } else {
      message.error('删除失败：' + res.data.message)
    }
  } catch (error) {
    console.error('删除失败：', error)
    message.error('删除失败')
  }
}

// 可视化编辑相关函数
const toggleEditMode = () => {
  console.log('[toggleEditMode] click', { previewReady: previewReady.value, previewUrl: previewUrl.value })

  const iframe = previewIframe.value
  if (!iframe) {
    message.warning('预览 iframe 未挂载')
    return
  }
  if (!previewReady.value) {
    message.warning('请等待页面加载完成')
    return
  }

  try {
    const newEditMode = visualEditor.toggleEditMode()
    isEditMode.value = newEditMode
    console.log('[toggleEditMode] newEditMode=', newEditMode)
  } catch (e) {
    console.error('[toggleEditMode] failed:', e)
    message.error('开启点击选择失败（可能跨域或未初始化）')
  }
}


const clearSelectedElement = () => {
  selectedElementInfo.value = null
  visualEditor.clearSelection()
}

const getInputPlaceholder = () => {
  if (selectedElementInfo.value) {
    return `正在编辑 ${selectedElementInfo.value.tagName.toLowerCase()} 元素，描述您想要的修改...`
  }
  return '请描述你想生成的网站，越详细效果越好哦'
}

const formatJSON = (value?: string) => {
  if (!value) return ''
  try {
    return JSON.stringify(JSON.parse(value), null, 2)
  } catch {
    return value
  }
}

const handleWindowMessage = (event: MessageEvent) => {
  visualEditor.handleIframeMessage(event)
}

// 页面加载时获取应用信息
onMounted(async () => {
  await fetchAppInfo()
  await loadInitialState()
  window.addEventListener('message', handleWindowMessage)
})

// 清理资源
onUnmounted(() => {
  stopAI()
  window.removeEventListener('message', handleWindowMessage)
})
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}

.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
  background: #e2e8f0;
  border-radius: 10px;
}

.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #cbd5e1;
}

.chat-list {
  padding: 22px 22px 18px;
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.empty-chat {
  min-height: 320px;
  height: 100%;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  text-align: center;
  user-select: none;
}

.empty-title {
  color: #111827;
  font-size: 16px;
  font-weight: 600;
}

.empty-subtitle {
  margin-top: 6px;
  color: #9ca3af;
  font-size: 13px;
}

.system-row {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #9ca3af;
  font-size: 12px;
}

.system-row::before,
.system-row::after {
  content: '';
  height: 1px;
  flex: 1;
  background: #f1f5f9;
}

.message-row {
  display: flex;
}

.user-row {
  justify-content: flex-end;
}

.assistant-row {
  justify-content: flex-start;
}

.user-message {
  max-width: 86%;
  padding: 10px 14px;
  border-radius: 18px;
  background: #f3f4f6;
  color: #111827;
}

.assistant-message {
  max-width: 94%;
  color: #111827;
}

.push-label,
.agent-label {
  margin-bottom: 4px;
  color: #9ca3af;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 11px;
}

.message-markdown {
  font-size: 14px;
}

.user-markdown {
  line-height: 1.55;
}

.file-message {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 240px;
  max-width: 360px;
  cursor: pointer;
  user-select: none;
}

.file-message:hover .file-name {
  color: #2563eb;
}

.file-icon {
  flex: 0 0 auto;
  color: #2563eb;
  font-size: 22px;
}

.file-main {
  min-width: 0;
}

.file-name {
  overflow: hidden;
  color: #111827;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 3px;
  color: #64748b;
  font-size: 11px;
  line-height: 1.4;
}

.assistant-markdown {
  line-height: 1.7;
}

.assistant-loading {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  color: #9ca3af;
  font-size: 12px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.loading-dot {
  width: 7px;
  height: 7px;
  border-radius: 999px;
  background: #9ca3af;
  animation: pulse-dot 1.2s ease-in-out infinite;
}

@keyframes pulse-dot {
  0%, 100% { opacity: 0.3; }
  50% { opacity: 1; }
}

.tool-list {
  margin-top: 10px;
  display: grid;
  gap: 6px;
}

.tool-item {
  color: #6b7280;
  font-size: 12px;
}

.tool-item summary {
  display: flex;
  align-items: center;
  gap: 7px;
  cursor: pointer;
  list-style: none;
}

.tool-item summary::-webkit-details-marker {
  display: none;
}

.tool-corner {
  width: 12px;
  height: 13px;
  border-left: 1px solid #d1d5db;
  border-bottom: 1px solid #d1d5db;
  border-bottom-left-radius: 5px;
}

.tool-status {
  color: #9ca3af;
}

.tool-name {
  color: #374151;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.tool-code {
  margin: 7px 0 2px 20px;
  max-height: 180px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-word;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  background: #f9fafb;
  padding: 8px 10px;
  color: #374151;
  font-size: 11px;
  line-height: 1.5;
}

.composer-wrap {
  padding: 14px 16px 12px;
  background: #fff;
  border-top: 1px solid #f1f5f9;
}

.question-panel {
  margin-bottom: 12px;
  padding: 14px;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.05);
}

.question-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 8px;
  color: #374151;
  font-size: 12px;
  font-weight: 600;
}

.question-count {
  color: #9ca3af;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-weight: 500;
}

.question-item {
  margin-bottom: 12px;
}

.question-content {
  color: #111827;
  font-size: 14px;
  line-height: 1.65;
  white-space: pre-wrap;
}

.question-options {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
}

.question-option {
  max-width: 100%;
  min-height: 30px;
  padding: 5px 10px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  background: #f9fafb;
  color: #374151;
  font-size: 12px;
  white-space: normal;
  text-align: left;
  transition: border-color 0.15s ease, background 0.15s ease, color 0.15s ease;
}

.question-option:hover {
  border-color: #cbd5e1;
  background: #f3f4f6;
}

.question-option-active {
  border-color: #2563eb !important;
  color: #1d4ed8 !important;
  background: #eff6ff !important;
}

.composer-hint {
  margin-top: 8px;
  padding: 0 4px;
  color: #9ca3af;
  font-size: 11px;
}

:deep(.message-markdown .custom-md p) {
  margin: 0;
}

:deep(.message-markdown .custom-md p + p) {
  margin-top: 0.6em;
}

:deep(.message-markdown .custom-md pre) {
  margin: 10px 0;
  border-radius: 8px;
  background: #f8fafc;
}

:deep(.message-markdown .custom-md code) {
  font-size: 12px;
}
</style>
