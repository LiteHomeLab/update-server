namespace DocuFiller.UpdateClient;

/// <summary>
/// 版本比较器
/// </summary>
public static class VersionComparer
{
    /// <summary>
    /// 比较两个版本号
    /// </summary>
    /// <returns>-1 (v1 < v2), 0 (v1 == v2), 1 (v1 > v2)</returns>
    public static int Compare(string v1, string v2)
    {
        // 移除 v 前缀
        v1 = v1.TrimStart('v');
        v2 = v2.TrimStart('v');

        var parts1 = ParseVersion(v1);
        var parts2 = ParseVersion(v2);

        var maxLen = Math.Max(parts1.Count, parts2.Count);

        for (int i = 0; i < maxLen; i++)
        {
            int p1 = i < parts1.Count ? parts1[i] : 0;
            int p2 = i < parts2.Count ? parts2[i] : 0;

            if (p1 > p2) return 1;
            if (p1 < p2) return -1;
        }

        return 0;
    }

    private static List<int> ParseVersion(string version)
    {
        var parts = version.Split('.');
        var result = new List<int>();

        foreach (var part in parts)
        {
            if (int.TryParse(part, out int num))
            {
                result.Add(num);
            }
            else
            {
                result.Add(0);
            }
        }

        return result;
    }
}
