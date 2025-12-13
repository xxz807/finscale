
# FinScale 产品需求文档 (PRD) (Core Ledger)

| 文档属性 | 内容 |
| :--- | :--- |
| **项目名称** | FinScale (Financial Scalable Core) |
| **子系统** | **Core Ledger (核心总账引擎)** |
| **代号** | Titan (泰坦) |
| **版本** | v1.0 (MVP) |
| **状态** | Draft |
| **负责人** | Jack Xia |
| **最后更新** | 2025-12-13 |

---

## 1. 概述 (Executive Summary)

### 1.1 背景
传统银行核心系统（Mainframe/Legacy）通常存在扩展性差、维护成本高、数据孤岛等问题。`FinScale` 旨在利用现代云原生技术栈（Go + PostgreSQL），构建一套高并发、强一致性、可审计的下一代金融基础设施。

**Core Ledger (总账引擎)** 是 `FinScale` 的心脏。它不处理具体的业务场景（如贷款利率计算），只负责最底层的**资金价值记录**。

### 1.2 核心理念
1.  **会计恒等 (Accounting Equation)**：任何时刻，全系统必须满足 `资产 = 负债 + 所有者权益`。
2.  **复式记账 (Double-Entry)**：有借必有贷，借贷必相等。
3.  **不可篡改 (Immutability)**：账目一旦入库（Posted），严禁物理删除或修改。错误的冲正必须通过“红冲蓝补”（Reversal Entries）实现。
4.  **精确性 (Precision)**：杜绝浮点数误差，全链路采用高精度定点数。

---

## 2. 领域模型 (Domain Model)

在开发前，需统一全团队（即视频观众）的术语：

*   **Chart of Accounts (COA, 科目表)**: 银行的账本目录。
    *   *示例*: `1001-库存现金`, `2001-客户存款`。
*   **Journal Entry (会计分录)**: 记录资金变动的最小单元，包含科目、金额、方向（借/贷）。
*   **Transaction (交易)**: 一个原子的业务动作，包含一组分录（Entries）。
    *   *约束*: 一个 Transaction 内的所有 Entries 的 `(借方总额 - 贷方总额)` 必须为 0。
*   **Posting (过账)**: 将交易永久写入数据库并更新科目余额的过程。

---

## 3. 功能需求 (Functional Requirements)

### F-01: 会计科目管理 (COA Management)
**目标**: 定义资金存放的“桶”。

*   **F-01-01**: 支持建立树状科目体系。
    *   *字段*: `ID`, `AccountCode` (Unique), `Name`, `Type` (Asset/Liability/Equity/Income/Expense), `Currency`, `Status` (Active/Frozen).
*   **F-01-02**: 科目余额方向控制。
    *   资产/费用类科目：借方增加，贷方减少。
    *   负债/权益/收入类科目：贷方增加，借方减少。

### F-02: 核心记账引擎 (The Posting Engine) —— **核心功能**
**目标**: 处理资金流动，保证 ACID。

*   **F-02-01 (原子性验证)**: 接收一笔交易请求，必须包含至少两条分录。
*   **F-02-02 (试算平衡检查)**:
    *   系统必须计算 `Sum(Debit)` 和 `Sum(Credit)`。
    *   如果 `Difference != 0`，拒绝交易并返回 `ERR_BALANCE_MISMATCH`。
*   **F-02-03 (余额更新)**:
    *   根据科目类型更新实时余额。
    *   **并发控制**: 必须处理“丢失更新”问题（即两个线程同时修改同一账户余额）。
*   **F-02-04 (幂等性控制)**:
    *   每个请求必须携带 `Idempotency-Key`。重复提交相同的 Key，系统应返回“成功”但不重复记账。

### F-03: 财务报表与查询 (Reporting)
*   **F-03-01 (账户余额查询)**: 查询指定账户在当前时刻的余额。
*   **F-03-02 (交易流水查询)**: 根据账户 ID 查询历史变动明细。
*   **F-03-03 (全行试算平衡表)**: 一个 Dashboard 接口，返回全行总资产、总负债，验证系统健康度。

