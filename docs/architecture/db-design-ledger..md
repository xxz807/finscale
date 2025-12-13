# 📘 FinScale 数据库设计文档 (Core Ledger)

| 文档信息 | 内容 |
| :--- | :--- |
| **模块** | **Ledger (Core Banking)** |
| **数据库** | PostgreSQL 15+ |
| **设计原则** | 3NF (第三范式), 强一致性 (ACID), 不可变性 (Immutability) |
| **负责人** | [你的名字] |
| **最后更新** | 2025-12-13 |

---

## 1. 设计概述 (Design Overview)

FinScale 的核心账务层采用 **PostgreSQL** 作为存储引擎，利用其以下特性：
1.  **ACID 事务**：保证借贷记账的原子性。
2.  **Row-Level Locking (行级锁)**：在高并发下保护账户余额。
3.  **JSONB**：存储交易的扩展元数据（Metadata）。
4.  **Constraints (约束)**：在数据库层面防止坏账（如金额不能为负）。

### 实体关系图 (ER Diagram 简述)
*   **`accounts` (1) --- (< N) `postings` (N) >--- (1) `transactions`**
*   一个交易 (`transaction`) 包含多个分录 (`postings`)。
*   每个分录 (`posting`) 归属于一个账户 (`account`)。

---

## 2. 表结构详解 (Table Definitions)

### 2.1 账户表 `accounts`
**用途**：存储会计科目及其当前余额。这是系统的“状态”表。

```sql
CREATE TABLE accounts (
    id              BIGSERIAL PRIMARY KEY,
    
    -- 业务主键：科目代码 (e.g., "1001", "2001-01")
    account_code    VARCHAR(32) NOT NULL UNIQUE, 
    
    -- 科目名称
    name            VARCHAR(100) NOT NULL,
    
    -- 账户类型：1-Asset, 2-Liability, 3-Equity, 4-Income, 5-Expense
    type            SMALLINT NOT NULL,
    
    -- 币种 (ISO 4217)
    currency        CHAR(3) NOT NULL DEFAULT 'CNY',
    
    -- 当前余额 (核心字段)
    -- 精度：20位总长度，4位小数。支持百兆亿级别的金额。
    balance         DECIMAL(20, 4) NOT NULL DEFAULT 0.0000,
    
    -- 乐观锁版本号 (用于并发控制)
    version         BIGINT NOT NULL DEFAULT 1,
    
    -- 状态：1-Active, 0-Frozen
    status          SMALLINT NOT NULL DEFAULT 1,
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 注释
COMMENT ON TABLE accounts IS '会计科目/账户表';
COMMENT ON COLUMN accounts.version IS '乐观锁版本控制，每次更新余额 +1';
COMMENT ON COLUMN accounts.balance IS '实时余额，正负取决于借贷方向';
```

### 2.2 交易主表 `transactions`
**用途**：交易的“Header”，记录一笔业务发生的元信息。

```sql
CREATE TABLE transactions (
    id              BIGSERIAL PRIMARY KEY,
    
    -- 幂等性键 (Idempotency Key)
    -- 外部系统必须传入此ID，防止网络超时导致的重复扣款
    reference_id    VARCHAR(64) NOT NULL UNIQUE,
    
    -- 业务类型 (e.g., "DEPOSIT", "TRANSFER", "FEE")
    tx_type         VARCHAR(32) NOT NULL,
    
    -- 交易摘要
    description     TEXT,
    
    -- 记账时间 (业务发生时间)
    posted_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- 扩展元数据 (JSONB)
    -- 存放如：操作员ID、地理位置、关联订单号等
    metadata        JSONB,
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_transactions_posted_at ON transactions(posted_at);
COMMENT ON TABLE transactions IS '交易主表，记录业务发生的上下文';
```

### 2.3 资金分录表 `postings`
**用途**：交易的“Lines”，记录具体的资金流动方向。这是系统的“流水”表。
**核心规则**：同一 `transaction_id` 下的所有 `amount`（结合 `direction`）总和必须平衡。

