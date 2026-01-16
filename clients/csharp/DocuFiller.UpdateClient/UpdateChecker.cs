using System.Runtime.CompilerServices;
using System.Security.Cryptography;
using System.Text;
using System.Text.Json;

namespace DocuFiller.UpdateClient;

/// <summary>
/// 更新检查器
/// </summary>
public class UpdateChecker
{
    private readonly Config _config;
    private readonly HttpClient _httpClient;

    public UpdateChecker(Config config)
    {
        _config = config;
        _httpClient = new HttpClient
        {
            Timeout = TimeSpan.FromSeconds(config.Timeout)
        };
    }

    /// <summary>
    /// 检查是否有新版本
    /// </summary>
    public async Task<UpdateInfo?> CheckUpdateAsync(string currentVersion, CancellationToken cancellationToken = default)
    {
        var url = $"{_config.ServerUrl}/api/programs/{_config.ProgramId}/versions/latest?channel={_config.Channel}";

        try
        {
            var response = await _httpClient.GetAsync(url, cancellationToken);

            if (response.StatusCode == System.Net.HttpStatusCode.NotFound)
            {
                throw new UpdateError("NO_VERSION", "No version found for this program");
            }

            response.EnsureSuccessStatusCode();

            var json = await response.Content.ReadAsStringAsync(cancellationToken);
            var info = JsonSerializer.Deserialize<UpdateInfo>(json, new JsonSerializerOptions
            {
                PropertyNameCaseInsensitive = true
            });

            if (info == null) return null;

            // 检查是否是新版本
            if (VersionComparer.Compare(info.Version, currentVersion) <= 0)
            {
                return null; // 没有新版本
            }

            return info;
        }
        catch (HttpRequestException ex)
        {
            throw new UpdateError("NETWORK_ERROR", $"Failed to connect to server: {ex.Message}");
        }
    }

    /// <summary>
    /// 下载更新包
    /// </summary>
    public async Task DownloadUpdateAsync(
        string version,
        string destPath,
        IProgress<DownloadProgress>? progress = null,
        CancellationToken cancellationToken = default)
    {
        var lastError = new Exception?();

        for (int attempt = 0; attempt <= _config.MaxRetries; attempt++)
        {
            if (attempt > 0)
            {
                await Task.Delay(attempt * 2000, cancellationToken); // 指数退避
            }

            try
            {
                await DownloadOnceAsync(version, destPath, progress, cancellationToken);
                return; // 成功
            }
            catch (Exception ex)
            {
                lastError = ex;
            }
        }

        throw lastError!;
    }

    private async Task DownloadOnceAsync(
        string version,
        string destPath,
        IProgress<DownloadProgress>? progress,
        CancellationToken cancellationToken)
    {
        var url = $"{_config.ServerUrl}/api/download/{_config.ProgramId}/{_config.Channel}/{version}";

        var response = await _httpClient.GetAsync(url, HttpCompletionOption.ResponseHeadersRead, cancellationToken);
        response.EnsureSuccessStatusCode();

        var total = response.Content.Headers.ContentLength ?? 0;
        var directory = Path.GetDirectoryName(destPath);
        if (!string.IsNullOrEmpty(directory))
        {
            Directory.CreateDirectory(directory);
        }

        using var contentStream = await response.Content.ReadAsStreamAsync(cancellationToken);
        using var fileStream = File.Create(destPath);

        var downloaded = 0L;
        var startTime = DateTime.UtcNow;
        var buffer = new byte[32 * 1024];

        int bytesRead;
        while ((bytesRead = await contentStream.ReadAsync(buffer, cancellationToken)) > 0)
        {
            await fileStream.WriteAsync(buffer.AsMemory(0, bytesRead), cancellationToken);
            downloaded += bytesRead;

            progress?.Report(new DownloadProgress
            {
                Version = version,
                Downloaded = downloaded,
                Total = total,
                Percentage = total > 0 ? (double)downloaded / total * 100 : 0,
                Speed = CalculateSpeed(downloaded, startTime)
            });
        }
    }

    private static double CalculateSpeed(long downloaded, DateTime startTime)
    {
        var elapsed = (DateTime.UtcNow - startTime).TotalSeconds;
        return elapsed > 0 ? downloaded / elapsed : 0;
    }

    /// <summary>
    /// 验证文件 SHA256 哈希
    /// </summary>
    public bool VerifyFile(string filePath, string expectedHash)
    {
        using var sha256 = SHA256.Create();
        using var stream = File.OpenRead(filePath);

        var hash = sha256.ComputeHash(stream);
        var actualHash = Convert.ToHexString(hash).ToLowerInvariant();

        return actualHash == expectedHash.ToLowerInvariant();
    }
}
