# Mock Gitea Server

一个用 Go 语言编写的 Mock Gitea API 服务器，用于测试和开发环境中模拟 Gitea 的部分核心功能。

## 功能特性

- **用户接口** - 获取当前用户信息
- **仓库搜索** - 支持分页的仓库列表搜索
- **仓库分支** - 获取仓库的分支列表
- **提交历史** - 获取仓库的提交记录
- **提交详情** - 获取单个提交的详细信息

## 技术栈

- Go 1.25+
- 标准库 net/http

## 快速开始

### 运行服务

```bash
go run main.go
```

服务默认监听端口 `3333`。

### API 接口

#### 获取当前用户信息

```bash
curl http://localhost:3333/api/v1/user
```

#### 搜索仓库列表

```bash
curl "http://localhost:3333/api/v1/repos/search?page=1&limit=10"
```

#### 获取仓库分支

```bash
curl "http://localhost:3333/api/v1/repos/zhangsan/backend-api/branches"
```

#### 获取仓库提交历史

```bash
curl "http://localhost:3333/api/v1/repos/zhangsan/backend-api/commits"
```

#### 获取单个提交详情

```bash
curl "http://localhost:3333/api/v1/repos/zhangsan/backend-api/git/commits/<sha>"
```

## 配置参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| Port | 3333 | 服务监听端口 |
| TotalUsers | 10 | 模拟用户数量 |
| TotalRepos | 30 | 模拟仓库数量 |
| CommitsPerBranch | 200 | 每个分支的提交数量 |
| MaxRecentDays | 180 | 提交时间范围（天） |

## 项目结构

```
mock-gitea/
├── main.go                    # 程序入口
├── internal/
│   ├── config/               # 配置定义
│   ├── data/                # 模拟数据生成
│   ├── models/              # 数据模型
│   ├── server/              # HTTP 处理器
│   └── utils/               # 工具函数
```

## 模拟数据

- **10 个用户**：zhangsan, lisi, wangwu, zhaoliu, chenxi, yangfan, zhoumo, sunqian, wenyu, linran
- **30 个仓库**：涵盖 backend, frontend, mobile, data, infra, tooling, ai 等领域
- **分支与提交**：每个仓库包含多个分支，每分支约 200 条提交记录

## 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件
