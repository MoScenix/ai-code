<template>
  <div class="h-screen w-full flex flex-col bg-slate-50 text-slate-900 overflow-hidden">
    <header
      class="h-16 flex items-center justify-between px-6 bg-white/80 backdrop-blur-md border-b border-slate-200 z-10">
      <div class="flex items-center gap-3">
        <div class="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center shadow-lg shadow-blue-200">
          <GlobalOutlined class="text-white text-lg" />
        </div>
        <h1 class="text-lg font-bold bg-clip-text text-transparent bg-gradient-to-r from-slate-900 to-slate-500">
          {{ appInfo?.appName || '网站生成器' }}
        </h1>
      </div>

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
        class="flex-1 min-w-[400px] flex flex-col bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden">
        <div ref="messagesContainer" class="flex-1 overflow-y-auto p-4 space-y-6 scroll-smooth custom-scrollbar">
          <div v-if="hasMoreHistory" class="flex justify-center">
            <a-button type="link" @click="loadMoreHistory" :loading="loadingHistory" size="small"
              class="text-slate-400 font-normal">
              查看更早的历史消息
            </a-button>
          </div>

          <div v-for="(message, index) in messages" :key="index"
            class="animate-in fade-in slide-in-from-bottom-2 duration-300">
            <div v-if="message.type === 'user'" class="flex justify-end gap-3 pl-12">
              <div class="max-w-[78%] bg-gradient-to-br from-blue-600 to-blue-500 text-white px-4 py-3
         rounded-2xl rounded-tr-md shadow-[0_10px_28px_rgba(37,99,235,0.25)]
         text-[15px] leading-[1.75] tracking-[0.2px] whitespace-pre-wrap break-words">
                {{ message.content }}
              </div>

              <a-avatar :src="loginUserStore.loginUser.userAvatar"
                class="flex-shrink-0 border-2 border-white shadow-sm" />
            </div>

            <div v-else class="flex justify-start gap-3 pr-12">
              <a-avatar :src="aiAvatar" class="flex-shrink-0 border border-slate-100 shadow-sm" />
              <div class="flex-1">
                <div class="max-w-[78%] bg-white text-slate-800 px-4 py-3 rounded-2xl rounded-tl-md
         border border-slate-200 shadow-[0_10px_28px_rgba(15,23,42,0.06)]
         text-[15px] leading-[1.75] tracking-[0.2px] relative break-words">
                  <div class="custom-md markdown-content">
                    <MarkdownRenderer v-if="message.content" :content="message.content" />
                  </div>

                  <div v-if="message.loading" class="flex items-center gap-2 py-2 text-slate-400">
                    <a-spin size="small" />
                    <span class="text-xs animate-pulse">正在构思方案...</span>
                  </div>
                </div>

              </div>
            </div>
          </div>
        </div>

        <div class="p-4 bg-white border-t border-slate-100">
          <div v-if="selectedElementInfo" class="mb-3">
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

          <div class="relative rounded-xl border-2 transition-all p-1" :class="[
            isEditMode ? 'border-amber-400 bg-amber-50/20' : 'border-slate-200 focus-within:border-blue-500 focus-within:ring-4 focus-within:ring-blue-500/10'
          ]">
            <a-textarea v-model:value="userInput" :placeholder="getInputPlaceholder()" :rows="4"
              :disabled="isGenerating || (!isOwner && !isAdmin)" :bordered="false"
              class="!bg-transparent !resize-none !text-slate-700 !shadow-none" @keydown.enter.prevent="sendMessage" />

            <div class="flex items-center justify-between px-2 pb-1">
              <div class="text-[11px] text-slate-400 flex items-center gap-1">
                <span v-if="isOwner">回车键发送，Shift+Enter 换行</span>
                <span v-else class="text-amber-500">
                  <LockOutlined /> 访客模式不可编辑
                </span>
              </div>
              <a-button type="primary" @click="sendMessage" :loading="isGenerating"
                :disabled="!isOwner || !userInput.trim()"
                class="!h-9 !w-9 !flex !items-center !justify-center !rounded-lg !bg-blue-600 shadow-lg shadow-blue-500/20">
                <template #icon>
                  <SendOutlined />
                </template>
              </a-button>
            </div>
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
            @click="updatePreview" />



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
import { ref, onMounted, nextTick, onUnmounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { useLoginUserStore } from '@/stores/loginUser'
import { getAppVoById, deployApp as deployAppApi, deleteApp as deleteAppApi } from '@/api/appController'
import { listAppChatHistory } from '@/api/chatHistoryController'
import request from '@/request'

import MarkdownRenderer from '@/components/MarkdownRenderer.vue'
import AppDetailModal from '@/components/AppDetailModal.vue'
import DeploySuccessModal from '@/components/DeploySuccessModal.vue'
import aiAvatar from '@/assets/aiAvatar.png'
import { API_BASE_URL, getStaticPreviewUrl } from '@/config/env'
import { VisualEditor, type ElementInfo } from '@/utils/visualEditor'

import {
  CloudUploadOutlined,
  SendOutlined,
  ExportOutlined,
  InfoCircleOutlined,
  DownloadOutlined,
  EditOutlined,
} from '@ant-design/icons-vue'

const route = useRoute()
const router = useRouter()
const loginUserStore = useLoginUserStore()

// 应用信息
const appInfo = ref<API.AppVO>()
const appId = ref<any>()

// 对话相关
interface Message {
  type: 'user' | 'ai'
  content: string
  loading?: boolean
  createTime?: string
}

const messages = ref<Message[]>([])
const userInput = ref('')
const isGenerating = ref(false)
const messagesContainer = ref<HTMLElement>()

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

// 应用详情相关
const appDetailVisible = ref(false)

// 显示应用详情
const showAppDetail = () => {
  appDetailVisible.value = true
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
        const historyMessages: Message[] = chatHistories
          .map((chat) => ({
            type: (chat.messageType === 'user' ? 'user' : 'ai') as 'user' | 'ai',
            content: chat.message || '',
            createTime: chat.createTime,
          }))
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
      if (appInfo.value.initPrompt && isOwner.value && messages.value.length === 0 && historyLoaded.value) {
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

// 发送初始消息
const sendInitialMessage = async (prompt: string) => {
  // 添加用户消息
  messages.value.push({
    type: 'user',
    content: prompt,
  })

  // 添加AI消息占位符
  const aiMessageIndex = messages.value.length
  messages.value.push({
    type: 'ai',
    content: '',
    loading: true,
  })

  await nextTick()
  scrollToBottom()

  // 开始生成
  isGenerating.value = true
  await generateCode(prompt, aiMessageIndex)
}

// 发送消息
const sendMessage = async () => {
  if (!userInput.value.trim() || isGenerating.value) {
    return
  }

  let msg = userInput.value.trim()
  // 如果有选中的元素，将元素信息添加到提示词中
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
  }
  userInput.value = ''

  // 添加用户消息（包含元素信息）
  messages.value.push({
    type: 'user',
    content: msg,
  })

  // 发送消息后，清除选中元素并退出编辑模式
  if (selectedElementInfo.value) {
    clearSelectedElement()
    if (isEditMode.value) {
      toggleEditMode()
    }
  }

  // 添加AI消息占位符
  const aiMessageIndex = messages.value.length
  messages.value.push({
    type: 'ai',
    content: '',
    loading: true,
  })

  await nextTick()
  scrollToBottom()

  // 开始生成
  isGenerating.value = true
  await generateCode(msg, aiMessageIndex)
}

// 生成代码 - 使用 EventSource 处理流式响应
const generateCode = async (userMessage: string, aiMessageIndex: number) => {
  let eventSource: EventSource | null = null
  let streamCompleted = false

  try {
    // 获取 axios 配置的 baseURL
    const baseURL = request.defaults.baseURL || API_BASE_URL

    // 构建URL参数
    const params = new URLSearchParams({
      appId: appId.value || '',
      message: userMessage,
    })

    const url = `${baseURL}/app/chat/gen/code?${params}`

    // 创建 EventSource 连接
    eventSource = new EventSource(url, {
      withCredentials: true,
    })

    let fullContent = ''

    // 处理接收到的消息
    eventSource.onmessage = function (event) {
      if (streamCompleted) return

      try {
        // 解析JSON包装的数据
        const parsed = JSON.parse(event.data)
        const content = parsed.d

        // 拼接内容
        if (content !== undefined && content !== null) {
          fullContent += content
          messages.value[aiMessageIndex].content = fullContent
          messages.value[aiMessageIndex].loading = false
          scrollToBottom()
        }
      } catch (error) {
        console.error('解析消息失败:', error)
        handleError(error, aiMessageIndex)
      }
    }

    // 处理done事件
    eventSource.addEventListener('done', function () {
      if (streamCompleted) return

      streamCompleted = true
      isGenerating.value = false
      eventSource?.close()

      // 延迟更新预览，确保后端已完成处理
      setTimeout(async () => {
        await fetchAppInfo()
        updatePreview()
      }, 1000)
    })

    // 处理business-error事件（后端限流等错误）
    eventSource.addEventListener('business-error', function (event: MessageEvent) {
      if (streamCompleted) return

      try {
        const errorData = JSON.parse(event.data)
        console.error('SSE业务错误事件:', errorData)

        // 显示具体的错误信息
        const errorMessage = errorData.message || '生成过程中出现错误'
        messages.value[aiMessageIndex].content = `❌ ${errorMessage}`
        messages.value[aiMessageIndex].loading = false
        message.error(errorMessage)

        streamCompleted = true
        isGenerating.value = false
        eventSource?.close()
      } catch (parseError) {
        console.error('解析错误事件失败:', parseError, '原始数据:', event.data)
        handleError(new Error('服务器返回错误'), aiMessageIndex)
      }
    })

    // 处理错误
    eventSource.onerror = function () {
      if (streamCompleted || !isGenerating.value) return
      // 检查是否是正常的连接关闭
      if (eventSource?.readyState === EventSource.CONNECTING) {
        streamCompleted = true
        isGenerating.value = false
        eventSource?.close()

        setTimeout(async () => {
          await fetchAppInfo()
          updatePreview()
        }, 1000)
      } else {
        handleError(new Error('SSE连接错误'), aiMessageIndex)
      }
    }
  } catch (error) {
    console.error('创建 EventSource 失败：', error)
    handleError(error, aiMessageIndex)
  }
}

// 错误处理函数
const handleError = (error: unknown, aiMessageIndex: number) => {
  console.error('生成代码失败：', error)
  messages.value[aiMessageIndex].content = '抱歉，生成过程中出现了错误，请重试。'
  messages.value[aiMessageIndex].loading = false
  message.error('生成失败，请重试')
  isGenerating.value = false
}

// 更新预览（已统一生成类型：不再依赖 codeGenType）
const updatePreview = () => {
  if (appId.value) {
    // 统一由后端/环境配置决定预览入口
    const newPreviewUrl = getStaticPreviewUrl(appId.value)
    previewUrl.value = newPreviewUrl
    previewReady.value = true
  }
}

// 滚动到底部
const scrollToBottom = () => {
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

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

// 页面加载时获取应用信息
onMounted(() => {
  fetchAppInfo()

  // 监听 iframe 消息
  window.addEventListener('message', (event) => {
    visualEditor.handleIframeMessage(event)
  })
})

// 清理资源
onUnmounted(() => {
  // EventSource 会在组件卸载时自动清理
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

/* 覆盖 Ant Design 部分全局样式 */
</style>
