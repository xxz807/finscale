import { apiClient } from '../../../api/client';

// === DTO 定义 (对应后端的 PostTransactionReq) ===

export interface Posting {
  account_code: string;
  direction: 'D' | 'C';
  amount: string; // ⚠️ 金额必须是字符串！
}

export interface PostTransactionReq {
  reference_id: string;
  tx_type: string;
  description: string;
  postings: Posting[];
}

export interface PostTransactionRes {
  tx_id: number;
  reference_id: string;
  posted_at: string;
  message: string;
}

// === API 方法 ===

export const ledgerApi = {
  // 记账接口
  postTransaction: (data: PostTransactionReq): Promise<PostTransactionRes> => {
    return apiClient.post('/ledger/transactions', data);
  },
  
  // (预留) 获取健康状态
  getHealth: () => {
    return apiClient.get('/health');
  }
};