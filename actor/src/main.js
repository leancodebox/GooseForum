import {createApp} from 'vue'
import App from './App.vue'
import router from './route/router'
import {createPinia} from 'pinia'
// !import 不能忘记引入katex的样式
// import 'katex/dist/katex.css'
// 引入katex下的自动渲染函数
// import renderMathInElement from 'katex/contrib/auto-render/auto-render'

// 定义自动渲染的配置参数,这些参数根据你的需求进行修改，下面的参数是官网上抄下来的
const renderOption = {
    delimiters: [
        {left: '$$', right: '$$', display: true},
        {left: '$', right: '$', display: false},
        {left: '\\(', right: '\\)', display: false},
        {left: '\\[', right: '\\]', display: true}
    ],
    throwOnError: false
}
// let title;  // 用于临时存放原来的title内容
// window.onblur = function(){
//     // onblur时先存下原来的title,再更改title内容
//     title = document.title;
//     document.title = title+"🤔";
// };
// window.onfocus = function () {
//     // onfocus时原来的title不为空才替换回去
//     // 防止页面还没加载完成且onblur时title=undefined的情况
//     if(title) {
//         document.title = title;
//     }
// }

const app = createApp(App)

app.use(router)
    .use(createPinia())
    .mount('#app')

