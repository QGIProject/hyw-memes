<script setup>
import { ref, onMounted, reactive, watch } from 'vue'
import { adminApi, categoryApi, imageApi } from '../services/api'
import { message, Modal } from 'ant-design-vue'
import { 
  DashboardOutlined, AuditOutlined, FileImageOutlined, FolderOutlined, LogoutOutlined,
  CheckCircleOutlined, CloseCircleOutlined, DeleteOutlined, EyeOutlined, EditOutlined, PlusOutlined
} from '@ant-design/icons-vue'

const isLoggedIn = ref(false)
const password = ref('')
const loginLoading = ref(false)

// Menu State: Ant Design Menu expects string[]
const selectedKeys = ref(['dashboard'])
const activeKey = ref('dashboard') 

watch(selectedKeys, (val) => {
    if (val && val.length > 0) {
        activeKey.value = val[0]
        if (val[0] === 'audit') loadPending()
        if (val[0] === 'images') loadAllImages()
    }
})

const stats = ref({ total_images: 0, pending_images: 0, approved_images: 0, total_categories: 0 })
const categories = ref([])
const pendingImages = ref([])
const allImages = ref([])
const imagePagination = reactive({ current: 1, pageSize: 10, total: 0 })
const allImagesLoading = ref(false)

// Debug
console.log('AdminView Setup')

// Login
const handleLogin = async () => {
  if(!password.value) return message.error('请输入密码')
  loginLoading.value = true
  try {
    await adminApi.login(password.value)
    message.success('登录成功')
    isLoggedIn.value = true
    loadInitialData()
  } catch(e) { message.error('密码错误/登录失败') } 
  finally { loginLoading.value = false }
}

const handleLogout = async () => {
  try { await adminApi.logout() } catch(e) { console.error(e) }
  isLoggedIn.value = false
  password.value = ''
  message.info('已退出')
}

// Data Loading
const loadInitialData = async () => {
  try {
    const [statRes, catRes] = await Promise.all([adminApi.getStats(), categoryApi.getAll()])
    stats.value = statRes.data
    categories.value = catRes.data
  } catch(e) { console.error(e) }
}

const loadPending = async () => {
  const res = await adminApi.getPending()
  pendingImages.value = res.data.images || []
}

const loadAllImages = async () => {
  allImagesLoading.value = true
  try {
    const res = await adminApi.getImages({ page: imagePagination.current, limit: imagePagination.pageSize })
    allImages.value = res.data.images || []
    imagePagination.total = res.data.total
  } catch(e) { message.error('加载图片失败') }
  finally { allImagesLoading.value = false }
}

// Actions
const approveImage = async (record) => {
  if(!record.tempCategoryId) return message.warning('请选择分类')
  try {
    await adminApi.approve(record.id, record.tempCategoryId)
    message.success('已批准')
    loadPending(); loadInitialData()
  } catch(e) { message.error('操作失败') }
}

const rejectImage = (id) => {
  Modal.confirm({
    title: '确认拒绝并删除?',
    onOk: async () => {
      try { await adminApi.reject(id); message.success('已拒绝'); loadPending(); loadInitialData() } 
      catch(e) { message.error('失败') }
    }
  })
}

const deleteImage = (id) => {
  Modal.confirm({
    title: '确认永久删除?',
    content: '此操作不可恢复',
    okType: 'danger',
    onOk: async () => {
      try { await adminApi.deleteImage(id); message.success('已删除'); loadAllImages(); loadInitialData() }
      catch(e) { message.error('删除失败') }
    }
  })
}

// Category Management
const catModalVisible = ref(false)
const catForm = reactive({ id: null, name: '', slug: '' })
const isEditing = ref(false)

const openCatModal = (cat = null) => {
  isEditing.value = !!cat
  if(cat) { Object.assign(catForm, cat) } else { Object.assign(catForm, { id: null, name: '', slug: '' }) }
  catModalVisible.value = true
}

const saveCategory = async () => {
  try {
    if(isEditing.value) await categoryApi.update(catForm.id, catForm)
    else await categoryApi.create(catForm)
    message.success('保存成功')
    catModalVisible.value = false
    loadInitialData()
  } catch(e) { message.error('保存失败') }
}

const deleteCategory = (id) => {
  Modal.confirm({
    title: '确认删除分类?',
    content: '如果分类下有图片将无法删除',
    okType: 'danger',
    onOk: async () => {
      try { await categoryApi.delete(id); message.success('已删除'); loadInitialData() }
      catch(e) { message.error('删除失败: 可能含有图片') }
    }
  })
}

// Batch Operation State
const selectedRowKeys = ref([])
const bulkModalVisible = ref(false)
const bulkCategoryId = ref(null)

const onSelectChange = (keys) => {
  selectedRowKeys.value = keys
}

