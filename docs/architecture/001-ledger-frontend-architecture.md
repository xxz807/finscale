
---

# 🖥️ FinScale Frontend Architecture Design

| 文档信息 | 内容 |
| :--- | :--- |
| **项目** | FinScale (Financial Scalable Core) |
| **模块** | **Console (管理控制台)** |
| **技术栈** | React 18 + TypeScript + Vite + Ant Design 5 |
| **架构风格** | Component-Driven + Server State Management |
| **版本** | v1.0.0 |
| **最后更新** | 2025-12-13 |

---

## 1. 架构概览 (Overview)

前端不仅仅是后端的“展示层”，它是用户的**驾驶舱 (Cockpit)**。
FinScale 的前端架构核心理念是：**“数据精确，响应迅速，交互严谨”**。

### 1.1 核心设计模式
1.  **Server State First (服务端状态优先)**:
    *   放弃传统的 Redux 全局一把抓。
    *   使用 **TanStack Query (React Query)** 管理所有 API 数据（缓存、同步、重试）。
    *   **理由**：金融数据实时性要求高，客户端不应持有过期的“陈旧数据”。
2.  **Strict Typing (严格类型契约)**:
    *   前端 TypeScript 类型定义 (`interface`) 必须与后端 DTO 保持 1:1 一致。
    *   **禁止使用 `any`**，特别是在涉及金额 (`amount`) 的字段上。
3.  **Modular Structure (模块化结构)**:
    *   采用 **Feature-based** 目录结构，与后端 `ledger`, `payment` 模块对应，实现“全栈模块化”。

---

## 2. 技术选型 (Tech Stack Strategy)

| 领域 | 选型 | 决策理由 (Decision Record) |
| :--- | :--- | :--- |
| **Core** | **React 18** | 社区标准，Hooks 模式适合逻辑复用。 |
| **Build Tool** | **Vite** | 极速冷启动，提升开发体验 (DX)。 |
| **Language** | **TypeScript 5+** | 金融系统必须强类型。 |
| **UI Framework** | **Ant Design 5.x** | **关键决策**：蚂蚁金服出品，内置了最专业的金融级 Form/Table/InputNumber 组件，开箱即用。 |
| **Data Fetching**| **TanStack Query v5** | 自动处理 Loading/Error/Caching 状态，极大减少模板代码。 |
| **State Mgmt** | **Zustand** | (备用) 处理纯客户端状态（如侧边栏收缩、主题切换），比 Redux 轻量。 |
| **HTTP Client** | **Axios** | 拦截器 (Interceptors) 统一处理 Token 和全局错误提示。 |
| **Formatting** | **bignumber.js** | **关键决策**：前端禁止进行金额加减运算，但在展示时需处理精度，原生 JS Math 不可靠。 |
| **Styling** | **Tailwind CSS** | 用于微调布局 (Margin/Padding)，比写 CSS Modules 快。 |

---

## 3. 目录结构设计 (Project Structure)

我们要构建一个可扩展的目录结构，避免将所有组件堆在 `components` 文件夹里。

```text
frontend/
├── src/
│   ├── api/                  # [HTTP Layer] Axios 实例与拦截器
│   │   └── client.ts
│   │
│   ├── assets/               # 静态资源
│   │
│   ├── components/           # [Shared UI] 全局通用组件
│   │   ├── Layout/           # 侧边栏, 顶部导航
│   │   └── Guard/            # 权限守卫
│   │
│   ├── features/             # === [核心业务模块] ===
│   │   │                     # 对应后端的 internal/ledger
│   │   ├── ledger/           
│   │   │   ├── api/          # 该模块的 API 定义 (Endpoints)
│   │   │   ├── components/   # 模块私有组件 (e.g., LedgerForm)
│   │   │   ├── hooks/        # React Query Hooks (usePostTransaction)
│   │   │   ├── types/        # TypeScript Interfaces (DTOs)
│   │   │   └── routes.tsx    # 模块路由定义
│   │   │
│   │   └── auth/             # 用户认证模块
│   │
│   ├── hooks/                # 全局通用 Hooks
│   ├── stores/               # 全局状态 (Zustand)
│   ├── utils/                # 工具函数 (MoneyFormatter)
│   ├── App.tsx
│   └── main.tsx
│
├── .env                      # 环境变量 (VITE_API_URL)
└── tsconfig.json
```

---

## 4. 关键架构实现细节 (Implementation Details)

### 4.1 数据流与 API 处理 (The Data Flow)

