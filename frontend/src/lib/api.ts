import axios from 'axios'

export const api = axios.create({
    baseURL: '/api',  // Use relative path for Vite proxy
    headers: { 'Content-Type': 'application/json' },
    withCredentials: true,
    timeout: 10000,
})

// Request interceptor - attach token to every request if present
api.interceptors.request.use((config) => {
    const token = localStorage.getItem('token')

    console.log(`🌐 API Request: ${config.method?.toUpperCase()} ${config.url}`);
    console.log(`   Token present: ${!!token}`);

    if (token) {
        config.headers.Authorization = `Bearer ${token}`
        console.log(`   Authorization header set`);
    }

    return config
}, (error) => {
    console.error("❌ Request interceptor error:", error);
    return Promise.reject(error);
})

// Response interceptor - handle errors
api.interceptors.response.use(
    (response) => {
        console.log(`✅ API Response: ${response.config.url} - ${response.status}`);
        return response
    },
    (error) => {
        console.error(`❌ API Error: ${error.config?.url}`);
        console.error(`   Status: ${error.response?.status}`);
        console.error(`   Message: ${error.message}`);

        if (error.response?.data) {
            console.error(`   Data:`, error.response.data);
        }

        if (error.response?.status === 401) {
            console.warn("⚠️ Unauthorized - clearing token and redirecting to login");
            localStorage.removeItem('token')
            localStorage.removeItem('email')
            if (typeof window !== 'undefined') {
                window.location.href = '/login'
            }
        }

        return Promise.reject(error)
    }
)

// Types
export interface Workspace {
    id: string
    name: string
    owner_id: string
    created_at: string
}

export interface Page {
    id: string
    workspace_id: string
    parent_id: string | null
    title: string
    icon: string | null
    cover: string | null
    created_by: string
    created_at: string
    updated_at: string
}