import axiosInstance from './axiosInstance';
import {enqueueMessage} from "@/utils/messageManager.ts";
import axios from 'axios';
import type {Result} from "@/types/articleInterfaces.ts";


// Mock 获取用户信息
export const getUserInfo = async () => {
    return axiosInstance.get("/get-user-info")
}

