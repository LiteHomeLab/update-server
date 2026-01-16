"""版本比较工具"""
from typing import List


def compare_versions(v1: str, v2: str) -> int:
    """
    比较两个版本号

    Args:
        v1: 版本号 1
        v2: 版本号 2

    Returns:
        -1 (v1 < v2), 0 (v1 == v2), 1 (v1 > v2)
    """
    # 移除 v 前缀
    v1 = v1.lstrip('v')
    v2 = v2.lstrip('v')

    parts1 = _parse_version(v1)
    parts2 = _parse_version(v2)

    max_len = max(len(parts1), len(parts2))

    for i in range(max_len):
        p1 = parts1[i] if i < len(parts1) else 0
        p2 = parts2[i] if i < len(parts2) else 0

        if p1 > p2:
            return 1
        if p1 < p2:
            return -1

    return 0


def _parse_version(version: str) -> List[int]:
    """解析版本号为数字列表"""
    parts = version.split('.')
    result = []

    for part in parts:
        try:
            result.append(int(part))
        except ValueError:
            result.append(0)

    return result
