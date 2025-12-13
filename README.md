
### 项目目录
finscale/                   
├── README.md               # 项目总说明书 (必须写得很漂亮)
├── docker-compose.yml      # 一键编排 (Postgres + Redis + Backend + FrontendDev)
├── Makefile                # 自动化命令 (make run, make build)
├── docs/                   # 文档存放
│   ├── architecture/       # 架构图 (draw.io / plantuml)
│   └── prd/                # 需求文档 (No.1 Ledger MVP.md)
│
├── backend/                # [Go] 后端代码 (模块化单体)
│   ├── go.mod              # Go 依赖定义
│   ├── cmd/                # 应用程序入口
│   │   └── api/            # 启动命令
│   │       └── main.go     # 程序主入口 (Wire wiring, Config loading)
│   ├── configs/            # 配置文件模板 (config.yaml)
│   ├── internal/           # 私有代码 (Go 规范：外部不可引用)
│   │   ├── platform/       # 基础设施层 (Database, Logger, Middleware)
│   │   │   ├── database/
│   │   │   └── server/     # HTTP Server 封装
│   │   │
│   │   └── ledger/         # === [核心业务] 总账模块 ===
│   │       ├── domain/     # 领域实体 (Account, Transaction) - 纯Go struct
│   │       ├── service/    # 业务逻辑 (Posting, Validation)
│   │       ├── repo/       # 数据库访问 (PostgreSQL Implementation)
│   │       └── api/        # 接口层 (HTTP Handlers / DTOs)
│   │
│   └── pkg/                # 公共库 (如果未来有工具类要跨项目共享)
│       └── decimal_utils/  # 比如高精度计算的封装
│
└── frontend/               # [React] 前端代码 (Vite + AntD)
    ├── package.json
    ├── vite.config.ts
    ├── src/
    │   ├── api/            # Axios 封装与后端接口定义
    │   ├── components/     # 公共组件
    │   ├── pages/          # 页面级组件
    │   │   ├── dashboard/  # 仪表盘
    │   │   └── ledger/     # 记账台
    │   ├── types/          # TypeScript 类型定义 (对应后端的 DTO)
    │   └── App.tsx
    └── public/