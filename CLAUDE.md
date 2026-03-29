# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

DNSKit is a DNS toolkit library written in Go (1.25.5).

Module: `github.com/serkanaltuntas/dnskit`

## Build & Test Commands

```bash
go build ./...          # build all packages
go test ./...           # run all tests
go test ./pkg/foo       # run tests in a single package
go test -run TestName   # run a specific test by name
go vet ./...            # static analysis
```
