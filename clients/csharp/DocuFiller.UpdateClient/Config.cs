namespace DocuFiller.UpdateClient;

/// <summary>
/// 客户端配置
/// </summary>
public class Config
{
    /// <summary>
    /// 服务器地址
    /// </summary>
    public string ServerUrl { get; set; } = "http://localhost:8080";

    /// <summary>
    /// 程序 ID
    /// </summary>
    public string ProgramId { get; set; } = string.Empty;

    /// <summary>
    /// 发布渠道
    /// </summary>
    public string Channel { get; set; } = "stable";

    /// <summary>
    /// 请求超时时间（秒）
    /// </summary>
    public int Timeout { get; set; } = 30;

    /// <summary>
    /// 最大重试次数
    /// </summary>
    public int MaxRetries { get; set; } = 3;

    /// <summary>
    /// 下载保存路径
    /// </summary>
    public string SavePath { get; set; } = "./updates";

    /// <summary>
    /// 创建默认配置
    /// </summary>
    public static Config Default() => new();
}
