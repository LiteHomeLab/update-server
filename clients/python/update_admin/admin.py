"""管理客户端"""
import requests


class UpdateAdmin:
    """更新服务器管理客户端"""

    def __init__(self, server_url: str, token: str):
        self.server_url = server_url
        self.token = token
        self.session = requests.Session()
        self.session.headers.update({
            "Authorization": f"Bearer {token}"
        })

    def upload_version(
        self,
        program_id: str,
        channel: str,
        version: str,
        file_path: str,
        notes: str = "",
        mandatory: bool = False
    ) -> None:
        """上传新版本"""
        url = f"{self.server_url}/api/version/{program_id}/upload"

        with open(file_path, "rb") as f:
            files = {"file": f}
            data = {
                "channel": channel,
                "version": version,
                "notes": notes,
                "mandatory": str(mandatory).lower()
            }

            response = self.session.post(url, files=files, data=data)
            response.raise_for_status()

        print(f"Version {program_id}/{channel}/{version} uploaded successfully")

    def delete_version(self, program_id: str, channel: str, version: str) -> None:
        """删除版本"""
        url = f"{self.server_url}/api/version/{program_id}/{channel}/{version}"

        response = self.session.delete(url)
        response.raise_for_status()

        print(f"Version {program_id}/{channel}/{version} deleted successfully")

    def list_versions(self, program_id: str, channel: str = None) -> list:
        """列出版本"""
        url = f"{self.server_url}/api/version/{program_id}/list"
        params = {}
        if channel:
            params["channel"] = channel

        response = self.session.get(url, params=params)
        response.raise_for_status()

        data = response.json()
        return data.get("versions", [])
