# FinScale 产品需求文档 (PRD)

| 文档属性 | 内容 |
| :--- | :--- |
| **项目名称** | FinScale (Financial Scalable Core) |
| **子系统** | **Core Ledger (核心总账引擎)** |
| **代号** | Titan (泰坦) |
| **架构模式** | Modular Monolith (Schema Isolation) |
| **版本** | **v1.3.0 (Final Architecture)** |
| **状态** | **Approved** |
| **负责人** | Jack Xia |
| **最后更新** | 2025-12-13 |

---

## 1. 概述 (Executive Summary)

### 1.1 背景
传统银行核心系统（Mainframe/Legacy）存在扩展性差、数据孤岛等问题。`FinScale` 利用现代云原生技术栈（Go + PostgreSQL），构建一套高并发、强一致性、可审计的下一代金融基础设施。

**Core Ledger (总账引擎)** 是心脏模块。它不处理具体的业务形态（如红包、理财），只负责最底层的**资金价值记录与平衡**。

### 1.2 核心设计理念
1.  **会计恒等 (Accounting Equation)**：`资产 = 负债 + 所有者权益` 是系统的绝对公理。
2.  **Schema 隔离 (Modular Isolation)**：总账模块拥有独立的数据库 Schema (`ledger`)，禁止其他模块直接 JOIN，为未来微服务拆分预留能力。
3.  **双重防御 (Double Defense)**：
    *   **物理定律**（如借贷方向）由数据库 `CHECK` 约束严防死守。
    *   **业务规则**（如交易类型）保持松散，适应业务快速迭代。
4.  **不可篡改 (Immutability)**：分录一旦入库，严禁 `UPDATE/DELETE`。修正必须通过“红冲蓝补”。

---

## 2. 领域模型 (Domain Model)

*   **Chart of Accounts (COA, 科目)**: 资金的容器。
    *   *物理属性*: 属于 5 大类之一 (Asset, Liability, Equity, Income, Expense)。
*   **Journal Entry (分录)**: 资金变动的原子单元。
    *   *规则*: 金额必须 > 0，方向只能是 Debit 或 Credit。
*   **Transaction (交易)**: 业务动作的集合。
    *   *ACID*: 一个 Transaction 内的所有 Entries 必须借贷平衡 `(Sum(Dr) == Sum(Cr))`，否则原子性回滚。

---

## 3. 功能需求 (Functional Requirements)

### F-01: 会计科目管理 (COA)
**目标**: 维护银行的账本目录。
*   **F-01-01**: 支持 5 种标准会计类型（数据库底层约束）。
*   **F-01-02**: 币种支持 ISO 4217 标准（如 `CNY`, `USD`），保留扩展性。
*   **F-01-03**: 状态机管理（正常 -> 冻结 -> 销户）。

### F-02: 核心记账引擎 (Posting Engine)
**目标**: 处理高并发资金流动。
*   **F-02-01 (借贷平衡检查)**: 交易提交时，系统计算 `Sum(Debit)` 与 `Sum(Credit)`，差值必须为 0。
*   **F-02-02 (并发控制)**: 采用 **乐观锁 (Optimistic Locking)** 机制 (`CAS on version`) 更新余额，防止热点账户金额覆盖。
*   **F-02-03 (幂等性)**: 基于 `reference_id` 进行唯一性拦截，防止网络超时导致的重复记账。

### F-03: 审计与报表
*   **F-03-01**: 实时试算平衡 (Trial Balance) 接口，监控全行健康度。

---

## 4. 数据架构设计 (Data Schema)

> 详细设计参考 `docs/architecture/db-design-ledger.md`

### 4.1 Schema 策略
所有表位于独立 Schema：**`ledger`**。

### 4.2 核心表定义

#### Table: `ledger.accounts` (状态表)
| 字段 | 类型 | 核心约束/逻辑 | 说明 |
| :--- | :--- | :--- | :--- |
| `account_code` | VARCHAR | UNIQUE | 业务主键 (e.g., "1001") |
| `type` | SMALLINT | **CHECK (1-5)** | **物理定律**：1:资产 ... 5:费用 |
| `currency` | CHAR(3) | 无约束 | **灵活性**：支持未来新币种 |
| `balance` | DECIMAL | DEFAULT 0 | **高精度**：保留 4 位小数 |
| `version` | BIGINT | DEFAULT 1 | **乐观锁**：每次变动 +1 |
| `status` | SMALLINT | **CHECK (1,0,-1)** | 1:正常, 0:冻结, -1:销户 |

#### Table: `ledger.transactions` (上下文表)
| 字段 | 类型 | 核心约束/逻辑 | 说明 |
| :--- | :--- | :--- | :--- |
| `reference_id` | VARCHAR | UNIQUE | **幂等键** |
| `tx_type` | VARCHAR | 无约束 | **灵活性**：如 "DEPOSIT", "FEE" |
| `posted_at` | TIMESTAMP | INDEX | 记账时间 |

#### Table: `ledger.postings` (流水表)
| 字段 | 类型 | 核心约束/逻辑 | 说明 |
| :--- | :--- | :--- | :--- |
| `direction` | CHAR(1) | **CHECK ('D','C')** | **物理定律**：只有借/贷 |
| `amount` | DECIMAL | **CHECK (> 0)** | **防御**：绝对值必须为正 |

---

## 5. API 接口定义 (Interface Contract)

### 5.1 记账接口 (Post Transaction)
*   **Endpoint**: `POST /api/v1/ledger/transactions`
*   **Design Note**: 金额字段必须使用 **String** 类型传输，避免 JSON 浮点数精度丢失。

*   **Request Body**:
    ```json
    {
      "reference_id": "tx_unique_20251213_001",
      "tx_type": "DEPOSIT",
      "description": "Customer Cash Deposit",
      "postings": [
        {
          "account_code": "1001", 
          "amount": "100.00",  // String! Not Number
          "direction": "D"     // 借：金库现金
        },
        {
          "account_code": "2001", 
          "amount": "100.00",
          "direction": "C"     // 贷：客户存款
        }
      ]
    }
    ```

*   **Error Response (409 Conflict)**:
    *   场景：余额被并发修改。
    *   Code: `E_CONCURRENT_MODIFICATION`
    *   Action: 客户端应重新获取最新余额或重新发起请求。

---

## 6. 非功能性需求 (NFR)

### 6.1 性能 (Performance)
*   **TPS**: 单实例目标 > 1,000 TPS (混合读写)。
*   **Latency**: P99 < 50ms (内部网络)。

### 6.2 精度 (Precision)
*   **Database**: `DECIMAL(20,4)`
*   **Go Lang**: 使用 `github.com/shopspring/decimal`，严禁使用 `float64` 进行金额运算。

### 6.3 风险控制 (Risk)
*   **数据底线**: 数据库层面的 `CHECK` 约束是最后一道防线，即使应用层代码由 Bug，也不能写入违反会计原理的数据（如负金额）。

---

### 💡 主要更新点（对比 v1.0）：
1.  **Schema**: 明确指定了 `ledger.` 前缀。
2.  **Constraints**: 区分了“物理定律”（强约束）和“业务规则”（弱约束）。
3.  **API**: 强调了 JSON 中金额必须传 String。
4.  **Concurrency**: 明确了乐观锁机制。