import {ref, watch} from "vue"
import {defineStore} from "pinia";

export const useUserStore = defineStore('user', () => {

    const userInfo = ref({username: '', userId: 0})
    let userInfoData = window.localStorage.getItem('userInfo') || ""

    if (userInfoData !== "") {
        let pData = JSON.parse(userInfoData)
        if (pData) {
            userInfo.value.username = pData.username || ""
            userInfo.value.userId = pData.userId || ""
        }
    }
    const token = ref(window.localStorage.getItem('token') || "")

    function saveUserInfo() {
        window.localStorage.setItem('userInfo', JSON.stringify(userInfo.value))
    }

    function login(userData) {
        console.log(userData)
        userInfo.value.username = userData.username
        userInfo.value.userId = userData.userId
        token.value = userData.token
        saveUserInfo()
    }

    function clearUserInfo() {
        userInfo.value = {username: '', userId: 0}
        token.value = ''
        sessionStorage.clear()
        localStorage.clear()
    }

    watch(() => token.value, () => {
        window.localStorage.setItem('token', token.value)
    })

    return {
        userInfo,
        token,
        login,
        clearUserInfo
    }
})
