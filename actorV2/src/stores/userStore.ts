import { defineStore } from 'pinia';
import { ref } from 'vue';
import { getUserInfo } from '@/utils/articleService'; // 引入获取用户信息的接口

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

export const useUserStore = defineStore('user', () => {
    const userInfo = ref<UserInfo | null>(null); // 设置 userInfo 的类型

    const fetchUserInfo = async () => {
        try {
            userInfo.value = await getUserInfo(); // 这里会调用 Mock 数据
        } catch (error) {
            console.error('获取用户信息失败:', error);
        }
    };

    return { userInfo, fetchUserInfo };
});
