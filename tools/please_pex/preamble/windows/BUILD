subinclude("///go//build_defs:go")

# The Windows counterpart of the C preamble next door. It is only ever built for Windows -
# please_pex embeds it alongside the native one and prepends whichever the target platform
# calls for.
go_binary(
    name = "preamble",
    srcs = ["main.go"],
    visibility = ["//tools/please_pex/pex:all"],
    deps = ["//tools/please_pex/preamble:preamble_go"],
)
