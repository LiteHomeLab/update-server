# C# Update Client SDK

DocuFiller 更新服务器的 C# 客户端 SDK。

## 安装

```bash
dotnet add reference DocuFiller.UpdateClient
```

## 使用示例

### 检查更新

```csharp
using DocuFiller.UpdateClient;

var config = new Config
{
    ProgramId = "myapp",
    ServerUrl = "http://localhost:8080"
};

var checker = new UpdateChecker(config);

var info = await checker.CheckUpdateAsync("1.0.0");
if (info == null)
{
    Console.WriteLine("No update available");
}
else
{
    Console.WriteLine($"New version available: {info.Version}");
}
```

### 下载更新

```csharp
var progress = new Progress<DownloadProgress>(p =>
{
    Console.WriteLine($"Downloaded: {p.Downloaded}/{p.Total} ({p.Percentage:F1}%)");
});

await checker.DownloadUpdateAsync(info.Version, "./update.zip", progress);

// 验证文件
if (checker.VerifyFile("./update.zip", info.FileHash))
{
    Console.WriteLine("Update downloaded and verified!");
}
```
