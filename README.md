[![Build Status](https://travis-ci.com/go-home-io/server.svg?branch=master)](https://travis-ci.com/go-home-io/server) 
[![Coverage Status](https://img.shields.io/coveralls/github/go-home-io/server/master.svg)](https://coveralls.io/github/go-home-io/server?branch=master) 
[![Go Report Card](https://goreportcard.com/badge/github.com/go-home-io/server)](https://goreportcard.com/report/github.com/go-home-io/server)
[![BCH compliance](https://bettercodehub.com/edge/badge/go-home-io/server?branch=master)](https://bettercodehub.com/) 

[go-home](https://go-home.io) server


#### Preparing commit

For running on MaOS, `gmake` has to be installed since regular make has version `3.8` which does not support `ONESHELL`. You can install it using `brew`:

```bash
 brew install homebrew/core/make
 ```

Since [gometalinter](https://github.com/alecthomas/gometalinter) has certain limitation when it comes to modules support, `lint-local` target exists for local validation.

To run all required validations simply run

```bash
gmake git
```

Which includes: 
* `dep-ensure` -- running `go mod tidy`
* `generate` -- auto-generating all required files
* `lint-local` -- running all configured linters
* `test-local` -- running all available tests