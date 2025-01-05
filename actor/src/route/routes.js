import sun from "@/pages/Main.vue";
import moon from "@/pages/Admin.vue";
import {
    CogOutline,
    DocumentsOutline,
    HammerOutline,
    HappyOutline,
    PeopleOutline,
    ReceiptOutline
} from '@vicons/ionicons5'


let allTool = () => import("@/pages/manager/AllTool.vue")
let articlesManager = () => import("@/pages/manager/ArticlesManager.vue")
let pageManager = () => import("@/pages/manager/PageManager.vue")
let userManager = () => import("@/pages/manager/UserManager.vue")
let roleManager = () => import("@/pages/manager/RoleManager.vue")
let optLog = () => import("@/pages/manager/OptLog.vue")

let about = () => import("@/pages/home/AboutPage.vue")
let index = () => import("@/pages/home/IndexPage.vue")
let bbs = () => import("@/pages/home/BBSIndex.vue")
let bbsPage = () => import("@/pages/home/bbs/ArticlesList.vue")
let articlesPage = () => import("@/pages/home/bbs/ArticlesDetail.vue")
let articlesEdit = () => import("@/pages/home/bbs/ArticlesEdit.vue")
let userCenter = () => import("@/pages/home/UserCenter.vue")
let userInfo = () => import("@/pages/home/user/UserInfo.vue")
let notificationCenter = () => import("@/pages/home/NotificationCenter.vue")
let userEdit = () => import("@/pages/home/UserEdit.vue")
let regOrLogin = () => import("@/pages/home/RegOrLogin.vue")

export let managerRouter = {
    belongMenu: true,
    path: '/manager', component: moon, children: [
        {showName: '', path: '', component: allTool, belongMenu: false},
        {showName: '工具', path: 'allTool', component: allTool, belongMenu: true, icon: HammerOutline},
        {showName: '用户管理', path: 'userManager', component: userManager, belongMenu: true, icon: PeopleOutline},
        {showName: '角色管理', path: 'roleManager', component: roleManager, belongMenu: true, icon: HappyOutline},
        {
            showName: '文章管理',
            path: 'articlesManager',
            component: articlesManager,
            belongMenu: true,
            icon: DocumentsOutline
        },
        {showName: '页面管理', path: 'pageManager', component: pageManager, belongMenu: true, icon: CogOutline},
        {showName: '操作记录', path: 'optLog', component: optLog, belongMenu: true, icon: ReceiptOutline},
        {
            showName: "分类管理",
            path: 'category',
            name: 'categoryManager',
            component: () => import('@/pages/admin/CategoryManager.vue'),
            belongMenu: true, icon: ReceiptOutline
        }
    ]
}

export let routes = [
    {
        path: '/:catchAll(.*)*', name: '', redirect: '/home/'
    },
    {
        path: '/home', component: sun, children: [
            {name: '', path: '', redirect: '/home/bbs'},
            {name: 'index', path: 'index', component: index},
            {
                path: 'bbs', component: bbs, children: [
                    {name: '', path: '', redirect: '/home/bbs/bbs'},
                    {name: 'bbs', path: 'bbs', component: bbsPage},
                    {name: 'articlesPage', path: 'articlesPage', component: articlesPage},
                    {name: 'articlesEdit', path: 'articlesEdit', component: articlesEdit},
                ]
            },
            {name: 'about', path: 'about', component: about},
            {name: 'regOrLogin', path: 'regOrLogin', component: regOrLogin},
            {
                name: 'userCenter', path: 'userCenter', component: userCenter, children: [
                    {name: 'userInfo', path: 'userInfo', component: userInfo},
                ]
            },
            {
                name: 'notificationCenter', path: 'notificationCenter', component: notificationCenter
            },
            {
                name: 'userEdit', path: 'userEdit', component: userEdit
            }
        ]
    },
    managerRouter,
]
