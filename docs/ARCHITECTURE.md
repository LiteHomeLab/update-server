# Update Server 架构说明

## 系统架构

三层架构：
- **update-server**：中央更新服务器（Go + Gin）
- **publish-client**：发布端（Go CLI）
- **update-client**：更新端（Go CLI）

## 认证机制

- Token 认证：Upload Token / Download Token
- Token 在 Web 界面生成和管理
- 通过 HTTP Header 传递

## 加密机制

端到端加密：
- 每个程序独立的 32 字节密钥
- AES-256-GCM 加密算法
- publish-client 上传前加密
- update-client 下载后解密
- 服务器只存储加密文件

## 数据模型

- programs：程序元数据
- versions：版本信息
- tokens：认证令牌
- encryption_keys：加密密钥
- admin_users：管理员账号

## 技术栈

- 后端：Go 1.23+ / Gin / GORM / SQLite
- 前端：嵌入式单页应用
- 加密：AES-256-GCM / HKDF