我们不直接在组件里写 `axios.get()`。我们封装 **Custom Hooks**。

**示例：提交记账请求**

```typescript
// features/ledger/hooks/usePostTransaction.ts
import { useMutation } from '@tanstack/react-query';
import { message } from 'antd';
import { postTransaction } from '../api/ledgerApi'; // Axios wrapper

export const usePostTransaction = () => {
  return useMutation({
    mutationFn: postTransaction,
    onSuccess: (data) => {
      message.success(`Transaction Posted! ID: ${data.tx_id}`);
      // 自动刷新余额查询缓存
      queryClient.invalidateQueries({ queryKey: ['accounts'] });
    },
    onError: (error: any) => {
      // 统一错误处理，解析后端 ErrorCode
      message.error(`Failed: ${error.response?.data?.message || 'Unknown error'}`);
    },
  });
};
```

**在组件中使用：**
```tsx
const { mutate, isPending } = usePostTransaction();
// 点击按钮时只需调用 mutate(payload)
```

### 4.2 金额精度处理 (The Precision Rule)

前端是金额展示的“最后一公里”，必须极其小心。

1.  **接收 (Input)**: 后端传来的金额字段（如余额）必须是 **String** 类型。
2.  **展示 (Display)**: 使用 `Intl.NumberFormat` 或 `accounting.js` 进行千分位格式化。
3.  **计算 (Calc)**: **原则上前端不进行任何金额计算**。如果非算不可（如前端预估手续费），必须使用 `bignumber.js`。

**Utils 示例**:
```typescript
// utils/money.ts
export const formatCurrency = (amount: string | number, currency = 'CNY') => {
  // 即使是 "1000.0000" 这种字符串也能被正确处理
  return new Intl.NumberFormat('zh-CN', {
    style: 'currency',
    currency: currency,
    minimumFractionDigits: 2
  }).format(Number(amount));
};
```

### 4.3 交互体验设计 (UX for Finance)

1.  **防重复提交 (Double Submit Protection)**:
    *   所有提交按钮（Button）必须绑定 `loading` 状态。
    *   这是防止运维人员手抖导致重复记账的第一道（前端）防线。
2.  **操作确认 (Confirmation)**:
    *   涉及资金变动的操作，必须弹出 `Modal.confirm` 二次确认。
3.  **大数字脱敏/高亮**:
    *   负数余额必须**标红**。
    *   大额数字（>100万）可以加粗显示。

---

## 5. 类型同步策略 (Type Sync)

为了保证前后端契约一致，建议手动（MVP阶段）或自动（进阶）同步类型。

**后端 DTO (`backend/internal/ledger/api/dto.go`)**:
```go
type PostTxRequest struct {
    RefID string `json:"reference_id"`
    // ...
}
```

**前端 Interface (`frontend/src/features/ledger/types/index.ts`)**:
```typescript
export interface PostTxRequest {
    reference_id: string;
    tx_type: string;
    description: string;
    postings: Posting[];
}

export interface Posting {
    account_code: string;
    amount: string; // ⚠️ 注意：后端是 string，这里必须也是 string
    direction: 'D' | 'C';
}
```

---

## 6. 安全性设计 (Security)

1.  **XSS 防护**: React 默认转义所有输出，但在使用 `dangerouslySetInnerHTML` 时需谨慎。
2.  **Token 存储**:
    *   建议将 JWT 存储在 `localStorage` (MVP) 或 `HttpOnly Cookie` (Production)。
    *   Axios Interceptor 负责每次请求自动附带 `Authorization: Bearer ...` 头。
3.  **路由守卫**:
    *   使用 React Router 的 Loader 或 Wrapper 组件，未登录直接踢回 Login 页。

---

## 7. 开发规范 (Dev Standards)

*   **Linting**: ESLint + Prettier (在保存时自动格式化)。
*   **Git Hooks**: 使用 `husky`，在 commit 前检查类型错误 (`tsc --noEmit`)。不允许带报错的代码提交。

---

### 📝 架构师备注 (Architect's Notes)

> "The UI is a function of State."
> (界面是状态的函数。)

在开发 Dashboard 时，不要手动去操作 DOM（比如 `document.getElementById('balance').innerText = ...`）。
**永远修改数据（State/Cache），让 React 自动渲染 UI。**

特别是使用 **Ant Design 的 Form** 组件时，利用它的 `onFinish` 和 `Rules` 校验功能，不要自己写大量的 `if (value === '')` 校验逻辑。这能让你的代码看起来像大厂出品。

---
