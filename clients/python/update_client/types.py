"""数据类型定义"""
from dataclasses import dataclass
from datetime import datetime
from typing import Callable, Optional


@dataclass
class UpdateInfo:
    """更新信息"""
    version: str
    channel: str
    file_name: str
    file_size: int
    file_hash: str
    release_notes: str
    publish_date: datetime
    mandatory: bool
    download_count: int


@dataclass
class DownloadProgress:
    """下载进度"""
    version: str
    downloaded: int
    total: int
    percentage: float
    speed: float  # bytes/second


ProgressCallback = Callable[[DownloadProgress], None]


class UpdateError(Exception):
    """更新错误"""
    def __init__(self, code: str, message: str):
        self.code = code
        self.message = message
        super().__init__(f"[{code}] {message}")
