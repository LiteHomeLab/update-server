"""更新检查器"""
import time
import hashlib
import os
from typing import Optional
from pathlib import Path
from datetime import datetime

import requests

from .config import Config
from .types import UpdateInfo, DownloadProgress, ProgressCallback, UpdateError
from .version import compare_versions


class UpdateChecker:
    """更新检查器"""

    def __init__(self, config: Config):
        """
        初始化更新检查器

        Args:
            config: 客户端配置
        """
        self.config = config
        self.session = requests.Session()
        self.session.timeout = config.timeout

    def check_update(self, current_version: str) -> Optional[UpdateInfo]:
        """
        检查是否有新版本

        Args:
            current_version: 当前版本号

        Returns:
            UpdateInfo 如果有新版本，否则 None
        """
        url = f"{self.config.server_url}/api/programs/{self.config.program_id}/versions/latest"
        params = {"channel": self.config.channel}

        try:
            response = self.session.get(url, params=params)

            if response.status_code == 404:
                raise UpdateError("NO_VERSION", "No version found for this program")

            response.raise_for_status()

            data = response.json()
            info = UpdateInfo(
                version=data["version"],
                channel=data["channel"],
                file_name=data["fileName"],
                file_size=data["fileSize"],
                file_hash=data["fileHash"],
                release_notes=data["releaseNotes"],
                publish_date=datetime.fromisoformat(data["publishDate"].replace("Z", "+00:00")),
                mandatory=data["mandatory"],
                download_count=data["downloadCount"]
            )

            # 检查是否是新版本
            if compare_versions(info.version, current_version) <= 0:
                return None

            return info

        except requests.RequestException as e:
            raise UpdateError("NETWORK_ERROR", f"Failed to connect to server: {e}")

    def download_update(
        self,
        version: str,
        dest_path: str,
        progress_callback: Optional[ProgressCallback] = None
    ) -> None:
        """
        下载更新包（带重试）

        Args:
            version: 版本号
            dest_path: 目标路径
            progress_callback: 进度回调
        """
        last_error = None

        for attempt in range(self.config.max_retries + 1):
            if attempt > 0:
                time.sleep(attempt * 2)  # 指数退避

            try:
                self._download_once(version, dest_path, progress_callback)
                return  # 成功
            except Exception as e:
                last_error = e

        raise last_error

    def _download_once(
        self,
        version: str,
        dest_path: str,
        progress_callback: Optional[ProgressCallback] = None
    ) -> None:
        """单次下载尝试"""
        url = f"{self.config.server_url}/api/download/{self.config.program_id}/{self.config.channel}/{version}"

        response = self.session.get(url, stream=True)
        response.raise_for_status()

        # 确保目录存在
        dest_path_obj = Path(dest_path)
        dest_path_obj.parent.mkdir(parents=True, exist_ok=True)

        total = int(response.headers.get("content-length", 0))
        downloaded = 0
        start_time = time.time()

        with open(dest_path, "wb") as f:
            for chunk in response.iter_content(chunk_size=32 * 1024):
                if chunk:
                    f.write(chunk)
                    downloaded += len(chunk)

                    # 进度回调
                    if progress_callback:
                        elapsed = time.time() - start_time
                        speed = downloaded / elapsed if elapsed > 0 else 0

                        progress_callback(DownloadProgress(
                            version=version,
                            downloaded=downloaded,
                            total=total,
                            percentage=(downloaded / total * 100) if total > 0 else 0,
                            speed=speed
                        ))

    def verify_file(self, file_path: str, expected_hash: str) -> bool:
        """
        验证文件 SHA256 哈希

        Args:
            file_path: 文件路径
            expected_hash: 期望的哈希值

        Returns:
            是否匹配
        """
        sha256_hash = hashlib.sha256()

        with open(file_path, "rb") as f:
            for chunk in iter(lambda: f.read(4096), b""):
                sha256_hash.update(chunk)

        actual_hash = sha256_hash.hexdigest()
        return actual_hash.lower() == expected_hash.lower()
