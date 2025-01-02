import {createRouter, createWebHashHistory} from "vue-router";
import {routes} from "./routes"
import {useUserStore} from "@/modules/user"

const router = createRouter({
    // 4. 内部提供了 history 模式的实现。为了简单起见，我们在这里使用 hash 模式。
    history: createWebHashHistory(),
    routes: routes, // `routes: routes` 的缩写
})

// 需要登录才能访问的路由
const authRoutes = [
    '/home/userCenter',
    '/home/notificationCenter',
    '/home/userEdit',
    '/home/bbs/articlesEdit',
    '/manager'
]

router.beforeEach((to, from, next) => {
    const userStore = useUserStore()
    
    // 检查是否需要登录
    if (authRoutes.some(path => to.path.startsWith(path)) && !userStore.token) {
        next({
            path: '/home/regOrLogin',
            query: { redirect: to.fullPath }
        })
    } else {
        next()
    }
})

export default router