const handleBulkApprove = async () => {
  if (!bulkCategoryId.value) return message.warning('请选择分类')
  try {
    await adminApi.bulkApprove(selectedRowKeys.value, bulkCategoryId.value)
    message.success(`批量审批成功 (${selectedRowKeys.value.length} 张)`)
    selectedRowKeys.value = []
    bulkModalVisible.value = false
    loadPending(); loadInitialData()
  } catch (e) { message.error('批量审批失败') }
}

const handleBulkDelete = () => {
  Modal.confirm({
    title: `确认永久删除这 ${selectedRowKeys.value.length} 张图片?`,
    content: '此操作不可恢复',
    okType: 'danger',
    onOk: async () => {
      try {
        await adminApi.bulkDelete(selectedRowKeys.value)
        message.success('批量删除成功')
        selectedRowKeys.value = []
        if (activeKey.value === 'audit') loadPending()
        else loadAllImages()
        loadInitialData()
      } catch (e) { message.error('批量删除失败') }
    }
  })
}

onMounted(() => {
  // Check session
  adminApi.getStats().then(() => { isLoggedIn.value = true; loadInitialData() }).catch(() => {})
})

// Columns
const pendingColumns = [
  { title: '封面', key: 'cover', width: 120 },
  { title: '原名', dataIndex: 'original_name' },
  { title: '上传时间', dataIndex: 'created_at' },
  { title: '分类指定', key: 'category' },
  { title: '操作', key: 'action', width: 200 }
]

const imageColumns = [
  { title: 'ID', dataIndex: 'id', width: 80 },
  { title: '封面', key: 'cover', width: 100 },
  { title: '名称', dataIndex: 'original_name', ellipsis: true },
  { title: '状态', key: 'status', width: 100 },
  { title: '分类', key: 'category' },
  { title: '时间', dataIndex: 'created_at' },
  { title: '操作', key: 'action', width: 100 }
]

const catColumns = [
  { title: 'ID', dataIndex: 'id', width: 80 },
  { title: '名称', dataIndex: 'name' },
  { title: 'Slug', dataIndex: 'slug' },
  { title: '操作', key: 'action', width: 200 }
]
</script>

