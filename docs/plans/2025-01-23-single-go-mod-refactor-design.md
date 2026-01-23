# 单一 go.mod 重构设计

## 概述

将项目从多模块结构（使用 go.work）重构为单一 go.mod 的标准 Go 布局。

## 目标结构

```
update-server/
├── go.mod (统一)
├── go.sum
├── config.yaml
├── build-all.bat
│
├── cmd/
│   ├── update-server/
│   │   └── main.go              (服务器入口)
│   ├── update-client/
│   │   └── main.go              (更新客户端 CLI)
│   └── update-publisher/
│       └── main.go              (发布客户端 CLI)
│
├── internal/
│   ├── client/                  (原 clients/go/client)
│   ├── config/
│   ├── database/
│   ├── handler/
│   ├── middleware/
│   ├── models/
│   └── service/
│
├── web/
├── scripts/
├── docs/
└── ...
```

## 关键变更

### 1. 删除
- `go.work` 和 `go.work.sum`
- 3 个子模块的 `go.mod` 和 `go.sum` 文件
- 空的 `clients/` 目录

### 2. 移动
- `clients/go/client/*` → `internal/client/`

### 3. 更新 import 路径

| 原路径 | 新路径 |
|--------|--------|
| `github.com/LiteHomeLab/update-client` | `docufiller-update-server/internal/client` |
| `github.com/LiteHomeLab/update-server/clients/go/client` | `docufiller-update-server/internal/client` |
| `github.com/LiteHomeLab/update-server/cmd/update-client` | `docufiller-update-server/cmd/update-client` |
| `update-publisher` | `docufiller-update-server/cmd/update-publisher` |

### 4. 合并依赖
主 go.mod 新增：
- `github.com/spf13/cobra v1.10.2`

## 执行步骤

1. 删除工作区文件
2. 移动客户端库到 internal/client
3. 删除子模块 go.mod 文件
4. 更新主 go.mod
5. 更新 import 路径
6. 执行 go mod tidy
7. 运行 build-all.bat 验证

## 验证标准

- build-all.bat 成功构建所有组件
- 所有 import 路径正确解析
- 无 go.work 相关错误
