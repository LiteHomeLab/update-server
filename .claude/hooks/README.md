# Claude Code Pushover 通知 Hook

## 问题诊断

如果通知未收到，按以下步骤排查：

### 快速诊断

在项目目录中运行：

```bash
python .claude/hooks/diagnose.py
```

### 常见问题

#### 1. 环境变量未设置 ❌

**症状**: `debug.log` 显示 `ERROR: Missing env vars`

**解决方法**:

```bash
# Windows CMD
set PUSHOVER_TOKEN=your_app_token
set PUSHOVER_USER=your_user_key

# Windows PowerShell
$env:PUSHOVER_TOKEN="your_app_token"
$env:PUSHOVER_USER="your_user_key"

# Linux/Mac
export PUSHOVER_TOKEN=your_app_token
export PUSHOVER_USER=your_user_key
```

#### 2. Hook 未触发

**症状**: `debug.log` 不存在或为空

**检查**:
- `.claude/settings.json` 是否在项目根目录
- 脚本路径是否正确
- 运行诊断脚本验证配置

#### 3. API 调用失败

**症状**: `debug.log` 显示 HTTP 400/401 错误

**检查**:
- Token 是否有效: https://pushover.net/apps
- User Key 是否正确: https://pushover.net/
- API 是否启用（确保是 API Token，不是仅 SDK）

## 测试通知

运行测试脚本发送测试通知：

```bash
python .claude/hooks/test-pushover.py
```

## 部署到新项目

1. 复制整个 `.claude` 文件夹到目标项目
2. 设置环境变量 `PUSHOVER_TOKEN` 和 `PUSHOVER_USER`
3. 运行诊断脚本验证配置
4. 触发一个 Claude Code 任务测试

## 日志位置

- 调试日志: `.claude/hooks/debug.log`
- 会话缓存: `.claude/cache/session-*.jsonl`
