import axios from 'axios';
import { message } from 'antd';

// 创建 axios 实例
export const apiClient = axios.create({
  // Vite 在 vite.config.ts 里配置了代理，所以这里直接写 /api
  baseURL: '/api/v1', 
  timeout: 10000, // 10秒超时
  headers: {
    'Content-Type': 'application/json',
  },
});

// 响应拦截器：统一处理错误
apiClient.interceptors.response.use(
  (response) => {
    return response.data; // 直接返回 data 部分，少写一层 .data
  },
  (error) => {
    // 如果后端返回了错误信息，优先显示后端的 message
    const errorMsg = error.response?.data?.error || 'Network Error or Server Down';
    
    // 使用 AntD 的 Message 组件弹出全局错误提示
    message.error(errorMsg);
    
    return Promise.reject(error);
  }
);