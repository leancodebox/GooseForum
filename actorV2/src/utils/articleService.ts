import axiosInstance from './axiosInstance';

// 定义文章类型
interface Article {
    id: number,
    title: string;
    type: string;
    categories: number[];
    content: string;
}

export const getArticleEnum = async () => {
    try {
        return await axiosInstance.get('bbs/get-articles-enum')
    } catch (error) {
        throw new Error(`提交文章失败: ${error}`);
    }
}

export const getArticlesOrigin = async (id: any) => {
    try {
        return await axiosInstance.post('/bbs/get-articles-origin', {
            id: parseInt(id)
        })
    } catch (error) {
        throw new Error(`提交文章失败: ${error}`);
    }
}

// 提交文章的函数
export const submitArticle = async (article: Article) => {
    try {
        return await axiosInstance.post('/bbs/write-articles', article); // 返回响应数据
    } catch (error) {
        throw new Error(`提交文章失败: ${error}`);
    }
};

// Mock 获取用户信息
export const getUserInfo = async () => {
    // 模拟用户数据
    return new Promise((resolve) => {
        setTimeout(() => {
            resolve({
                avatarUrl: '/file/img/avatars/avatar_1_1736935027/a93ee576-ff66-442a-8def-ac52e56872a3.png',
                bio: '',
                email: 'abandon1a2b@outlook.com',
                isAdmin: true,
                nickname: '昵称昵称妮妮称',
                signature: '',
                userId: 1,
                username: 'abandon',
                website: ''
            });
        }, 1); // 模拟网络延迟
    });

    try {
        return axiosInstance.get("/get-user-info")
    } catch (error) {
    }
};
