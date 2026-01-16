"""版本比较测试"""
import pytest
from update_client.version import compare_versions


def test_equal_versions():
    """测试相等版本"""
    assert compare_versions("1.0.0", "1.0.0") == 0


def test_v_prefix():
    """测试 v 前缀"""
    assert compare_versions("v1.0.0", "1.0.0") == 0


def test_v1_greater():
    """测试 v1 大于 v2"""
    assert compare_versions("1.2.0", "1.1.0") == 1


def test_v2_greater():
    """测试 v2 大于 v1"""
    assert compare_versions("1.0.0", "2.0.0") == -1


def test_three_parts():
    """测试三位版本号"""
    assert compare_versions("1.2.3", "1.2.2") == 1


def test_four_parts():
    """测试四位版本号"""
    assert compare_versions("1.2.3.4", "1.2.3.3") == 1
