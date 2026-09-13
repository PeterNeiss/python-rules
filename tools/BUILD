_PLEASE_PEX_VERSION = "3.0.2"

# Two platforms come from this fork rather than from upstream.
#
# windows_amd64, because upstream publishes no Windows please_pex at all, and without one a
# native Windows plz cannot build a Python target.
#
# linux_amd64, because please_pex is a *host* tool: a Linux machine cross-compiling to Windows
# uses the Linux build, and the released Linux build has no Windows preamble in it. A .pex is a
# zip with an executable stub in front, and the stub upstream prepends is an ELF binary that
# Windows will not run. Without this, every repo cross-building a pex for Windows has to point
# PexTool at this plugin's source instead.
#
# Both go away when upstream ships a please_pex carrying the preamble.
_OWNERS = {
    "linux_amd64": "PeterNeiss",
    "windows_amd64": "PeterNeiss",
}
_PLATFORM = f"{CONFIG.OS}_{CONFIG.ARCH}"
_OWNER = _OWNERS.get(_PLATFORM, "please-build")

# The Windows build carries fixes the other platforms' builds do not need, so it is released under
# its own version rather than replacing an asset earlier commits pin by hash.
_VERSIONS = {"windows_amd64": "3.0.2-windows.2"}
_VERSION = _VERSIONS.get(_PLATFORM, _PLEASE_PEX_VERSION)

remote_file(
    name = "please_pex",
    url = f"https://github.com/{_OWNER}/python-rules/releases/download/please_pex-v{_VERSION}/please_pex-{_VERSION}-{_PLATFORM}",
    # The downloaded name carries the version and platform, which on Windows leaves it with no
    # extension in PATHEXT - and Go's exec will not run such a file even given its full path.
    out = "please_pex.exe" if CONFIG.OS == "windows" else None,
    hashes = [
        "b9049a96aaead27f83e713d4d9f60b2b9a30411c7479722bec0e67254016b4a9",  # darwin_amd64
        "ba5868b241dc136509d827db6a6a65b71a5b9d003f9dbf439777467d5cd3ebd5",  # darwin_arm64
        "26f8624d81036f3672a2280e6937b04ef1a54ebde041f61e5b34b6837df24817",  # freebsd_amd64
        "54945401920dcac97a8141b4fd884d364c78a65badd20c6e5cf142146b861640",  # linux_amd64
        "a20e952208b421d12a4ebd5ef313cbe4ee740de51a975e739a758b30bc39cb9f",  # linux_arm64
        "29b7712e7de5575d6cd7cf0f504ee39b8542e564c90e0ad9b6c84cac22061ef6",  # windows_amd64
    ],
    binary = True,
    visibility = ["PUBLIC"],
)
