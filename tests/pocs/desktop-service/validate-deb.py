"""Validate the generated P04 GUI deb, without installing or extracting it.

Creator: HIHU-COCO. Created: 2026-10-19.
"""
from pathlib import Path, PurePosixPath
import re
import shlex
import subprocess
import sys
import tarfile


def validate(package: Path) -> None:
    metadata = subprocess.check_output(
        ["dpkg-deb", "--field", str(package)], text=True
    )
    fields = dict(
        line.split(": ", 1) for line in metadata.splitlines() if ": " in line
    )
    if not re.fullmatch(r"[a-z0-9][a-z0-9+.-]+", fields.get("Package", "")):
        raise ValueError("invalid Debian package identifier")
    if fields.get("Architecture") != "amd64":
        raise ValueError("expected an amd64 P04 package")
    if "libwebkit2gtk-4.1-0" not in fields.get("Depends", ""):
        raise ValueError("GUI package must declare WebKitGTK 4.1")

    binaries = set()
    desktops = []
    icons = set()
    process = subprocess.Popen(
        ["dpkg-deb", "--fsys-tarfile", str(package)], stdout=subprocess.PIPE
    )
    try:
        with tarfile.open(fileobj=process.stdout, mode="r|") as archive:
            for member in archive:
                path = PurePosixPath(member.name)
                if path.is_absolute() or ".." in path.parts:
                    raise ValueError("unsafe package member path")
                name = path.as_posix().removeprefix("./")
                if member.isfile() and name.startswith("usr/bin/"):
                    stream = archive.extractfile(member)
                    if stream is None or stream.read(4) != b"\x7fELF":
                        raise ValueError("GUI executable is not an ELF binary")
                    if member.mode & 0o111 == 0:
                        raise ValueError("GUI binary lacks executable permission")
                    binaries.add(path.name)
                if member.isfile() and name.startswith("usr/share/applications/") and name.endswith(".desktop"):
                    stream = archive.extractfile(member)
                    if stream is None or member.size > 65536:
                        raise ValueError("invalid desktop entry")
                    desktops.append(stream.read().decode("utf-8"))
                if member.isfile() and name.startswith("usr/share/icons/"):
                    icons.add(path.stem)
        if process.wait() != 0:
            raise ValueError("dpkg-deb could not read package payload")
    finally:
        if process.stdout:
            process.stdout.close()
        if process.poll() is None:
            process.kill()
            process.wait()

    if len(desktops) != 1:
        raise ValueError("expected exactly one GUI desktop entry")
    entry = dict(line.split("=", 1) for line in desktops[0].splitlines() if "=" in line and not line.startswith("#"))
    if entry.get("Type") != "Application" or "跃点" not in entry.get("Name", ""):
        raise ValueError("desktop entry must retain the Chinese product name")
    command = shlex.split(entry.get("Exec", ""))
    if not command or PurePosixPath(command[0]).name not in binaries:
        raise ValueError("desktop entry does not launch a packaged executable")
    if entry.get("Icon") not in icons:
        raise ValueError("desktop entry icon is missing from the package")
    print(f"PASS: {fields['Package']} {fields['Version']}; ELF, WebKitGTK dependency, Chinese desktop entry and icon verified")


if __name__ == "__main__":
    if len(sys.argv) != 2:
        raise SystemExit("usage: validate-deb.py PACKAGE.deb")
    validate(Path(sys.argv[1]))
