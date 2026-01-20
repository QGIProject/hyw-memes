import axios from 'axios'

const api = axios.create({
    baseURL: '/api',
    headers: {
        'Content-Type': 'application/json'
    }
})

// Add auth token to requests
api.interceptors.request.use(config => {
    const token = localStorage.getItem('token')
    if (token) {
        config.headers.Authorization = `Bearer ${token}`
    }
    return config
})

// Auth API
export const authApi = {
    register: (username, password) => api.post('/auth/register', { username, password }),
    login: (username, password) => api.post('/auth/login', { username, password }),
    getMe: () => api.get('/auth/me')
}

// Category API
export const categoryApi = {
    getAll: () => api.get('/categories'),
    create: (data) => api.post('/admin/categories', data, { headers: { 'X-Admin-Token': 'admin' } }),
    update: (id, data) => api.put(`/admin/categories/${id}`, data, { headers: { 'X-Admin-Token': 'admin' } }),
    delete: (id) => api.delete(`/admin/categories/${id}`, { headers: { 'X-Admin-Token': 'admin' } })
}

// Image API
export const imageApi = {
    getApproved: (page = 1, limit = 20, categoryId = '') => {
        let url = `/images?page=${page}&limit=${limit}`
        if (categoryId) url += `&category_id=${categoryId}`
        return api.get(url)
    },
    getRandom: (categoryId = '') => {
        let url = '/images/random'
        if (categoryId) url += `?category_id=${categoryId}`
        return api.get(url)
    },
    upload: (files, categoryId = '') => {
        const formData = new FormData()
        // Handle both single file and array of files
        if (Array.isArray(files)) {
            files.forEach(file => formData.append('images', file))
        } else {
            formData.append('images', files)
        }

        if (categoryId) formData.append('category_id', categoryId)
        return api.post('/images/upload', formData, {
            headers: { 'Content-Type': 'multipart/form-data' }
        })
    }
}

// Admin API
export const adminApi = {
    login: (password) => api.post('/admin/login', { password }),
    logout: () => api.post('/admin/logout'),
    getPending: () => api.get('/admin/pending', { headers: { 'X-Admin-Token': 'admin' } }),
    getImages: (params) => api.get('/admin/images', { params, headers: { 'X-Admin-Token': 'admin' } }),
    approve: (id, categoryId) => api.post(`/admin/approve/${id}`, { category_id: categoryId }, { headers: { 'X-Admin-Token': 'admin' } }),
    reject: (id) => api.post(`/admin/reject/${id}`, {}, { headers: { 'X-Admin-Token': 'admin' } }),
    deleteImage: (id) => api.delete(`/admin/images/${id}`, { headers: { 'X-Admin-Token': 'admin' } }),
    bulkApprove: (ids, categoryId) => api.post('/admin/bulk-approve', { ids, category_id: categoryId }, { headers: { 'X-Admin-Token': 'admin' } }),
    bulkDelete: (ids) => api.post('/admin/bulk-delete', { ids }, { headers: { 'X-Admin-Token': 'admin' } }),
    getStats: () => api.get('/admin/stats', { headers: { 'X-Admin-Token': 'admin' } })
}

export default api
