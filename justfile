

# (This recipie, list just targets)
default:
    just --list

# Bootstrap the deve environment
bootstrap: getdeps 
    echo "Install air"
    go install github.com/air-verse/air@latest

# Launch the proxy inside air
[unix]
[linux]
launch: (xlaunch "tmp/proxy")
[windows]
launch: (xlaunch "tmp/proxy.exe")

[private]
xlaunch binname:
    air --build.bin {{binname}} --build.cmd "just tbuild"

[unix]
[linux]
tbuild: getdeps (xbuild "tmp/proxy")
[windows]
tbuild: getdeps (xbuild "tmp/proxy.exe")


# Build the target executable (output)
[unix]
[linux]
build: (xbuild "bin/proxy")
[windows]
build: (xbuild "bin/proxy.exe")

[private]
xbuild binname:
    echo "Building {{binname}}"
    @go build -o {{binname}} proxy.go

getdeps:
    @echo "Pulling dependencies"
    go mod download
    go mod tidy
