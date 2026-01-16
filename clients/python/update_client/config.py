"""配置管理"""
from dataclasses import dataclass
from typing import Optional


@dataclass
class Config:
    """客户端配置"""
    server_url: str
    program_id: str
    channel: str = "stable"
    timeout: int = 30
    max_retries: int = 3
    save_path: str = "./updates"

    @classmethod
    def default(cls) -> "Config":
        """返回默认配置"""
        return cls(
            server_url="http://localhost:8080",
            channel="stable",
            timeout=30,
            max_retries=3,
            save_path="./updates"
        )
