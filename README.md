# Tidal

[![Build Status](https://travis-ci.com/rotationalio/tidal.svg?branch=master)](https://travis-ci.com/rotationalio/tidal)
[![codecov](https://codecov.io/gh/rotationalio/tidal/branch/master/graph/badge.svg)](https://codecov.io/gh/rotationalio/tidal)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/rotationalio/tidal)](https://pkg.go.dev/github.com/rotationalio/tidal)
[![Go Report Card](https://goreportcard.com/badge/github.com/rotationalio/tidal)](https://goreportcard.com/report/github.com/rotationalio/tidal)

**Database schema migration management and code generation.**

Tidal provides a mechanism to define and manage database schema migrations using SQL files that both specify the up (apply) and down (rollback) actions to ensure consistent changes in the database schema as application versions change. Tidal includes a CLI tool to generate these files into descriptors, directly adding them to your application source code so they can be compiled into the binary. The tidal package implements utilities for managing the state of the database with respect to the migrations, even across different binaries and application versions.
