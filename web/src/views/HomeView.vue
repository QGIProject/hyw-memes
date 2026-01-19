<script setup>
import { ref, onMounted } from 'vue'
import { imageApi, categoryApi } from '../services/api'
import { RocketOutlined, ReloadOutlined, PictureOutlined } from '@ant-design/icons-vue'

const images = ref([])
const loading = ref(true)
const randomImage = ref(null)
const showModal = ref(false)
const page = ref(1)
const total = ref(0)
const limit = 20

const categories = ref([])
const activeCategory = ref('')

const fetchCategories = async () => {
  try {
    const res = await categoryApi.getAll()
    categories.value = res.data
  } catch (err) { console.error(err) }
}

const fetchImages = async () => {
  loading.value = true
  try {
    const res = await imageApi.getApproved(page.value, limit, activeCategory.value)
    if (page.value === 1) images.value = res.data.images || []
    else images.value = [...images.value, ...(res.data.images || [])]
    total.value = res.data.total
  } catch (err) { console.error(err) } 
  finally { loading.value = false }
}

const getRandomImage = async () => {
  try {
    const res = await imageApi.getRandom(activeCategory.value)
    randomImage.value = res.data
    showModal.value = true
  } catch (err) { alert('没有找到图片') }
}

const loadMore = () => { page.value++; fetchImages() }

const selectCategory = (id) => {
  activeCategory.value = id
  page.value = 1
  fetchImages()
}

onMounted(() => {
  fetchCategories()
  fetchImages()
})
</script>

<template>
  <div class="container mx-auto px-4 py-8">
    <!-- Hero -->
    <div class="text-center py-12 border-b mb-8">
      <h1 class="text-5xl font-black mb-4 tracking-tighter">HYW PICS</h1>
      <p class="text-lg text-gray-500 mb-8 font-medium">Raw, Filtered, Authentic.</p>
      <a-button type="primary" size="large" @click="getRandomImage" class="h-12 px-8 text-lg">
        <template #icon><RocketOutlined /></template> RANDOM EXPLORE
      </a-button>
    </div>

    <!-- Filter -->
    <div class="flex justify-center flex-wrap gap-2 mb-12">
      <a-button 
        :type="activeCategory === '' ? 'primary' : 'default'" 
        @click="selectCategory('')"
      >ALL</a-button>
      <a-button 
        v-for="cat in categories" 
        :key="cat.id"
        :type="activeCategory === cat.id ? 'primary' : 'default'"
        @click="selectCategory(cat.id)"
      >{{ cat.name }}</a-button>
    </div>

    <!-- Content -->
    <div v-if="loading && page === 1" class="text-center py-20">
      <a-spin size="large" />
    </div>

    <div v-else-if="images.length === 0" class="py-20">
      <a-empty description="No Images Yet">
        <router-link to="/upload"><a-button type="primary">Upload Now</a-button></router-link>
      </a-empty>
    </div>

    <div v-else class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
      <div 
        v-for="img in images" 
        :key="img.id" 
        class="border bg-white p-2 cursor-pointer hover:shadow-lg transition-all duration-300"
        @click="randomImage = img; showModal = true"
      >
        <img 
          :src="`/uploads/${img.filename}`" 
          class="w-full h-auto object-cover grayscale hover:grayscale-0 transition-all duration-500"
          loading="lazy"
        />
      </div>
    </div>

    <!-- Load More -->
    <div v-if="images.length > 0 && images.length < total" class="text-center mt-12 mb-8">
      <a-button size="large" @click="loadMore" :loading="loading" class="w-48">LOAD MORE</a-button>
    </div>

    <!-- Preview Modal -->
    <a-modal v-model:open="showModal" :footer="null" width="80%" wrapClassName="full-modal" :bodyStyle="{ padding: 0 }">
       <div class="bg-gray-100 flex justify-center items-center min-h-[50vh] p-4 relative">
          <img 
            v-if="randomImage"
            :src="`/uploads/${randomImage.filename}`" 
            class="max-w-full max-h-[80vh] shadow-xl"
          />
       </div>
       <div class="p-4 flex justify-between items-center bg-white">
          <div>
            <h3 class="text-lg font-bold font-mono m-0">{{ randomImage?.original_name }}</h3>
             <a-tag v-if="randomImage?.category_id" class="mt-2">
              {{ categories.find(c => c.id === randomImage.category_id)?.name }}
            </a-tag>
          </div>
          <a-button type="primary" size="large" @click="getRandomImage">
            <template #icon><ReloadOutlined /></template> NEXT RANDOM
          </a-button>
       </div>
    </a-modal>
  </div>
</template>
