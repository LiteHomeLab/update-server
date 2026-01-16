# Python Update Client SDK

DocuFiller 更新服务器的 Python 客户端 SDK。

## 安装

```bash
pip install -r requirements.txt
```

## 使用示例

### 检查更新

```python
from update_client import Config, UpdateChecker

config = Config.default()
config.program_id = "myapp"
config.server_url = "http://localhost:8080"

checker = UpdateChecker(config)

info = checker.check_update("1.0.0")
if info is None:
    print("No update available")
else:
    print(f"New version available: {info.version}")
```

### 下载更新

```python
def progress_callback(p):
    print(f"Downloaded: {p.downloaded}/{p.total} ({p.percentage:.1f}%)")

checker.download_update(info.version, "./update.zip", progress_callback)

# 验证文件
if checker.verify_file("./update.zip", info.file_hash):
    print("Update downloaded and verified!")
```

## 管理工具

```bash
# 上传新版本
python -m update_admin.cli upload --program-id myapp --channel stable \
  --version 1.0.0 --file ./myapp-1.0.0.zip --notes "Initial release"

# 列出版本
python -m update_admin.cli list --program-id myapp --channel stable

# 删除版本
python -m update_admin.cli delete --program-id myapp --channel stable --version 1.0.0
```
