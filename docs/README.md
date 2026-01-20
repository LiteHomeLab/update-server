# Update Server 快速指南

## 一、部署服务器

1. 解压 `update-server.zip` 到服务器目录
2. 运行 `update-server.exe`
3. 浏览器自动打开初始化向导，按提示完成配置：
   - 创建管理员账号
   - 设置服务端口和服务器 URL
4. 完成后自动进入管理后台

## 二、创建程序

1. 登录管理后台 → 程序管理 → 创建新程序
2. 填写 Program ID（如：docufiller）
3. 系统自动生成密钥和 Token
4. 记录显示的信息（或稍后在程序详情页查看）

## 三、发布更新

1. 在程序详情页下载 `docufiller-publish-client.zip`
2. 解压并编辑 `publish-config.yaml`
3. 运行 `publish-client.exe`

## 四、集成更新功能

1. 在程序详情页下载 `docufiller-update-client.zip`
2. 解压到您的项目
3. 程序启动时执行 `update-client.exe --check`
4. 根据返回结果处理更新