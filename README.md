[![Go Reference](https://pkg.go.dev/badge/github.com/squat/schemasonnet.svg)](https://pkg.go.dev/github.com/squat/schemasonnet)
[![Go Report Card](https://goreportcard.com/badge/github.com/squat/schemasonnet)](https://goreportcard.com/report/github.com/squat/schemasonnet)
[![Build Status](https://github.com/squat/schemasonnet/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/squat/schemasonnet/actions/workflows/ci.yaml)
[![Built with Nix](https://img.shields.io/static/v1?logo=nixos&logoColor=white&label=&message=Built%20with%20Nix&color=41439a)](https://builtwithnix.org)

# Schemasonnet

Schemasonnet is a package and CLI that allows generating [JSON Schema](https://json-schema.org/) files from Jsonnet packages containing with [Docsonnet](https://github.com/jsonnet-libs/docsonnet) type annotations.
These generated JSON Schema are useful for statically validating and linting JSON/YAML inputs given to Jsonnet packages.

## Usage

[embedmd]:# (help.txt)
```txt
Convert Docsonnet type annotations to JSON Schema

Usage:
  schemasonnet [file] [flags]

Flags:
      --ext-code stringArray        Set code value of extVar (Format: key=<code>)
      --ext-code-file stringArray   Set code value of extVar from file (Format: key=filename)
  -V, --ext-str stringArray         Set string value of extVar (Format: key=value)
      --ext-str-file stringArray    Set string value of extVar from file (Format: key=filename)
  -h, --help                        help for schemasonnet
  -J, --jpath strings               Specify an additional library search dir (right-most wins) (default [vendor])
      --tla-code stringArray        Set code value of top level function (Format: key=<code>)
      --tla-code-file stringArray   Set code value of top level function from file (Format: key=filename)
  -A, --tla-str stringArray         Set string value of top level function (Format: key=value)
      --tla-str-file stringArray    Set string value of top level function from file (Format: key=filename)
  -v, --version                     version for schemasonnet
```
