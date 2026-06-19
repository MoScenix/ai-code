<template>
  <div
    class="prompt-box"
    @dragover.prevent
    @drop.prevent="handleDrop"
  >
    <div v-if="files.length > 0" class="flex flex-wrap gap-2 px-4 pt-3">
      <div v-for="(file, index) in files" :key="index" class="relative group">
        <img
          :src="file.url"
          :alt="file.name"
          class="w-14 h-14 rounded-xl object-cover cursor-pointer"
          @click="previewImage = file.url"
        />
        <button class="remove-file" @click="files.splice(index, 1)">×</button>
      </div>
    </div>

    <div class="px-4 pt-3 pb-1">
      <textarea
        ref="textareaRef"
        :value="modelValue"
        rows="1"
        :placeholder="placeholder"
        :disabled="disabled"
        class="prompt-textarea"
        @input="onInput"
        @keydown="onKeydown"
      />
    </div>

    <div class="flex items-center justify-between px-3 pb-2">
      <div class="flex items-center gap-1">
        <button class="icon-btn" title="上传图片" @click="fileInputRef?.click()">
          <PaperClipOutlined />
        </button>
        <input ref="fileInputRef" type="file" accept="image/*" class="hidden" @change="onFileChange" />

        <div class="w-px h-5 bg-gray-200 mx-1"></div>

        <button class="mode-btn" :class="{ active: mode === 'search' }" @click="toggleMode('search')">
          <GlobalOutlined />
          <span v-if="mode === 'search'">搜索</span>
        </button>
        <button class="mode-btn think" :class="{ active: mode === 'think' }" @click="toggleMode('think')">
          <BulbOutlined />
          <span v-if="mode === 'think'">思考</span>
        </button>
      </div>

      <button
        class="send-btn"
        :class="{ active: hasContent, stop: isLoading && !hasContent, idle: !isLoading && !hasContent }"
        :disabled="disabled && !isLoading"
        :title="buttonTitle"
        @click="handleButtonClick"
      >
        <ArrowUpOutlined v-if="hasContent || !isLoading" />
        <svg v-else viewBox="0 0 24 24" class="w-3.5 h-3.5" fill="currentColor">
          <rect x="6" y="6" width="12" height="12" rx="1.5" />
        </svg>
      </button>
    </div>
  </div>

  <Teleport to="body">
    <div v-if="previewImage" class="fixed inset-0 z-[9999] bg-black/60 flex items-center justify-center" @click="previewImage = null">
      <img :src="previewImage" class="max-w-[90vw] max-h-[85vh] rounded-2xl shadow-2xl" />
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { ArrowUpOutlined, BulbOutlined, GlobalOutlined, PaperClipOutlined } from '@ant-design/icons-vue'

type Mode = '' | 'search' | 'think'

const props = defineProps<{
  modelValue: string
  isLoading?: boolean
  disabled?: boolean
  placeholder?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  send: [message: string, files?: File[]]
  cancel: []
}>()

const textareaRef = ref<HTMLTextAreaElement>()
const fileInputRef = ref<HTMLInputElement>()
const mode = ref<Mode>('')
const previewImage = ref<string | null>(null)
const files = ref<Array<{ name: string; url: string; file: File }>>([])

const hasContent = computed(() => props.modelValue.trim().length > 0 || files.value.length > 0)
const buttonTitle = computed(() => {
  if (hasContent.value && props.isLoading) return '追加到当前任务'
  if (hasContent.value) return '发送'
  if (props.isLoading) return '停止'
  return '无内容'
})

function autoResize() {
  if (!textareaRef.value) return
  textareaRef.value.style.height = 'auto'
  textareaRef.value.style.height = `${Math.min(textareaRef.value.scrollHeight, 180)}px`
}

watch(() => props.modelValue, () => nextTick(autoResize))

function onInput(e: Event) {
  emit('update:modelValue', (e.target as HTMLTextAreaElement).value)
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    handleButtonClick()
  }
}

function toggleMode(next: Mode) {
  mode.value = mode.value === next ? '' : next
  nextTick(() => textareaRef.value?.focus())
}

function handleButtonClick() {
  if (!hasContent.value) {
    if (props.isLoading) emit('cancel')
    return
  }
  let content = props.modelValue.trim()
  if (mode.value === 'search') content = `[搜索: ${content}]`
  if (mode.value === 'think') content = `[思考: ${content}]`
  emit('send', content, files.value.map((item) => item.file))
  emit('update:modelValue', '')
  files.value = []
}

function onFileChange() {
  const file = fileInputRef.value?.files?.[0]
  if (file) processFile(file)
  if (fileInputRef.value) fileInputRef.value.value = ''
}

function handleDrop(e: DragEvent) {
  const file = Array.from(e.dataTransfer?.files || []).find((item) => item.type.startsWith('image/'))
  if (file) processFile(file)
}

function processFile(file: File) {
  if (!file.type.startsWith('image/')) return
  const reader = new FileReader()
  reader.onload = (e) => {
    files.value = [{ name: file.name, url: e.target?.result as string, file }]
  }
  reader.readAsDataURL(file)
}
</script>

<style scoped>
.prompt-box {
  display: flex;
  flex-direction: column;
  border: 1px solid #e5e7eb;
  border-radius: 24px;
  background: #fff;
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.05);
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.prompt-box:focus-within {
  border-color: #cbd5e1;
  box-shadow: 0 10px 28px rgba(15, 23, 42, 0.08);
}

.prompt-textarea {
  width: 100%;
  min-height: 28px;
  max-height: 180px;
  border: none;
  outline: none;
  resize: none;
  background: transparent;
  color: #1f2937;
  font-size: 14px;
  line-height: 1.6;
}

.prompt-textarea::placeholder {
  color: #9ca3af;
}

.icon-btn,
.mode-btn,
.send-btn {
  height: 32px;
  border: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
}

.icon-btn {
  width: 32px;
  border-radius: 999px;
  color: #9ca3af;
  background: transparent;
}

.icon-btn:hover {
  color: #4b5563;
  background: #f3f4f6;
}

.mode-btn {
  gap: 4px;
  padding: 0 10px;
  border-radius: 999px;
  color: #9ca3af;
  background: transparent;
  font-size: 12px;
}

.mode-btn:hover {
  color: #4b5563;
  background: #f3f4f6;
}

.mode-btn.active {
  color: #111827;
  background: #f3f4f6;
}

.mode-btn.think.active {
  color: #111827;
  background: #f3f4f6;
}

.send-btn {
  width: 32px;
  border-radius: 999px;
  background: #f3f4f6;
  color: #9ca3af;
}

.send-btn.idle {
  cursor: default;
}

.send-btn.active,
.send-btn.stop {
  background: #111827;
  color: #fff;
}

.send-btn.active:hover,
.send-btn.stop:hover {
  background: #374151;
}

.remove-file {
  position: absolute;
  top: -6px;
  right: -6px;
  width: 20px;
  height: 20px;
  border-radius: 999px;
  border: none;
  color: #fff;
  background: #111827;
  opacity: 0;
  transition: opacity 0.15s ease;
}

.group:hover .remove-file {
  opacity: 1;
}
</style>
