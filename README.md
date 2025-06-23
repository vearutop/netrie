# netrie

[![Build Status](https://github.com/vearutop/netrie/workflows/test-unit/badge.svg)](https://github.com/vearutop/netrie/actions?query=branch%3Amaster+workflow%3Atest-unit)
[![Coverage Status](https://codecov.io/gh/vearutop/netrie/branch/master/graph/badge.svg)](https://codecov.io/gh/vearutop/netrie)
[![GoDevDoc](https://img.shields.io/badge/dev-doc-00ADD8?logo=go)](https://pkg.go.dev/github.com/vearutop/netrie)
[![Time Tracker](https://wakatime.com/badge/github/vearutop/netrie.svg)](https://wakatime.com/badge/github/vearutop/netrie)
![Code lines](https://sloc.xyz/github/vearutop/netrie/?category=code)
![Comments](https://sloc.xyz/github/vearutop/netrie/?category=comments)

<!--- TODO Update README.md -->

Project template with GitHub actions for Go.

## Install

```
go install github.com/vearutop/netrie@latest
$(go env GOPATH)/bin/netrie --help
```

Or download binary from [releases](https://github.com/vearutop/netrie/releases).

### Linux AMD64

```
wget https://github.com/vearutop/netrie/releases/latest/download/linux_amd64.tar.gz && tar xf linux_amd64.tar.gz && rm linux_amd64.tar.gz
./netrie -version
```

### Macos Intel

```
wget https://github.com/vearutop/netrie/releases/latest/download/darwin_amd64.tar.gz && tar xf darwin_amd64.tar.gz && rm darwin_amd64.tar.gz
codesign -s - ./netrie
./netrie -version
```

### Macos Apple Silicon (M1, etc...)

```
wget https://github.com/vearutop/netrie/releases/latest/download/darwin_arm64.tar.gz && tar xf darwin_arm64.tar.gz && rm darwin_arm64.tar.gz
codesign -s - ./netrie
./netrie -version
```


## Usage

Create a new repository from this template, check out it and run `./run_me.sh` to replace template name with name of
your repository.
