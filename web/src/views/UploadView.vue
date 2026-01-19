<script setup>
import { ref } from 'vue'
import { imageApi } from '../services/api'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { InboxOutlined } from '@ant-design/icons-vue'

const router = useRouter()
const file = ref(null)
const uploading = ref(false)
const previewUrl = ref(null)

const beforeUpload = (f) => {
  file.value = f
  previewUrl.value = URL.createObjectURL(f)
  return false // Prevent auto upload
}

const handleRemove = () => {
  file.value = null
  previewUrl.value = null
}

const handleUpload = async () => {
  if (!file.value) return message.warning('请选择图片')
  
  uploading.value = true
  // Fix: Pass file object directly, api.js builds the FormData
  try {
    await imageApi.upload(file.value)
    message.success('上传成功！等待管理员审核。')
    setTimeout(() => router.push('/'), 1500)
  } catch (error) {
    message.error('上传失败')
  } finally {
    uploading.value = false
  }
}
</script>

<template>
  <div class="max-w-xl mx-auto py-12 px-4">
    <div class="mb-8 text-center">
      <h1 class="text-3xl font-bold mb-2">Upload Image</h1>
      <p class="text-gray-500">Share your meme with the world</p>
    </div>

    <a-card :bordered="false" class="shadow-md">
      <div v-if="!file" class="h-64">
        <a-upload-dragger
          name="image"
          :multiple="false"
          :beforeUpload="beforeUpload"
          :showUploadList="false"
          class="h-full"
        >
          <p class="ant-upload-drag-icon">
            <InboxOutlined />
          </p>
          <p class="ant-upload-text">点击或拖拽图片到此处上传</p>
          <p class="ant-upload-hint">支持 JPG, PNG, GIF, WebP 格式</p>
        </a-upload-dragger>
      </div>

      <div v-else class="text-center">
        <div class="mb-4 bg-gray-100 rounded p-4 relative group">
           <img :src="previewUrl" class="max-h-64 mx-auto object-contain" />
           <div class="absolute inset-0 bg-black/50 opacity-0 group-hover:opacity-100 flex items-center justify-center transition-opacity rounded">
             <a-button type="primary" danger @click="handleRemove">更换图片</a-button>
           </div>
        </div>
        <p class="font-mono mb-6 truncate">{{ file.name }}</p>
        
        <div class="flex gap-4">
          <a-button block size="large" @click="handleRemove" :disabled="uploading">取消</a-button>
          <a-button type="primary" block size="large" @click="handleUpload" :loading="uploading">确认上传</a-button>
        </div>
      </div>
    </a-card>
    
    <div class="text-center mt-6">
      <router-link to="/"><a-button type="link">返回首页</a-button></router-link>
    </div>
  </div>
</template>
