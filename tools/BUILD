_PLEASE_PEX_VERSION = "3.0.2"

# Upstream publishes no windows_amd64 please_pex, and without one a native Windows plz cannot
# build a Python target at all - the tool it needs has nothing to download. This fork publishes
# that one asset and nothing else, so only that architecture is redirected here. When upstream
# publishes its own, this goes away.
_OWNERS = {"windows_amd64": "PeterNeiss"}
_PLATFORM = f"{CONFIG.OS}_{CONFIG.ARCH}"
_OWNER = _OWNERS.get(_PLATFORM, "please-build")

remote_file(
    name = "please_pex",
    url = f"https://github.com/{_OWNER}/python-rules/releases/download/please_pex-v{_PLEASE_PEX_VERSION}/please_pex-{_PLEASE_PEX_VERSION}-{_PLATFORM}",
    # The downloaded name carries the version and platform, which on Windows leaves it with no
    # extension in PATHEXT - and Go's exec will not run such a file even given its full path.
    out = "please_pex.exe" if CONFIG.OS == "windows" else None,
    hashes = [
        "b9049a96aaead27f83e713d4d9f60b2b9a30411c7479722bec0e67254016b4a9",  # darwin_amd64
        "ba5868b241dc136509d827db6a6a65b71a5b9d003f9dbf439777467d5cd3ebd5",  # darwin_arm64
        "26f8624d81036f3672a2280e6937b04ef1a54ebde041f61e5b34b6837df24817",  # freebsd_amd64
        "1ee0046ed5c2a8b3300a36510ac76dd15eb2b6e9dc494313eda698987467bcac",  # linux_amd64
        "a20e952208b421d12a4ebd5ef313cbe4ee740de51a975e739a758b30bc39cb9f",  # linux_arm64
        "1fd268aca7b022090c40f4d44af5ef825a042d459dde2911574609e41f97f0d9",  # windows_amd64
    ],
    binary = True,
    visibility = ["PUBLIC"],
)