<template>
  <div class="h-screen bg-gray-50">
    <div v-if="!isLoggedIn" class="flex items-center justify-center h-full">
      <a-card title="管理员登录" style="width: 400px" :bordered="false" class="shadow-lg">
        <a-input-password v-model:value="password" placeholder="请输入密码" size="large" @pressEnter="handleLogin" />
        <a-button type="primary" block size="large" class="mt-4" :loading="loginLoading" @click="handleLogin">登录</a-button>
        <div class="mt-4 text-center">
           <router-link to="/">返回首页</router-link>
        </div>
      </a-card>
    </div>

    <a-layout v-else class="h-full">
      <a-layout-sider theme="light" width="240" class="shadow-md z-10">
        <div class="p-4 text-center border-b">
          <h1 class="text-xl font-bold m-0 tracking-tight">HYW ADMIN</h1>
        </div>
        <a-menu v-model:selectedKeys="selectedKeys" mode="inline" class="border-0 mt-2">
          <a-menu-item key="dashboard" @click="activeKey='dashboard'">
            <DashboardOutlined /> <span>仪表盘</span>
          </a-menu-item>
          <a-menu-item key="audit" @click="activeKey='audit'; loadPending()">
            <AuditOutlined /> <span>待审核</span>
            <a-badge :count="stats.pending_images" :offset="[10, 0]" />
          </a-menu-item>
          <a-menu-item key="images" @click="activeKey='images'; loadAllImages()">
            <FileImageOutlined /> <span>图片管理</span>
          </a-menu-item>
          <a-menu-item key="categories" @click="activeKey='categories'">
            <FolderOutlined /> <span>分类管理</span>
          </a-menu-item>
        </a-menu>
        <div class="absolute bottom-0 w-full p-4 border-t">
          <a-button danger block @click="handleLogout">
            <LogoutOutlined /> 退出登录
          </a-button>
        </div>
      </a-layout-sider>

      <a-layout-content class="p-6 overflow-auto">
        
        <!-- Dashboard -->
        <div v-if="activeKey === 'dashboard'">
          <h2 class="text-2xl font-bold mb-6">仪表盘</h2>
          <a-row :gutter="16">
            <a-col :span="6">
              <a-card>
                <a-statistic title="总图片数" :value="stats.total_images" />
              </a-card>
            </a-col>
            <a-col :span="6">
              <a-card>
                <a-statistic title="待审核" :value="stats.pending_images" valueStyle="color: #faad14" />
              </a-card>
            </a-col>
            <a-col :span="6">
              <a-card>
                <a-statistic title="已发布" :value="stats.approved_images" valueStyle="color: #52c41a" />
              </a-card>
            </a-col>
            <a-col :span="6">
              <a-card>
                <a-statistic title="活跃分类" :value="stats.total_categories" />
              </a-card>
            </a-col>
          </a-row>
        </div>

        <!-- Audit -->
        <div v-if="activeKey === 'audit'">
          <div class="flex justify-between items-center mb-4">
            <h2 class="text-2xl font-bold m-0">待审核图片</h2>
            <a-space>
              <a-button v-if="selectedRowKeys.length > 0" type="primary" @click="bulkModalVisible = true">
                批量批准 ({{ selectedRowKeys.length }})
              </a-button>
              <a-button v-if="selectedRowKeys.length > 0" danger @click="handleBulkDelete">
                批量拒绝/删除
              </a-button>
              <a-button @click="loadPending; selectedRowKeys = []">刷新</a-button>
            </a-space>
          </div>
          <a-table 
            :dataSource="pendingImages" 
            :columns="pendingColumns" 
            rowKey="id"
            :row-selection="{ selectedRowKeys: selectedRowKeys, onChange: onSelectChange }"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'cover'">
                <a-image :src="`/uploads/${record.filename}`" width="80px" height="60px" class="object-cover rounded" />
              </template>
              <template v-if="column.key === 'category'">
                 <a-select 
                    v-model:value="record.tempCategoryId" 
                    placeholder="选择分类" 
                    style="width: 150px"
                    :options="categories.map(c => ({ label: c.name, value: c.id }))"
                 />
              </template>
              <template v-if="column.key === 'action'">
                <a-space>
                  <a-button type="primary" size="small" @click="approveImage(record)">
                    <template #icon><CheckCircleOutlined /></template> 批准
                  </a-button>
                  <a-button danger size="small" @click="rejectImage(record.id)">
                     <template #icon><CloseCircleOutlined /></template> 拒绝
                  </a-button>
                </a-space>
              </template>
            </template>
          </a-table>
        </div>

        <!-- Image Management -->
        <div v-if="activeKey === 'images'">
          <div class="flex justify-between items-center mb-4">
            <h2 class="text-2xl font-bold m-0">图片管理</h2>
            <a-space>
              <a-button v-if="selectedRowKeys.length > 0" danger @click="handleBulkDelete">
                批量删除 ({{ selectedRowKeys.length }})
              </a-button>
              <a-button @click="loadAllImages; selectedRowKeys = []">刷新</a-button>
            </a-space>
          </div>
          <a-table 
             :dataSource="allImages" 
             :columns="imageColumns" 
             rowKey="id"
             :loading="allImagesLoading"
             :pagination="imagePagination"
             :row-selection="{ selectedRowKeys: selectedRowKeys, onChange: onSelectChange }"
             @change="(p) => { imagePagination.current = p.current; loadAllImages() }"
          >
            <template #bodyCell="{ column, record }">
               <template v-if="column.key === 'cover'">
                  <a-image :src="`/uploads/${record.filename}`" width="60px" height="40px" class="object-cover rounded" />
               </template>
               <template v-if="column.key === 'status'">
                 <a-tag :color="record.status === 'approved' ? 'green' : 'orange'">{{ record.status }}</a-tag>
               </template>
               <template v-if="column.key === 'category'">
                 {{ categories.find(c => c.id === record.category_id)?.name || '-' }}
               </template>
               <template v-if="column.key === 'action'">
                 <a-button type="text" danger @click="deleteImage(record.id)">删除</a-button>
               </template>
            </template>
          </a-table>
        </div>

        <!-- Categories -->
        <div v-if="activeKey === 'categories'">
           <div class="flex justify-between items-center mb-4">
            <h2 class="text-2xl font-bold m-0">分类管理</h2>
            <a-button type="primary" @click="openCatModal()">
              <PlusOutlined /> 新建分类
            </a-button>
          </div>
          <a-table :dataSource="categories" :columns="catColumns" rowKey="id">
             <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'action'">
                   <a-space>
                     <a-button type="link" size="small" @click="openCatModal(record)">编辑</a-button>
                     <a-button type="link" danger size="small" @click="deleteCategory(record.id)">删除</a-button>
                   </a-space>
                </template>
             </template>
          </a-table>
        </div>
      </a-layout-content>
    </a-layout>

    <!-- Category Modal -->
    <a-modal v-model:open="catModalVisible" :title="isEditing ? '编辑分类' : '新建分类'" @ok="saveCategory">
       <a-form layout="vertical">
          <a-form-item label="名称" required>
             <a-input v-model:value="catForm.name" placeholder="请输入分类名称" />
          </a-form-item>
          <a-form-item label="标识 (Slug)" required>
             <a-input v-model:value="catForm.slug" placeholder="例如: anime" />
          </a-form-item>
       </a-form>
    </a-modal>

    <!-- Bulk Approve Modal -->
    <a-modal v-model:open="bulkModalVisible" title="批量批准分类" @ok="handleBulkApprove">
       <p class="mb-4">将选中的 {{ selectedRowKeys.length }} 张图片移动到：</p>
       <a-select 
          v-model:value="bulkCategoryId" 
          placeholder="选择目标分类" 
          style="width: 100%"
          :options="categories.map(c => ({ label: c.name, value: c.id }))"
       />
    </a-modal>
  </div>
</template>
