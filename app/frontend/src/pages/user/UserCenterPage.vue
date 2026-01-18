<template>
  <div id="userCenterPage" class="min-h-[calc(100vh-120px)] bg-[#f8fafc] py-10 px-4">
    <div class="max-w-5xl mx-auto flex flex-col md:flex-row gap-8">
      
      <div class="w-full md:w-1/3">
        <div class="group bg-white border border-gray-100 rounded-2xl p-8 transition-all duration-300 hover:-translate-y-1.5 hover:shadow-xl text-center">
          <div class="relative mb-6 inline-block">
            <a-avatar 
              :size="100" 
              :src="formState.userAvatar || loginUser.userAvatar" 
              class="border-4 border-blue-50 shadow-sm transition-all duration-500" 
            />
            <div class="absolute bottom-1 right-1 w-6 h-6 bg-green-500 border-4 border-white rounded-full"></div>
          </div>
          
          <h2 class="text-xl font-bold text-gray-800 mb-1">{{ loginUser.userName ?? '无名用户' }}</h2>
          <p class="text-gray-400 text-sm mb-6">{{ loginUser.userAccount }}</p>
          
          <div class="w-full space-y-3">
            <div class="flex justify-between items-center text-sm py-2 border-b border-gray-50">
              <span class="text-gray-400">用户角色</span>
              <span class="px-3 py-1 bg-blue-50 text-blue-600 rounded-full text-xs font-bold">
                {{ loginUser.userRole === 'admin' ? '管理员' : '普通用户' }}
              </span>
            </div>
            <div class="flex justify-between items-center text-sm py-2 border-b border-gray-50">
              <span class="text-gray-400">注册时间</span>
              <span class="text-gray-600 font-medium">{{ formatDate(loginUser.createTime) }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="w-full md:w-2/3">
        <div class="bg-white border border-gray-100 rounded-2xl p-8 md:p-10 shadow-sm">
          <div class="flex items-center gap-3 mb-8">
            <div class="w-1.5 h-6 bg-blue-600 rounded-full"></div>
            <h3 class="text-lg font-bold text-gray-800 m-0">基本资料设置</h3>
          </div>

          <a-form :model="formState" layout="vertical" @finish="onFinish">
            
            <div class="mb-8">
              <label class="block text-sm font-medium text-gray-500 mb-4">个人头像</label>
              <div class="flex flex-col sm:flex-row items-center gap-6">
                <a-upload
                  name="file"
                  list-type="picture-card"
                  class="!m-0 overflow-hidden"
                  :show-upload-list="false"
                  action="/api/file/upload"
                  @change="handleAvatarChange"
                >
                  <div v-if="formState.userAvatar" class="w-full h-full relative group/avatar">
                    <img :src="formState.userAvatar" class="w-full h-full object-cover rounded-xl" />
                    <div class="absolute inset-0 bg-black/40 opacity-0 group-hover/avatar:opacity-100 flex items-center justify-center transition-opacity rounded-xl">
                      <span class="text-white text-xs">更换图片</span>
                    </div>
                  </div>
                  <div v-else class="flex flex-col items-center">
                    <loading-outlined v-if="loading" class="text-blue-500" />
                    <plus-outlined v-else class="text-gray-400 text-lg" />
                    <div class="mt-2 text-[11px] text-gray-400">点击上传</div>
                  </div>
                </a-upload>

                <div class="flex-grow w-full">
                  <div class="text-[11px] text-gray-400 mb-1.5 ml-1 flex items-center gap-1">
                    <link-outlined /> 或者输入外部链接
                  </div>
                  <a-input 
                    v-model:value="formState.userAvatar" 
                    placeholder="https://example.com/avatar.jpg" 
                    size="large"
                    class="!rounded-xl !border-gray-100 hover:!border-blue-200 focus:!border-blue-400 !bg-gray-50/30"
                  />
                </div>
              </div>
            </div>

            <a-form-item label="用户昵称" name="userName" class="mb-6">
              <a-input 
                v-model:value="formState.userName" 
                placeholder="起个好听的名字吧" 
                size="large" 
                class="!rounded-xl !border-gray-100"
              />
            </a-form-item>

            <a-form-item label="个人简介" name="userProfile" class="mb-8">
              <a-textarea 
                v-model:value="formState.userProfile" 
                placeholder="介绍一下你自己..." 
                :rows="4" 
                class="!rounded-xl !border-gray-100 !p-4"
              />
            </a-form-item>

            <div class="flex justify-end border-t border-gray-50 pt-6">
              <a-button 
                type="primary" 
                html-type="submit" 
                size="large"
                class="!h-12 !px-12 !rounded-xl !bg-blue-600 hover:!bg-blue-700 !border-none font-bold text-white shadow-lg shadow-blue-100 transition-all active:scale-95"
              >
                保存资料
              </a-button>
            </div>
          </a-form>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, onMounted, computed, ref } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined, LoadingOutlined, LinkOutlined } from '@ant-design/icons-vue'
import { useLoginUserStore } from '@/stores/loginUser.ts'
import { updateUser, getLoginUser } from '@/api/userController.ts'

const loginUserStore = useLoginUserStore()
const loginUser = computed(() => loginUserStore.loginUser)
const loading = ref<boolean>(false)

const formState = reactive<API.UserUpdateRequest>({
  id: undefined,
  userName: '',
  userAvatar: '',
  userProfile: '',
})

/**
 * 格式化注册时间
 */
const formatDate = (date?: string) => {
  if (!date) return '未知'
  return date.split('T')[0]
}

/**
 * 处理头像上传
 */
const handleAvatarChange = (info: any) => {
  if (info.file.status === 'uploading') {
    loading.value = true
    return
  }
  if (info.file.status === 'done') {
    const res = info.file.response
    if (res.code === 0) {
      formState.userAvatar = res.data
      message.success('本地上传成功')
    } else {
      message.error('上传失败：' + res.message)
    }
    loading.value = false
  }
}

/**
 * 初始化加载
 */
const loadData = async () => {
  const res = await getLoginUser()
  if (res.data.code === 0 && res.data.data) {
    const user = res.data.data
    formState.id = user.id
    formState.userName = user.userName
    formState.userAvatar = user.userAvatar
    formState.userProfile = user.userProfile
  }
}

/**
 * 提交保存
 */
const onFinish = async (values: API.UserUpdateRequest) => {
  if (!formState.id) return
  const res = await updateUser({
    ...values,
    id: formState.id,
    userAvatar: formState.userAvatar
  })
  if (res.data.code === 0) {
    message.success('资料更新成功')
    await loginUserStore.fetchLoginUser()
  } else {
    message.error('更新失败：' + res.data.message)
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
:deep(.ant-upload.ant-upload-select-picture-card) {
  width: 100px !important;
  height: 100px !important;
  border-radius: 12px !important;
  border: 2px dashed #f3f4f6 !important;
  background-color: #fafafa !important;
  margin: 0 !important;
}

:deep(.ant-form-item-label > label) {
  font-weight: 600 !important;
  color: #6b7280 !important;
}
</style>