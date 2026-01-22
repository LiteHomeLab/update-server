# 管理后台核心功能设计文档

**日期**: 2025-01-22
**作者**: Claude & 用户协作设计
**状态**: 设计完成，待实施

---

## 概述

本文档描述了 Update Server 管理后台的核心功能设计，目标是实现最基本但完整的应用程序管理功能，包括：
1. 创建应用程序
2. 生成和管理 Token 与加密密钥
3. 打包下载配置好的客户端工具
4. 显示命令行工具的使用示例

---

## 整体架构

### 布局结构

管理后台采用**左右分栏布局**：

**左侧边栏**：
- 顶部：服务器名称 + "创建新程序"按钮
- 下方：程序列表（卡片式），每个显示程序名称、ID、图标
- 点击程序卡片切换右侧视图

**右侧主内容区**：
- 分为三个水平排列的模块卡片：
  1. **Token & 密钥管理**
  2. **版本管理**
  3. **客户端工具**

---

## 模块一：左侧边栏 - 程序列表

### 程序列表卡片

每个程序卡片显示：
- 程序名称（大字）
- 程序ID（小字灰色）
- 选中状态：高亮边框

### 创建新程序

点击"创建新程序"按钮后：

**模态框内容**：
- 标题："创建新程序"
- 单个输入框：程序名称（必填）
- 说明："程序ID将自动生成，例如：my-app-20250122-abc123"
- 按钮：[取消] [创建]

**创建流程**：
1. 用户输入程序名称，如"DocuFiller"
2. 点击"创建"，调用 `POST /api/admin/programs`
3. 后端返回完整的程序信息
4. 前端自动将新程序添加到列表并选中

---

## 模块二：Token & 密钥管理

### 布局结构

**Upload Token 区域**：
- 标题："Upload Token"（用于发布更新）
- 显示：`upload-xxxxx-xxxxx` 格式的 token
- 按钮：[复制] [重新生成]
- 说明："update-publisher 发布时使用"

**Download Token 区域**：
- 标题："Download Token"（用于客户端下载）
- 显示：`download-xxxxx-xxxxx` 格式的 token
- 按钮：[复制] [重新生成]
- 说明："update-client 检查更新时使用"

**加密密钥区域**：
- 标题："加密密钥"（用于签名验证）
- 显示：32字节的十六进制字符串
- 按钮：[复制] [重新生成]
- 警告："重新生成后旧版本将无法验证！"

### API 调用

- 查询当前 token：`GET /api/admin/programs/:programId`
- 重新生成 token：`POST /api/admin/programs/:programId/tokens/regenerate?type=upload|download`
- 重新生成密钥：`POST /api/admin/programs/:programId/encryption/regenerate`

---

## 模块三：版本管理

### 布局结构

**版本列表**：
- 表格形式，包含列：版本号、通道（stable/beta）、文件大小、发布时间、下载次数、操作
- 最新版本高亮显示
- 每行有删除按钮

**上传新版本区域**：
- 按钮："上传新版本"
- 点击后弹出模态框，包含：
  - 文件选择（拖拽或点击）
  - 版本号输入
  - 通道选择（stable/beta）
  - 发布说明文本框
  - 强制更新开关
  - 上传进度条

### API 调用

- 列表加载：`GET /api/admin/programs/:programId/versions`
- 删除版本：`DELETE /api/admin/programs/:programId/versions/:version`
- 上传版本：`POST /api/programs/:programId/versions`（需要 upload token 认证）

---

## 模块四：客户端工具

### 布局结构

**Update Client（更新客户端）区域**：
- 标题："Update Client - 嵌入到你的应用中"
- 下载按钮：`[下载已配置的客户端包]`
- 点击后调用：`GET /api/admin/programs/:programId/client/update`
- 返回 ZIP 包，包含：
  - `update-client.exe` - 可执行程序
  - `config.yaml` - 已配置好服务器地址、programID、download token
- 使用说明：简短的集成指南

**Update Publisher（发布工具）区域**：
- 标题："Update Publisher - 命令行发布工具"
- 说明："此工具为命令行程序，无法打包。请使用以下命令："
- 代码框显示命令示例：

```bash
# 上传新版本
./update-publisher upload \
  --server=http://your-server:8080 \
  --program-id=your-program-id \
  --token=upload-xxxxx-xxxxx \
  --file=./your-app-v1.0.0.zip \
  --version=1.0.0 \
  --channel=stable \
  --notes="修复了若干bug"
```

### 动态内容

所有示例中的 token、programID、服务器地址都从当前程序的真实数据动态填充。

### API 调用

- 下载 update client：`GET /api/admin/programs/:programId/client/update`
- 下载 publish client：`GET /api/admin/programs/:programId/client/publish`

---

## 前端实现

### 状态管理

```javascript
let currentProgram = null;  // 当前选中的程序
let programList = [];       // 程序列表
```

### 核心函数

**加载程序列表**：
```javascript
async function loadPrograms() {
    const response = await fetch('/api/admin/programs');
    programList = await response.json();
    renderProgramList();
    if (!currentProgram && programList.length > 0) {
        selectProgram(programList[0].programId);
    }
}
```

**选择程序**：
```javascript
async function selectProgram(programId) {
    currentProgram = programId;
    const response = await fetch(`/api/admin/programs/${programId}`);
    const data = await response.json();
    renderProgramDetail(data);
    updateCommandExamples(data);
}
```

**更新命令示例**：
```javascript
function updateCommandExamples(data) {
    const serverUrl = window.location.origin;
    const uploadToken = data.uploadToken;
    const programId = data.program.programId;

    const command = `./update-publisher upload \\
  --server=${serverUrl} \\
  --program-id=${programId} \\
  --token=${uploadToken} \\
  --file=./your-app.zip \\
  --version=1.0.0 \\
  --channel=stable`;

    document.getElementById('publisher-command').textContent = command;
}
```

---

## 错误处理和用户体验

### 错误处理

- 所有 fetch 调用添加 `.catch()` 处理网络错误
- 检查 `response.ok`，非 2xx 状态显示错误消息
- 使用 toast 通知替代 `alert()`

### Toast 通知

- 成功操作：绿色提示，2秒后自动消失
- 错误操作：红色提示，5秒后自动消失
- 警告操作：黄色提示，需手动关闭

### 加载状态

- 操作进行中显示加载动画
- 禁用相关按钮防止重复点击

---

## HTML/CSS 修改

### HTML 修改

现有 `admin.html` 需要修改：
1. 侧边栏改为程序列表
2. 移除现有的"仪表板"、"文件管理"等无用导航
3. 主内容区改为三个模块卡片布局

### 新增 CSS 样式

需要添加：
- 程序列表卡片样式
- 三个模块卡片的网格布局
- Token/密钥区域的样式
- 命令示例代码框样式
- Toast 通知样式

---

## 数据流程图

```
用户点击"创建新程序"
    ↓
输入程序名称
    ↓
POST /api/admin/programs
    ↓
后端创建程序记录，生成 tokens 和密钥
    ↓
返回程序信息
    ↓
前端添加到列表并选中
    ↓
显示三个模块：
    - Token & 密钥管理
    - 版本管理
    - 客户端工具
```

---

## 实施计划

1. **第一阶段**：修改 HTML 结构和 CSS 样式
2. **第二阶段**：实现前端 JavaScript 逻辑
3. **第三阶段**：测试和修复

---

## 备注

- 本设计遵循 YAGNI 原则，只实现最核心的功能
- 后续可以根据需求扩展更多功能
- 所有 API 端点已在后端实现，无需修改后端代码
