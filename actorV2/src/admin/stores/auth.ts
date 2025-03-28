import {defineStore} from 'pinia'
import {ref} from 'vue'
import {getUserInfo} from "../utils/authService.ts";
import type {Result} from "../types/adminInterfaces.ts";

// 定义用户信息的接口
interface UserInfo {
    avatarUrl: string;
    bio: string;
    email: string;
    isAdmin: boolean;
    nickname: string;
    signature: string;
    userId: number;
    username: string;
    website: string;
}


export const useUserStore = defineStore('auth', () => {
    const userInfo = ref<UserInfo | null>(null); // 设置 userInfo 的类型
    const fetchUserInfo = async () => {
        try {
            // 使用类型断言，确保返回值符合 UserInfo 接口
            let res = await getUserInfo() as unknown as Result<UserInfo>; // 这里会调用 Mock 数据
            userInfo.value = res.result
        } catch (error) {
            console.error('获取用户信息失败:', error);
        }
    };

    const handleLogout = async () => {
        try {
            const response = await fetch('/logout', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                }
            });
            const data = await response.json();
            window.location.href = '/';
        } catch (error) {
            console.error('退出失败:', error);
        }
    }
    return {userInfo, fetchUserInfo, handleLogout};
});
