# RustDesk 社区 API 服务调研

> 调研时间：2026-06-27
> 目的：评估自建 RustDesk API 服务的可行性，为开源项目决策提供依据

---

## 1. 背景

RustDesk 官方只开源了服务端核心（`hbbs` + `hbbr`），Web 控制台 / API / OIDC 等功能属于闭源的 **RustDesk Server Pro**（付费产品）。

社区通过逆向 RustDesk 开源客户端的源码，实现了兼容的 API 服务端。本文将调研所有已知社区实现。

---

## 2. 各项目详情

### 2.1 lejianwen/rustdesk-api（热度最高）

| 属性 | 值 |
|---|---|
| 仓库 | https://github.com/lejianwen/rustdesk-api |
| Stars / Forks | 2,942 / 668 |
| 语言 | Go（98%）、HTML、Shell |
| 协议 | MIT |
| 创建时间 | 2024-09-13 |
| 最后推送 | **2025-09-29**（已停更 9 个月） |
| Releases | 76 个，最新 v2.7（2025-09-28） |
| Open Issues | 102 |

**技术栈**：Go 1.22 + Gin v1.9.0 + GORM v1.25.7 + Swag（Swagger）

**核心功能**：
- Web 管理后台（Vue3 + Element Plus，独立仓库 `rustdesk-api-web`）
- Web Client（v2 版本因 DMCA 于 v2.7 被移除）
- OIDC / LDAP 登录
- JWT 鉴权
- 地址簿、设备管理
- 连接日志、文件传输日志
- i18n 多语言
- 配套 server fork：`lejianwen/rustdesk-server`（556⭐，同样已停更）

**Docker 镜像**：`lejianwen/rustdesk-server-s6`（100K+ pulls，内置 API + S6-overlay）

**DMCA 事件**：RustDesk 官方于 2025-09-26 提交 DMCA，要求删除 `resources/web2` 目录（打包了官方闭源 Web Client 前端文件）。v2.7 移除了 webclient2。

---

### 2.2 kingmo888/rustdesk-api-server

| 属性 | 值 |
|---|---|
| 仓库 | https://github.com/kingmo888/rustdesk-api-server |
| Stars / Forks | 1,567 / 408 |
| 语言 | Python（100%，Django） |
| 创建时间 | 2023-12-05 |
| 最后推送 | **2024-09-25**（已停更 21 个月） |
| Releases | 4 个，最新 v1.5.3（2024-09-09） |
| Open Issues | 35 |

**特点**：基于 Django，支持 Web 注册、管理、展示。最早实现之一。

---

### 2.3 lantongxue/rustdesk-api-server-pro

| 属性 | 值 |
|---|---|
| 仓库 | https://github.com/lantongxue/rustdesk-api-server-pro |
| Stars / Forks | 317 / 77 |
| 语言 | Go + TypeScript + Vue |
| 协议 | AGPL-3.0 |
| 创建时间 | 2024-06-17 |
| 提交数 | 174 |
| Releases | **无** |
| Open Issues | 10 |

**技术栈**：
- 后端：Go + Iris v12 + XORM + gocron + JWT
- 前端：Soybean Admin（Vue3 + TypeScript）+ pnpm
- 数据库：SQLite / MySQL
- 测试：Playwright E2E + Go test
- 部署：ghcr.io Docker 镜像 + nginx 反代

**亮点**：
- 有 E2E 测试（Playwright），CI 含 typecheck + lint
- 详细架构文档（DeepWiki）
- 适配客户端 1.4.6，有明确的兼容性声明表
- 支持 2FA + 邮箱验证码
- CLI 工具完善（`user add`、`sync`、`start` 等）

**风险**：
- 单人项目（lantongxue 占 83/174 提交）
- README 开头注明 *"This project will be rewrite"*（见 issue #30）
- 无 release，无法版本锁定
- 无 OIDC/LDAP、无 Web Client
- 分组功能仍在讨论中（issue #37，2026-05）

---

### 2.4 xiaoyi510/rustdesk-api-server

| 属性 | 值 |
|---|---|
| 仓库 | https://github.com/xiaoyi510/rustdesk-api-server |
| Stars / Forks | 205 / 65 |
| 语言 | Go |
| 协议 | Apache 2.0 |
| 创建时间 | 2022-11-29 |
| 最后推送 | **~2023-06**（已停更 3 年） |
| Releases | 极少 |

**特点**：最早的 Go 实现，支持 SQLite 和 MySQL。已完全停滞。

---

### 2.5 v5star/rustdesk-api

| 属性 | 值 |
|---|---|
| 仓库 | https://github.com/v5star/rustdesk-api |
| Stars / Forks | 182 / 73 |
| 语言 | PHP |
| 协议 | 无声明 |
| 创建时间 | 2023-08-26 |
| 状态 | 停更 |

**特点**：轻量 PHP 实现，支持 SQLite 和 MySQL，Docker 镜像仅 32MB。功能较基础。

---

## 3. 综合对比

| 维度 | lejianwen | kingmo888 | lantongxue | xiaoyi510 | v5star |
|---|---|---|---|---|---|
| Stars | **2,942** | 1,567 | 317 | 205 | 182 |
| 活跃度 | 停更(2025-09) | 停更(2024-09) | 低活跃 | 死 | 死 |
| 语言 | Go | Python | Go+TS | Go | PHP |
| OIDC/LDAP | ✅ | ❌ | ❌ | ❌ | ❌ |
| Web Client | ✅(已删) | ❌ | ❌ | ❌ | ❌ |
| 2FA | ❌ | ❌ | ✅ | ❌ | ❌ |
| E2E 测试 | ❌ | ❌ | ✅ | ❌ | ❌ |
| Release 管理 | ✅(76个) | ✅(4个) | ❌ | ❌ | ❌ |
| Docker | ✅ | ✅ | ✅(ghcr) | ✅ | ✅ |
| 协议 | MIT | — | AGPL-3.0 | Apache 2.0 | — |

