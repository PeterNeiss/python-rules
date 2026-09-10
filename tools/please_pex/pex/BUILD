subinclude("///go//build_defs:go")

go_library(
    name = "pex",
    srcs = ["pex.go"],
    resources = glob(["**/*.py"]) + [
        ":preamble",
        ":preamble_windows",
    ],
    visibility = [
        "//tools/please_pex:all",
    ],
    deps = [
        "//tools/please_pex/preamble:preamble_go",
        "//tools/please_pex/zip",
    ],
)

genrule(
    name = "preamble",
    srcs = ["//tools/please_pex/preamble"],
    outs = ["preamble"],
    cmd = "cp $SRCS $OUTS",
)

# Windows has no exec(), so its preamble is a separate program rather than a port of the C one;
# see the comment on it. It is cross-compiled here because please_pex itself always runs on the
# host, whatever platform it is building a .pex for.
#
# The label has to name this repo explicitly. An architecture subrepo is addressed as
# <repo>@<arch>, and a bare ///windows_amd64// means the *root* repo's - which is not this one
# when these rules are being used as a plugin. subrepo_name() is empty in our own repo.
_repo = subrepo_name()
_arch = "windows_amd64"
_windows = f"{_repo}@{_arch}" if _repo and not _repo.endswith(_arch) else (_repo or _arch)

genrule(
    name = "preamble_windows",
    srcs = [f"///{_windows}//tools/please_pex/preamble/windows:preamble"],
    outs = ["preamble_windows"],
    cmd = "cp $SRCS $OUTS",
)
