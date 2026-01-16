"""DocuFiller Update Client SDK"""
from .config import Config
from .types import UpdateInfo, DownloadProgress, ProgressCallback, UpdateError
from .version import compare_versions
from .checker import UpdateChecker

__version__ = "1.0.0"
__all__ = [
    "Config",
    "UpdateInfo",
    "DownloadProgress",
    "ProgressCallback",
    "UpdateError",
    "compare_versions",
    "UpdateChecker",
]
