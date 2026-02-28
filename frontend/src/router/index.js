import { createRouter, createWebHistory } from 'vue-router'

// ==============================
// 1. 前台页面组件
// ==============================
import Home from '@/views/Home.vue'
import ProductDetail from '@/views/ProductDetail.vue'
import Cart from '@/views/Cart.vue'
import Search from '@/views/Search.vue'
import UserOrders from '@/views/UserOrders.vue'
import UserProfile from '@/views/UserProfile.vue'
import MessageList from '@/views/MessageList.vue'
import ChatRoom from '@/views/ChatRoom.vue'
import Publish from '@/views/Publish.vue'
import MockPay from '@/views/MockPay.vue'

// ★★★ 新增引入：我卖出的页面 ★★★
import UserSales from '@/views/UserSales.vue'

// ★★★ 卖家专属：商品管理页 ★★★
const ProductManage = () => import('@/views/ProductManage.vue')

// ==============================
// 2. 后台管理组件 (懒加载)
// ==============================
const AdminLogin = () => import('@/views/admin/Login.vue')
const AdminLayout = () => import('@/views/admin/Layout.vue')
const AdminDashboard = () => import('@/views/admin/Dashboard.vue')
const AdminUserManagement = () => import('@/views/admin/UserManagement.vue')
const AdminProductAudit = () => import('@/views/admin/ProductAudit.vue')
const AdminOrderManagement = () => import('@/views/admin/OrderManagement.vue')
const AdminAIModelManagement = () => import('@/views/admin/AIModelManagement.vue')
const AdminAIModelDashboard = () => import('@/views/admin/AIModelDashboard.vue')

const routes = [
    // --------------------------
    // 前台商城路由
    // --------------------------
    {
        path: '/',
        name: 'Home',
        component: Home
    },
    {
        path: '/search',
        name: 'Search',
        component: Search
    },

    // 发布页 (需要登录)
    {
        path: '/publish',
        name: 'Publish',
        component: Publish,
        meta: { requiresAuth: true }
    },
    {
        path: '/pay/mock',
        name: 'MockPay',
        component: MockPay,
        meta: { requiresAuth: true }
    },

    // 需要登录的前台页
    {
        path: '/cart',
        name: 'Cart',
        component: Cart,
        meta: { requiresAuth: true }
    },
    {
        path: '/orders',
        name: 'UserOrders',
        component: UserOrders,
        meta: { requiresAuth: true }
    },

    // ★★★ 新增路由：我卖出的 ★★★
    {
        path: '/mysales',
        name: 'UserSales',
        component: UserSales,
        meta: { requiresAuth: true }
    },

    {
        path: '/user/:id',
        name: 'OtherUserProfile',
        component: UserProfile,
        meta: { requiresAuth: true }
    },
    {
        path: '/profile',
        name: 'UserProfile',
        component: UserProfile,
        meta: { requiresAuth: true }
    },
    {
        path: '/messages',
        name: 'MessageList',
        component: MessageList,
        meta: { requiresAuth: true }
    },
    // 在 routes 数组中添加：
    {
        path: '/chat/:id',
        name: 'ChatRoom',
        component: () => import('../views/ChatRoom.vue'),
        meta: { requiresAuth: true }
    },

    // 动态路由
    {
        path: '/product/:id',
        name: 'ProductDetail',
        component: ProductDetail
    },
    {
        path: '/product/manage/:id',
        name: 'ProductManage',
        component: ProductManage,
        meta: { requiresAuth: true }
    },

    // --------------------------
    // 后台管理路由
    // --------------------------
    {
        path: '/admin/login',
        name: 'AdminLogin',
        component: AdminLogin
    },
    {
        path: '/admin',
        component: AdminLayout,
        redirect: '/admin/dashboard',
        beforeEnter: (to, from, next) => {
            const adminToken = localStorage.getItem('admin_token')
            if (!adminToken) {
                next('/admin/login')
            } else {
                next()
            }
        },
        children: [
            {
                path: 'dashboard',
                name: 'AdminDashboard',
                component: AdminDashboard
            },
            {
                path: 'users',
                name: 'AdminUsers',
                component: AdminUserManagement
            },
            {
                path: 'products',
                name: 'AdminProducts',
                component: AdminProductAudit
            },
            {
                path: 'orders',
                name: 'AdminOrders',
                component: AdminOrderManagement
            },
            {
                path: 'ai-models',
                name: 'AdminAIModels',
                component: AdminAIModelManagement
            },
            {
                path: 'ai-dashboard',
                name: 'AdminAIDashboard',
                component: AdminAIModelDashboard
            },
        ]
    }
]

const router = createRouter({
    history: createWebHistory(),
    routes
})

// 全局路由守卫
router.beforeEach((to, from, next) => {
    if (to.meta.requiresAuth) {
        const token = localStorage.getItem('token')
        if (!token) {
            next({
                path: '/',
                query: {
                    auth: 'login',
                    redirect: to.fullPath
                }
            })
        } else {
            next()
        }
    } else {
        next()
    }
})

export default router
