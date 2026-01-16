"""CLI 工具"""
import click
from .admin import UpdateAdmin


@click.group()
@click.option("--server", default="http://localhost:8080", help="Server URL")
@click.option("--token", required=True, help="API token")
@click.option("--program-id", required=True, help="Program ID")
@click.pass_context
def cli(ctx, server, token, program_id):
    """DocuFiller Update Server Admin Tool"""
    ctx.ensure_object(dict)
    ctx.obj["admin"] = UpdateAdmin(server, token)
    ctx.obj["program_id"] = program_id


@cli.command()
@click.option("--channel", required=True, help="Channel (stable/beta)")
@click.option("--version", required=True, help="Version number")
@click.option("--file", required=True, help="File path", type=click.Path(exists=True))
@click.option("--notes", default="", help="Release notes")
@click.option("--mandatory", is_flag=True, help="Mandatory update")
@click.pass_context
def upload(ctx, channel, version, file, notes, mandatory):
    """Upload a new version"""
    admin = ctx.obj["admin"]
    program_id = ctx.obj["program_id"]

    admin.upload_version(program_id, channel, version, file, notes, mandatory)


@cli.command()
@click.option("--channel", required=True, help="Channel (stable/beta)")
@click.option("--version", required=True, help="Version number")
@click.pass_context
def delete(ctx, channel, version):
    """Delete a version"""
    admin = ctx.obj["admin"]
    program_id = ctx.obj["program_id"]

    admin.delete_version(program_id, channel, version)


@cli.command()
@click.option("--channel", help="Channel filter (stable/beta)")
@click.pass_context
def list(ctx, channel):
    """List versions"""
    admin = ctx.obj["admin"]
    program_id = ctx.obj["program_id"]

    versions = admin.list_versions(program_id, channel)

    for v in versions:
        print(f"{v['version']} ({v['channel']}) - {v['publishDate'][:10]} - {v['fileSize']} bytes")


if __name__ == "__main__":
    cli()
