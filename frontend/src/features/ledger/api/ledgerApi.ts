import { apiClient } from "../../../api/client"

// === DTO 定义 (对应后端的 PostTransactionReq) ===

export interface Posting {
    account_code: string
    direction: "D" | "C"
    amount: string // ⚠️ 金额必须是字符串！
}

export interface PostTransactionReq {
    reference_id: string
    tx_type: string
    description: string
    postings: Posting[]
}

export interface PostTransactionRes {
    tx_id: number
    reference_id: string
    posted_at: string
    message: string
}

export interface Account {
    ID: number
    AccountCode: string
    Name: string
    Type: number
    Currency: string
    Balance: string // 后端 decimal 序列化为 string
    Status: number
}

// === API 方法 ===

export const ledgerApi = {
    // 获取账户列表
    getAccounts: (): Promise<Account[]> => {
        return apiClient.get("/ledger/accounts")
    },

    // 记账接口
    postTransaction: (
        data: PostTransactionReq
    ): Promise<PostTransactionRes> => {
        return apiClient.post("/ledger/transactions", data)
    },

    // (预留) 获取健康状态
    getHealth: () => {
        return apiClient.get("/health")
    },
}