---

## 4. 数据架构设计 (Data Schema)

本系统使用 PostgreSQL，利用其强大的事务能力和 JSONB 扩展性。

### 4.1 表结构概览

#### Table: `accounts` (科目表)
| 字段名 | 类型 | 约束 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | BIGSERIAL | PK | 内部主键 |
| `account_code` | VARCHAR(32) | UNIQUE | 业务主键 (e.g., "1001") |
| `balance` | DECIMAL(20, 4) | NOT NULL | **当前余额 (核心)** |
| `currency` | CHAR(3) | NOT NULL | ISO 4217 (CNY, USD) |
| `account_type`| SMALLINT | NOT NULL | 1:Asset, 2:Liability... |
| `version` | BIGINT | DEFAULT 0 | **乐观锁版本号** |

#### Table: `transactions` (交易主表)
| 字段名 | 类型 | 约束 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | BIGSERIAL | PK | |
| `reference_id`| VARCHAR(64) | UNIQUE | 外部业务ID (幂等键) |
| `posted_at` | TIMESTAMPTZ | NOT NULL | 记账时间 |
| `metadata` | JSONB | | 扩展字段 (备注、操作人等) |

#### Table: `postings` (分录明细表)
| 字段名 | 类型 | 约束 | 说明 |
| :--- | :--- | :--- | :--- |
| `id` | BIGSERIAL | PK | |
| `tx_id` | BIGINT | FK | 关联 `transactions.id` |
| `account_id` | BIGINT | FK | 关联 `accounts.id` |
| `amount` | DECIMAL(20, 4) | > 0 | **绝对值** (必须为正) |
| `direction` | CHAR(1) | 'D'/'C' | 借贷方向 (Debit/Credit) |

---

## 5. API 接口定义 (Interface Contract)

### 5.1 记账接口 (Post Transaction)
*   **Method**: `POST /api/v1/ledger/transactions`
*   **Request Body**:
    ```json
    {
      "reference_id": "tx_20251213_001",
      "description": "Customer Deposit",
      "postings": [
        {
          "account_code": "1001", // 库存现金 (资产, 借增)
          "amount": "100.00",
          "direction": "DEBIT"
        },
        {
          "account_code": "2001", // 客户存款 (负债, 贷增)
          "amount": "100.00",
          "direction": "CREDIT"
        }
      ]
    }
    ```
*   **Success Response (200 OK)**:
    ```json
    { "tx_id": "105", "status": "POSTED" }
    ```
*   **Error Response (400 Bad Request)**:
    ```json
    { "code": "E_LEDGER_IMBALANCE", "message": "Debits (100) do not equal Credits (90)" }
    ```

---

## 6. 非功能性需求 (NFR)

### 6.1 性能指标 (Performance)
*   **TPS (Transactions Per Second)**: 单机（2C4G）需支持至少 1,000 TPS 的并发记账。
*   **Latency (延迟)**: P99 < 100ms。

### 6.2 精度与舍入 (Precision)
*   所有金额计算必须使用 `decimal` 库（Go语言使用 `github.com/shopspring/decimal`）。
*   数据库存储使用 `DECIMAL(20,4)`，保留 4 位小数，前端展示时截断为 2 位。

### 6.3 异常处理 (Error Handling)
*   系统必须定义明确的错误码体系 (Error Codes)，如：
    *   `E_ACCOUNT_NOT_FOUND`
    *   `E_INSUFFICIENT_FUNDS` (针对不允许透支的账户)
    *   `E_IDEMPOTENCY_VIOLATION`

---

## 7. 风险与合规 (Risk & Compliance)

*   **审计痕迹 (Audit Trail)**: 所有 `postings` 表的数据均为 `Append-only`，严禁 `UPDATE` 操作修改金额。
*   **负余额检查**: 资产类账户（如现金）默认不允许余额为负，除非开启“透支 (Overdraft)”标志。

---
