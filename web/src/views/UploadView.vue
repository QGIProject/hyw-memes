<script setup>
import { ref } from 'vue'
import { imageApi } from '../services/api'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { InboxOutlined } from '@ant-design/icons-vue'

const router = useRouter()
const fileList = ref([])
const uploading = ref(false)

const beforeUpload = (f) => {
  fileList.value = [...fileList.value, f]
  return false // Prevent auto upload
}

const handleRemove = (file) => {
  const index = fileList.value.indexOf(file)
  const newFileList = fileList.value.slice()
  newFileList.splice(index, 1)
  fileList.value = newFileList
}

const handleUpload = async () => {
  if (fileList.value.length === 0) return message.warning('请选择图片')
  
  uploading.value = true
  try {
    await imageApi.upload(fileList.value)
    message.success(`成功上传 ${fileList.value.length} 张图片！等待管理员审核。`)
    setTimeout(() => router.push('/'), 1500)
  } catch (error) {
    message.error('上传失败，请检查网络或文件格式')
  } finally {
    uploading.value = false
  }
}
</script>

<template>
  <div class="max-w-2xl mx-auto py-12 px-4">
    <div class="mb-8 text-center">
      <h1 class="text-3xl font-bold mb-2">批量上传</h1>
      <p class="text-gray-500">支持同时上传多个文件，上传后需等待管理员审核</p>
    </div>

    <a-card :bordered="false" class="shadow-md">
      <a-upload-dragger
        name="images"
        :multiple="true"
        :beforeUpload="beforeUpload"
        :file-list="fileList"
        @remove="handleRemove"
        accept="image/*"
      >
        <p class="ant-upload-drag-icon">
          <InboxOutlined />
        </p>
        <p class="ant-upload-text">点击或拖拽图片到此处上传</p>
        <p class="ant-upload-hint">支持 JPG, PNG, GIF 格式</p>
      </a-upload-dragger>

      <div v-if="fileList.length > 0" class="mt-8 pt-6 border-t font-sans">
        <div class="flex justify-between items-center mb-6">
          <span class="text-gray-600">已选择 <span class="font-bold text-blue-600">{{ fileList.length }}</span> 个文件</span>
          <a-button type="link" danger @click="fileList = []">清空列表</a-button>
        </div>
        
        <div class="flex gap-4">
          <a-button block size="large" @click="$router.push('/')" :disabled="uploading">取消</a-button>
          <a-button type="primary" block size="large" @click="handleUpload" :loading="uploading">
            确认上传
          </a-button>
        </div>
      </div>
    </a-card>
    
    <div class="text-center mt-6">
      <router-link to="/"><a-button type="link">返回首页</a-button></router-link>
    </div>
  </div>
</template>
