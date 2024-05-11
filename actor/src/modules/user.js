import {ref, watch} from "vue"
import {defineStore} from "pinia";

export const useUserStore = defineStore('user', () => {
    let userInfoData = window.localStorage.getItem('userInfo')
    const userInfo = ref({username: ''})
    if (userInfoData !== undefined) {
        let pData = JSON.parse(userInfoData)
        if (pData) {
            userInfo.username = pData.username || ""
        }
    }
    const token = ref(window.localStorage.getItem('token') || "")

    function login(userData) {
        userInfo.value.username = userData.username
        token.value = userData.token
    }

    function layout() {
        userInfo.value = null
        token.value = ''
        sessionStorage.clear()
        localStorage.clear()
    }

    watch(() => token.value, () => {
        window.localStorage.setItem('token', token.value)
    })

    watch(() => userInfo.value, () => {
        userInfo.value = JSON.stringify(userInfo.value)
    })

    return {
        userInfo,
        token,
        login
    }
})
