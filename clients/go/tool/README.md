# Update Admin Tool

用于 DocuFiller 更新服务器的命令行管理工具。

## 编译

```bash
go build -ldflags "-s -w" -o update-admin.exe .
```

## 配置

### 环境变量（推荐）

```cmd
setx UPDATE_SERVER_URL "http://your-server:8080"
setx UPDATE_TOKEN "your-api-token"
```

### 命令行参数

- `--server url`：服务器地址（覆盖环境变量）
- `--token value`：认证令牌（覆盖环境变量）
- `--program-id id`：程序标识符（**必须指定**）

## 使用示例

### 上传新版本

```cmd
update-admin.exe upload --program-id myapp --channel stable --version 1.0.0 --file myapp.zip --notes "Initial release"
```

### 列出版本

```cmd
update-admin.exe list --program-id myapp --channel stable
```

### 删除版本

```cmd
update-admin.exe delete --program-id myapp --version 1.0.0 --channel stable
```

## 命令参数

### upload 命令

| 参数 | 说明 |
|------|------|
| `--channel` | 发布通道（stable/beta） |
| `--version` | 版本号 |
| `--file` | 文件路径 |
| `--notes` | 发布说明（可选） |
| `--mandatory` | 强制更新标记（可选） |

### list 命令

| 参数 | 说明 |
|------|------|
| `--channel` | 通道过滤（可选） |

### delete 命令

| 参数 | 说明 |
|------|------|
| `--version` | 版本号 |
| `--channel` | 通道（stable/beta，默认：stable） |

## CI/CD 集成

### GitHub Actions 示例

```yaml
- name: Upload Release
  run: |
    ./update-admin.exe upload \
      --program-id myapp \
      --channel stable \
      --version ${{ github.ref_name }} \
      --file ./dist/myapp.zip \
      --notes "Release ${{ github.ref_name }}"
  env:
    UPDATE_SERVER_URL: ${{ secrets.UPDATE_SERVER_URL }}
    UPDATE_TOKEN: ${{ secrets.UPDATE_TOKEN }}
```
