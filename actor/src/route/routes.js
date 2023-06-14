import {h, resolveComponent} from "vue";
import {LogoTwitter, LogoWebComponent} from '@vicons/ionicons5'

import sun from "@/pages/HomePage.vue";
import moon from "@/pages/AllManager.vue";

let about = () => import("@/pages/home/AboutPage.vue")
let allTool = () => import("@/pages/manager/AllTool.vue")
let index = () => import("@/pages/home/IndexPage.vue")
let sysInfo = () => import("@/pages/manager/SysInfo.vue")
let login = () => import("@/pages/Login.vue")


let bbs = () => import("@/pages/home/bbs/BBSIndex.vue")
let bbsPage = () => import("@/pages/home/bbs/BBSPage.vue")
let articlesPage = () => import("@/pages/home/bbs/ArticlesPage.vue")
let articlesEdit = () => import("@/pages/home/bbs/ArticlesEdit.vue")


export default [
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
    {
        belongMenu: true,
        path: '/manager', component: moon, children: [
            {showName: '', path: '', component: allTool, belongMenu: false},
            {showName: 'all tool', path: 'allTool', component: allTool, belongMenu: true},
            {showName: 'sysInfo', path: 'sysInfo', component: sysInfo, belongMenu: true},
        ]
    },
]