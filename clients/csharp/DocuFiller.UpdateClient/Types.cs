using System.Text.Json.Serialization;

namespace DocuFiller.UpdateClient;

/// <summary>
/// 更新信息
/// </summary>
public record UpdateInfo
{
    [JsonPropertyName("programId")]
    public string ProgramId { get; init; } = string.Empty;

    [JsonPropertyName("version")]
    public string Version { get; init; } = string.Empty;

    [JsonPropertyName("channel")]
    public string Channel { get; init; } = string.Empty;

    [JsonPropertyName("fileName")]
    public string FileName { get; init; } = string.Empty;

    [JsonPropertyName("fileSize")]
    public long FileSize { get; init; }

    [JsonPropertyName("fileHash")]
    public string FileHash { get; init; } = string.Empty;

    [JsonPropertyName("releaseNotes")]
    public string ReleaseNotes { get; init; } = string.Empty;

    [JsonPropertyName("publishDate")]
    public DateTime PublishDate { get; init; }

    [JsonPropertyName("mandatory")]
    public bool Mandatory { get; init; }

    [JsonPropertyName("downloadCount")]
    public int DownloadCount { get; init; }
}

/// <summary>
/// 下载进度
/// </summary>
public record DownloadProgress
{
    public string Version { get; init; } = string.Empty;
    public long Downloaded { get; init; }
    public long Total { get; init; }
    public double Percentage { get; init; }
    public double Speed { get; init; } // bytes/second
}

/// <summary>
/// 更新错误
/// </summary>
public class UpdateError : Exception
{
    public string Code { get; }

    public UpdateError(string code, string message) : base($"[{code}] {message}")
    {
        Code = code;
    }
}
