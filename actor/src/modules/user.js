import {ref, watch, computed} from "vue"
import {defineStore} from "pinia";

export const useUserStore = defineStore('user', () => {

    const userInfo = ref({
        username: '',
        userId: 0,
        avatarUrl: '',
        isAdmin: false,
    })
    let userInfoData = window.localStorage.getItem('userInfo') || ""

    if (userInfoData !== "") {
        let pData = JSON.parse(userInfoData)
        if (pData) {
            userInfo.value.username = pData.username || ""
            userInfo.value.userId = pData.userId || ""
            userInfo.value.avatarUrl = pData.avatarUrl || ""
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
        userInfo.value.avatarUrl = userData.avatarUrl || ""
        token.value = userData.token
        saveUserInfo()
    }

    function clearUserInfo() {
        userInfo.value = {
            username: '',
            userId: 0,
            avatarUrl: ''
        }
        token.value = ''
        sessionStorage.clear()
        localStorage.clear()
    }

    watch(() => token.value, () => {
        window.localStorage.setItem('token', token.value)
    })

    const isLogin = computed(() => {
        return token.value !== '' && userInfo.value.userId !== 0
    })

    return {
        userInfo,
        token,
        login,
        clearUserInfo,
        isLogin
    }
})