---

## 4. 配套 Server Fork

| 项目 | Stars | 状态 | 说明 |
|---|---|---|---|
| `lejianwen/rustdesk-server` | 556 | 停更(2025-09) | fork 官方，添加 API 超时修复、强制登录、WebSocket |
| `lejianwen/rustdesk/rustdesk` | — | 停更 | 客户端 fork，修复 API 登录慢的问题 |
| 官方 `rustdesk/rustdesk-server` | **10,000** | **活跃** | 最新 release v1.1.15（2026-01），最后推送 2026-06-04 |

> ⚠️ lejianwen 的 server fork 停留在 v0.1.2，与官方 v1.1.15 差距约 1.5 年。

---

## 5. API 逆向方法

RustDesk **从未公开过 API 文档**。社区实现通过以下方式获取接口信息：

1. **阅读客户端源码** — RustDesk 客户端开源（`rustdesk/rustdesk`），代码中硬编码了 API 调用逻辑。直接读 `src/client/` 即可获取 endpoint、请求格式、认证流程。
2. **抓包分析** — 客户端支持配置自定义 API 服务器地址，配合抓包工具可观察完整 HTTP 交互。
3. **官方 Demo 仓库** — `rustdesk/rustdesk-server-demo` 提供 protobuf 协议参考。
4. **已有社区实现** — 上述项目均为 AGPL/MIT 协议，可直接参考路由和 handler。

**API 架构**（基于 lantongxue 的文档）：
- `/api/*` — 客户端 API（login、system/sysinfo、audit 等）
- `/admin/*` — 管理后台 API（user CRUD、session 监控等）
- 两套各自独立的 JWT 认证中间件

---

## 6. 维护成本分析

所有 5 个项目都死了/半死，共同原因：

| 死因 | 影响项目 | 说明 |
|---|---|---|
| **单人维护 burn out** | 全部 | 所有项目主要贡献者 ≤ 3 人 |
| **追客户端版本** | 全部 | 官方客户端更新，API 兼容性需持续适配 |
| **DMCA 风险** | lejianwen | 打包官方闭源前端文件被 takedown |
| **无用户网络效应** | xiaoyi510, v5star | 没人用 → 没人贡献 → 死循环 |
| **官方 Pro 竞品** | 全部 | RustDesk 的商业模型就是卖 API/Web 功能 |

---

## 7. 自建开源项目建议

### 核心策略：Fork lejianwen/rustdesk-api（MIT）

**理由**：
- 2,942 stars 的用户基础
- 76 个 release 积累的 issue 反馈
- MIT 协议，无限制
- 功能最完整（OIDC/LDAP/WebSocket/地址簿）

**而不是从零开始**：从零写 = 花 3 个月追平现有功能 + 没人用。Fork = 花 2 周修复更新 + 发布。

### 启动清单

| 优先级 | 任务 | 时间 |
|---|---|---|
| P0 | Fork + 适配最新 RustDesk 客户端 | 1 周 |
| P0 | 发布第一个 release + Docker 镜像 | 3 天 |
| P1 | CONTRIBUTING.md + 本地开发指南 | 2 天 |
| P1 | CI/CD（lint + test + build + docker push） | 3 天 |
| P2 | 清理 open issues | 持续 |

### 差异化机会

- **稳定的 release 节奏**（竞品都没做到）
- **版本兼容性矩阵**（声明每个 release 支持的客户端版本）
- **PostgreSQL 支持**（竞品只有 SQLite/MySQL）
- **Docker Compose 一键部署**
- **英文文档**（lejianwen 以中文为主）

### DMCA 红线

❌ **绝对不要**在仓库中包含 RustDesk 官方闭源前端文件（`.js`、`.wasm`、`.html`）。

API 服务端代码本身是 AGPL 兼容的，不会被 takedown。lejianwen 被 DMCA 是因为 `resources/web2` 目录打包了官方闭源 Web Client。

---

## 8. 相关链接

### 官方资源
- RustDesk Server：https://github.com/rustdesk/rustdesk-server（10K⭐，AGPL-3.0）
- RustDesk Server Pro：https://github.com/rustdesk/rustdesk-server-pro（闭源）
- RustDesk Server Demo：https://github.com/rustdesk/rustdesk-server-demo
- 官方 S6 Docker 镜像：`rustdesk/rustdesk-server-s6` / `ghcr.io/rustdesk/rustdesk-server-s6`

### 社区项目
- lejianwen API：https://github.com/lejianwen/rustdesk-api（2.9K⭐，MIT，Go）
- lejianwen Server Fork：https://github.com/lejianwen/rustdesk-server（556⭐，AGPL-3.0）
- kingmo888：https://github.com/kingmo888/rustdesk-api-server（1.5K⭐，Python/Django）
- lantongxue：https://github.com/lantongxue/rustdesk-api-server-pro（317⭐，Go+Vue）
- xiaoyi510：https://github.com/xiaoyi510/rustdesk-api-server（205⭐，Go）
- v5star：https://github.com/v5star/rustdesk-api（182⭐，PHP）

### Docker 镜像
- `lejianwen/rustdesk-server-s6`（100K+ pulls，已停更）
- `ghcr.io/lantongxue/rustdesk-api-server-pro:latest`
- 官方：`rustdesk/rustdesk-server-s6:latest`