```sql
CREATE TABLE postings (
    id              BIGSERIAL PRIMARY KEY,
    
    -- 关联主表
    transaction_id  BIGINT NOT NULL REFERENCES transactions(id),
    
    -- 关联账户
    account_id      BIGINT NOT NULL REFERENCES accounts(id),
    
    -- 借贷方向：'D' (Debit) 或 'C' (Credit)
    direction       CHAR(1) NOT NULL CHECK (direction IN ('D', 'C')),
    
    -- 变动金额
    -- 必须为正数！我们在借贷记账中不说"借-100"，而说"贷100"
    amount          DECIMAL(20, 4) NOT NULL CHECK (amount > 0),
    
    -- 业务发生时的汇率 (预留给多币种，MVP阶段默认为 1.0)
    exchange_rate   DECIMAL(10, 6) DEFAULT 1.000000
);

-- 索引：高频查询某账户的流水
CREATE INDEX idx_postings_account_id ON postings(account_id);
CREATE INDEX idx_postings_tx_id ON postings(transaction_id);

COMMENT ON TABLE postings IS '会计分录表，记录资金流动的原子操作';
COMMENT ON COLUMN postings.amount IS '绝对值金额，必须 > 0';
```

---

## 3. 关键设计决策 (Design Rationale)

### 3.1 为什么不使用浮点数 (Float/Double)?
在金融系统中，`Float` 存在精度丢失问题（例如 `0.1 + 0.2 != 0.3`）。这会导致账目在经过数百万次计算后出现“几分钱”的误差，导致审计失败。
**决策**：全链路使用 `DECIMAL(20,4)` 固定精度类型。

### 3.2 为什么需要 `version` 字段?
这是为了处理 **并发热点账户 (Hot Account)** 问题。
如果不加锁，两个并发请求同时读取余额 100，分别扣 10，最后可能余额变成 90 而不是 80。
**决策**：使用乐观锁 (Optimistic Locking)。
SQL 逻辑：
```sql
UPDATE accounts 
SET balance = balance + ?, version = version + 1 
WHERE id = ? AND version = ?;
```
如果 `Affected Rows == 0`，说明在读取和写入之间有别人修改了余额，当前事务回滚并重试。

### 3.3 为什么分录金额必须大于 0?
遵循复式记账法原则。
*   错误写法：借方 -100。
*   正确写法：贷方 100。
这保证了数据的语义清晰，方便后续审计。

### 3.4 为什么没有 Delete / Update?
**铁律**：总账系统是 **Append-Only (仅追加)** 的。
如果一笔交易记错了，不能直接去数据库改数，必须发起一笔新的 **“冲正交易” (Reversal Transaction)**，借贷方向相反，把账做平。这留下了完整的审计痕迹。

---

## 4. 预置数据 (Seed Data - MVP)

在系统初始化时，我们将插入以下基础科目，构建一个最小化的银行账本：

```sql
-- 1. 资产类 (Assets)
INSERT INTO accounts (account_code, name, type, currency, balance) VALUES 
('1001', 'Vault Cash (金库现金)', 1, 'CNY', 10000000.0000), -- 初始一千万本金
('1002', 'User Wallet Pool (用户资金池)', 1, 'CNY', 0.0000);

-- 2. 负债类 (Liabilities)
INSERT INTO accounts (account_code, name, type, currency, balance) VALUES 
('2001', 'Customer Deposits (客户存款)', 2, 'CNY', 0.0000),
('2002', 'Accounts Payable (应付账款)', 2, 'CNY', 0.0000);

-- 3. 权益类 (Equity)
INSERT INTO accounts (account_code, name, type, currency, balance) VALUES 
('3001', 'Owners Capital (实收资本)', 3, 'CNY', 10000000.0000); -- 对应金库现金
```

---
