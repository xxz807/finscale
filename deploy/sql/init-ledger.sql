/*
 * =============================================================================
 * FinScale - Next Gen Financial Core System
 * Module: Ledger (Core Banking)
 * Database Initialization Script
 * 
 * Version: v1.3.0 (Final Architecture)
 * Database: PostgreSQL 15+
 * Architecture: Modular Monolith (Schema Isolation)
 * Author: FinScale Architect
 * =============================================================================
 */

-- 1. 开启事务：保证初始化过程的原子性 (All or Nothing)
BEGIN;

-- 2. 创建独立 Schema
-- 架构决策：使用独立 Schema 'ledger' 实现模块化隔离。
-- 未来若需拆分为微服务，只需迁移此 Schema 下的数据即可。
CREATE SCHEMA IF NOT EXISTS ledger;

-- 3. 设置搜索路径 (可选，方便后续脚本执行，但在生产代码中建议显式指定 schema)
SET search_path TO ledger, public;

-- 4. 清理旧对象 (仅用于开发/测试环境重置，生产环境请注释掉)
DROP TABLE IF EXISTS ledger.postings CASCADE;
DROP TABLE IF EXISTS ledger.transactions CASCADE;
DROP TABLE IF EXISTS ledger.accounts CASCADE;

-- =============================================================================
-- Table: ledger.accounts (会计科目/账户表)
-- Description: 核心状态表。利用数据库 Check 约束构建"物理定律"防线。
-- =============================================================================
CREATE TABLE ledger.accounts (
    id              BIGSERIAL PRIMARY KEY,
    
    -- 业务主键：科目代码 (e.g., "1001", "2001-01")
    -- 必须全局唯一，用于业务查找
    account_code    VARCHAR(32) NOT NULL UNIQUE, 
    
    -- 科目名称
    name            VARCHAR(100) NOT NULL,
    
    -- [约束强化] 账户类型：1-Asset, 2-Liability, 3-Equity, 4-Income, 5-Expense
    -- 架构决策：这是金融系统的物理定律，500年未变，必须在数据库底层锁死。
    type            SMALLINT NOT NULL CHECK (type BETWEEN 1 AND 5),
    
    -- [灵活性预留] 币种：使用 ISO 4217 代码 (e.g., 'CNY', 'USD', 'BTC')
    -- 架构决策：不加 Check 约束，允许未来动态支持新币种，无需修改 Schema。
    currency        CHAR(3) NOT NULL DEFAULT 'CNY',
    
    -- [核心字段] 当前余额
    -- 精度决策：使用 DECIMAL(20,4) 杜绝浮点数精度丢失问题。
    balance         DECIMAL(20, 4) NOT NULL DEFAULT 0.0000,
    
    -- [并发控制] 乐观锁版本号
    -- 每次余额变动必须 version + 1，防止并发覆盖 (Lost Update)。
    version         BIGINT NOT NULL DEFAULT 1,
    
    -- [约束强化] 状态：1=正常, 0=冻结, -1=销户
    -- 状态机的合法性由数据库底座保证。
    status          SMALLINT NOT NULL DEFAULT 1 CHECK (status IN (1, 0, -1)),
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 添加表级注释，方便数据库工具查看
COMMENT ON TABLE ledger.accounts IS '核心会计科目表 (State)';
COMMENT ON COLUMN ledger.accounts.type IS '1:资产, 2:负债, 3:权益, 4:收入, 5:费用';
COMMENT ON COLUMN ledger.accounts.version IS '乐观锁版本号，每次变动+1';

-- =============================================================================
-- Table: ledger.transactions (交易主表)
-- Description: 交易 Header。TxType 使用字符串以适应业务多变性。
-- =============================================================================
CREATE TABLE ledger.transactions (
    id              BIGSERIAL PRIMARY KEY,
    
    -- 幂等性键 (Idempotency Key)
    -- 外部系统必须传入此ID，防止网络超时导致的重复扣款。
    reference_id    VARCHAR(64) NOT NULL UNIQUE,
    
    -- [灵活性预留] 业务类型：使用字符串 (e.g., "DEPOSIT", "MARKETING_REWARD")
    -- 架构决策：业务场景是发散的，不加 Check 约束，方便运维阅读日志和快速扩展业务。
    tx_type         VARCHAR(32) NOT NULL,
    
    -- 交易摘要
    description     TEXT,
    
    -- 记账时间 (业务发生时间)
    posted_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- 扩展元数据 (JSONB)
    -- 用于存储非结构化上下文，如：操作员IP、地理位置、关联订单快照。
    metadata        JSONB,
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_posted_at ON ledger.transactions(posted_at);
CREATE INDEX idx_transactions_ref_id ON ledger.transactions(reference_id);
COMMENT ON TABLE ledger.transactions IS '交易主表 (Immutable Log Header)';

-- =============================================================================
-- Table: ledger.postings (资金分录表)
-- Description: 资金流水 Lines。方向和金额必须严谨。
-- =============================================================================
CREATE TABLE ledger.postings (
    id              BIGSERIAL PRIMARY KEY,
    
    -- 关联主表 (Transaction)
    transaction_id  BIGINT NOT NULL REFERENCES ledger.transactions(id),
    
    -- 关联账户 (Account)
    account_id      BIGINT NOT NULL REFERENCES ledger.accounts(id),
    
    -- [物理定律] 借贷方向：只有 'D' (Debit) 和 'C' (Credit)
    -- 严禁其他值。
    direction       CHAR(1) NOT NULL CHECK (direction IN ('D', 'C')),
    
    -- [物理定律] 金额必须为正
    -- 复式记账法中，金额永远是正数，方向决定增减。
    amount          DECIMAL(20, 4) NOT NULL CHECK (amount > 0),
    
    -- 业务发生时的汇率 (预留给多币种支持)
    exchange_rate   DECIMAL(10, 6) DEFAULT 1.000000
);

CREATE INDEX idx_postings_account_id ON ledger.postings(account_id);
CREATE INDEX idx_postings_tx_id ON ledger.postings(transaction_id);
COMMENT ON TABLE ledger.postings IS '会计分录表 (Immutable Log Lines)';


-- =============================================================================
-- Seed Data: 初始化银行账本 (Bootstrapping)
-- Description: 构建"创世区块"，确保系统启动时即处于平衡状态。
-- =============================================================================

-- 1. 定义初始科目体系 (Chart of Accounts)
-- 注意：这里显式指定了 schema prefix 'ledger.'
INSERT INTO ledger.accounts (account_code, name, type, currency, balance, status) VALUES 
-- 资产类 (Assets)
('1001', 'Vault Cash (金库现金)', 1, 'CNY', 0.0000, 1),
('1002', 'User Wallet Pool (用户资金池)', 1, 'CNY', 0.0000, 1),
-- 负债类 (Liabilities)
('2001', 'Customer Deposits (客户存款)', 2, 'CNY', 0.0000, 1),
('2002', 'System Payable (系统应付)', 2, 'CNY', 0.0000, 1),
-- 权益类 (Equity)
('3001', 'Owners Capital (实收资本)', 3, 'CNY', 0.0000, 1);


-- 2. 模拟一笔“初始注资”交易 (Capital Injection)
-- 场景：股东向银行金库注入 1,000,000 元作为启动资金
-- 分录：借 1001 (现金+)，贷 3001 (资本+)

-- A. 插入交易 Header
INSERT INTO ledger.transactions (id, reference_id, tx_type, description, posted_at) 
VALUES (1, 'INIT-GENESIS-001', 'CAPITAL_INJECTION', 'Genesis Capital Setup', NOW());

-- B. 插入交易 Lines (Postings)
INSERT INTO ledger.postings (transaction_id, account_id, direction, amount) VALUES
(1, (SELECT id FROM ledger.accounts WHERE account_code = '1001'), 'D', 1000000.00), -- 借：现金
(1, (SELECT id FROM ledger.accounts WHERE account_code = '3001'), 'C', 1000000.00); -- 贷：资本

-- C. 更新账户余额 (模拟 Post 过程)
-- 逻辑：资产类借方增加，权益类贷方增加
UPDATE ledger.accounts SET balance = balance + 1000000.00, version = version + 1 WHERE account_code = '1001';
UPDATE ledger.accounts SET balance = balance + 1000000.00, version = version + 1 WHERE account_code = '3001';

-- 5. 提交所有变更
COMMIT;

-- Print Success Message (Optional, for CLI tools)
-- DO $$ BEGIN RAISE NOTICE 'FinScale Core Ledger Database Initialized Successfully.'; END $$;