import {createApp} from 'vue'
import App from './App.vue'
import router from './route/router'
import {createPinia} from 'pinia'

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

