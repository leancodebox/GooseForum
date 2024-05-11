import {h, resolveComponent} from "vue";
import {LogoTwitter, LogoWebComponent} from '@vicons/ionicons5'

import sun from "@/pages/HomePage.vue";
import moon from "@/pages/AllManager.vue";


let allTool = () => import("@/pages/manager/AllTool.vue")
let articlesManager = () => import("@/pages/manager/ArticlesManager.vue")
let pageManager = () => import("@/pages/manager/PageManager.vue")
let userManager = () => import("@/pages/manager/UserManager.vue")

let about = () => import("@/pages/home/AboutPage.vue")
let index = () => import("@/pages/home/IndexPage.vue")
let login = () => import("@/pages/Login.vue")
let bbs = () => import("@/pages/home/bbs/BBSIndex.vue")
let bbsPage = () => import("@/pages/home/bbs/BBSPage.vue")
let articlesPage = () => import("@/pages/home/bbs/ArticlesPage.vue")
let articlesEdit = () => import("@/pages/home/bbs/ArticlesEdit.vue")

export let managerRouter = {
    belongMenu: true,
    path: '/manager', component: moon, children: [
        {showName: '', path: '', component: allTool, belongMenu: false},
        {showName: '工具', path: 'allTool', component: allTool, belongMenu: true},
        {showName: '用户管理', path: 'articlesManager', component: articlesManager, belongMenu: true},
        {showName: '文章管理', path: 'pageManager', component: pageManager, belongMenu: true},
        {showName: '页面管理', path: 'userManager', component: userManager, belongMenu: true},
    ]
}

export let routes= [
    {
        path: '/:catchAll(.*)*', name: '', redirect: '/home/'
    },
    {
        path: '/login', component: login
    },
    {
        path: '/home', component: sun, children: [
            {name: '', path: '', redirect: '/home/index'},
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

        ]
    },
    managerRouter
]
