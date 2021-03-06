[![Build Status](https://travis-ci.com/go-home-io/server.svg?branch=master)](https://travis-ci.com/go-home-io/server) [![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fgo-home-io%2Fserver.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fgo-home-io%2Fserver?ref=badge_shield)

[![Coverage Status](https://img.shields.io/coveralls/github/go-home-io/server/master.svg)](https://coveralls.io/github/go-home-io/server?branch=master) 
[![Go Report Card](https://goreportcard.com/badge/github.com/go-home-io/server)](https://goreportcard.com/report/github.com/go-home-io/server)
[![BCH compliance](https://bettercodehub.com/edge/badge/go-home-io/server?branch=master)](https://bettercodehub.com/) 

[go-home](https://go-home.io) server

#### Development environment 

`go-home` uses go 1.13 with modules support, but to provide compatibility with no-modules environments, all scripts are expecting to have `go-home.io/x` under `${GOPATH}/src` folder.

For running on MaOS, `gmake` has to be installed since regular make has version `3.8` which does not support `ONESHELL`. You can install it using `brew`:

```bash
 brew install homebrew/core/make
 ```
  
You'll need at least `nsq` running on your machine. You can install it through `brew`: 

```bash
brew install nsq
```

Then start it: 

```bash
nsqd -broadcast-address=127.0.0.1
```

Checkout both server and [providers](https://github.com/go-home-io/providers) repos into `${GOPATH}/src/go-home.io/x` folder, place your config files under `server/configs`. Minimal required configuration is: 

```yaml
system: bus
provider: nsq
server: 127.0.0.1:4150

---

system: go-home
provider: master
port: 8000
delayedStart: 0

---

system: go-home
provider: worker
name: worker-1
```

Start both worker and master by running: 

```bash
gmake run-worker
gmake run-only-server
```

#### Preparing commit

To run all required validations simply run:

```bash
gmake git
```

Which includes: 
* `dep-ensure` -- running `go mod tidy`
* `generate` -- auto-generating all required files
* `lint` -- running all configured linters
* `test-local` -- running all available tests


#### Known issues

`x/tools/cmd/goimports` [doesn't work well](https://github.com/golang/go/issues/26882) with modules, sometimes it removes correct package. To bypass this install this package without modules, e.g.: 

```bash
GO111MODULE=off go get github.com/vkorn/go-miio
```

[`monkey patching`](https://github.com/bouk/monkey) from an IDE must be executed with an extra params: 

```bash
-gcflags=-l
```

## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fgo-home-io%2Fserver.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fgo-home-io%2Fserver?ref=badge_large)