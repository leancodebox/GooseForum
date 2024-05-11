import {login as loginApi,reg as regApi} from "@/service/request";
import {ref, watch} from "vue"
import {defineStore} from "pinia";
import router from "@/route/router"

export const useUserStore = defineStore('user', () => {
    const userInfo = ref({
        uuid: '',
        nickName: '',
    })
    const token = ref(window.localStorage.getItem('token') || '')

    function login(username, password) {
        loginApi(username, password).then(r => {
            token.value = r.result.token
            router.push({name: 'bbs', replace: true})
        })
    }

    function reg(username, password, email) {
        regApi(email, username, password).then(r => {
            token.value = r.result.token
            router.push({name: 'bbs', replace: true})
        })
    }

    function layout() {
        token.value = ''
        sessionStorage.clear()
        localStorage.clear()
    }

    watch(() => token.value, () => {
        window.localStorage.setItem('token', token.value)
    })
    function updateToken(newToken){
        token.value = newToken
    }

    return {
        userInfo,
        token,
        login,
        updateToken
    }
})
