"""UpdateChecker 测试"""
import pytest
from update_client import Config, UpdateChecker
from update_client.version import compare_versions


def test_new_checker():
    """测试创建 UpdateChecker"""
    config = Config.default()
    config.program_id = "testapp"

    checker = UpdateChecker(config)

    assert checker.config.program_id == "testapp"
    assert checker.config.channel == "stable"
