import axiosInstance from './axiosInstance.ts';
import type {
    Result,
} from './adminInterfaces.ts';

// Mock 获取用户信息
export const getUserInfo = async (): Promise<Result<any>> => {
    return axiosInstance.get('api/get-user-info')
}