import axiosInstance from './axiosInstance';

// 定义文章类型
interface Article {
    id: number,
    title: string;
    type: string;
    categories: string[];
    content: string;
}

// 提交文章的函数
export const submitArticle = async (article: Article) => {
    try {
        return await axiosInstance.post('/api/bbs/write-articles', article); // 返回响应数据
    } catch (error) {
        throw new Error(`提交文章失败: ${error}`);
    }
};

export const getUserInfo = async () => {
    try {
        return axiosInstance.get("/api/get-user-info")
    } catch (error) {
    }
}
