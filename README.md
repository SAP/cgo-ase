<!--
SPDX-FileCopyrightText: 2020 - 2025 SAP SE

SPDX-License-Identifier: Apache-2.0
-->

# Deprecation Notice

This public repository is read-only and no longer maintained.

![](https://img.shields.io/badge/STATUS-NOT%20CURRENTLY%20MAINTAINED-red.svg?longCache=true&style=flat)

# cgo-ase

[![PkgGoDev](https://pkg.go.dev/badge/github.com/SAP/cgo-ase)](https://pkg.go.dev/github.com/SAP/cgo-ase)
[![Go Report Card](https://goreportcard.com/badge/github.com/SAP/cgo-ase)](https://goreportcard.com/report/github.com/SAP/cgo-ase)
[![REUSE
status](https://api.reuse.software/badge/github.com/SAP/cgo-ase)](https://api.reuse.software/info/github.com/SAP/cgo-ase)
![Actions: CI](https://github.com/SAP/cgo-ase/workflows/CI/badge.svg)

## 🛑 Repository Sunset Notice (2025-04-14)

Dear Contributers and Users,

This is an announcement that we are planning to officially sunset this repository. This decision was made after careful consideration of the project's current trajectory, community engagement, and resource allocation.

The sunset will happen after a grace period of four weeks following this announcement.
Development and support will cease, and the project will be archived.

Thank You: We are grateful for your contributions and support.

For questions or concerns, please reach out.

## Description

`cgo-ase` is a driver for the [`database/sql`][pkg-database-sql] package
of [Go (golang)][go] to provide access to SAP ASE instances.
It is delivered as Go module.

SAP ASE is the shorthand for [SAP Adaptive Server Enterprise][sap-ase],
a relational model database server originally known as Sybase SQL
Server.

[cgo][cgo] enables Go to call C code and to link against shared objects.
A pure go implementation can be found [here][purego].

## Requirements

The `cgo` driver requires the shared objects from either the ASE itself
or Client-Library to compile.

The required shared objects from ASE can be found in the installation
path of the ASE under `OCS-16_0/lib`, where `16_0` is the version of
your ASE installation.

After [installing the Client-Library SDK][cl-sdk-install-guide] the
shared objects can be found in the folder `lib` at the chosen
installation path.

The headers are provided in the `includes` folder.

Aside from the shared object the cgo driver has no special requirements
other than Go standard library and the third party modules listed in
`go.mod`, e.g. `github.com/SAP/go-dblib`.

## Download and Installation

The packages in this repo can be `go get` and imported as usual, e.g.:

```sh
go get github.com/SAP/cgo-ase
```

For specifics on how to use `database/sql` please see the
[documentation][pkg-database-sql].

The command-line application `cgoase` can be `go install`ed:

```sh
$ go install github.com/SAP/cgo-ase/cmd/cgoase@latest
go: downloading github.com/SAP/cgo-ase v0.0.0-20210506101112-3f277f8e0603
$ cgoase -h
Usage of cgoase:
      --appname string        Application Name to transmit to ASE
      --database string       Database
  -f, --f string              Read SQL commands from file
      --host string           Hostname to connect to
      --key string            Key of userstore data to use for login
      --log-client-msgs       Log client messages
      --log-server-msgs       Log server messages
      --maxColLength int      Maximum number of characters to print for column (default 50)
      --password string       Password
      --port string           Port (Example: '443' or 'tls') to connect to
      --tls-hostname string   Expected server TLS hostname to pass to C driver
      --username string       Username
2021/05/06 11:00:22 cgoase failed: pflag: help requested
```

## Usage

Example code:

```go
package main

import (
    "database/sql"
    _ "github.com/SAP/cgo-ase"
)

func main() {
    db, err := sql.Open("ase", "ase://user:pass@host:port/")
    if err != nil {
        log.Printf("Failed to open database: %v", err)
        return
    }
    defer db.Close()

    if err := db.Ping(); err != nil {
        log.Printf("Failed to ping database: %v", err)
        return
    }
}
```

`/path/to/OCS` is the path to your Client-Library SDK installation.
`/lib` is the folder inside of the SDK installation containing the
shared objects required for the cgo driver.

### Compilation

```sh
CGO_LDFLAGS="-L/path/to/OCS/lib" go build -o cgoase ./cmd/cgoase/
```

### Execution

```sh
LD_LIBRARY_PATH="/path/to/OCS/lib:/path/to/OCS/lib3p:/path/to/OCS/lib3p64:" ./cgoase
```

While `/path/to/OCS/lib` contains the libraries of the Open Client, `/path/to/OCS/lib3p`
and `/path/to/OCS/lib3p64` contain the libraries needed to use ASE user store keys.

### Examples

More examples can be found in the folder `examples`.

### Integration tests

Integration tests are available and can be run using `go test --tags=integration` and
`go test ./examples/... --tags=integration`.

These require the following environment variables to be set:

- `ASE_USERSTOREKEY`
- `ASE_HOST`
- `ASE_PORT`
- `ASE_USER`
- `ASE_PASS`

The integration tests will create new databases for each connection type to run tests
against. After the tests are finished the created databases will be removed.

## Configuration

The configuration is handled through either a data source name (DSN) in
one of two forms or through a configuration struct passed to a connector.

All of these support additional properties which can tweak the
connection, configuration options in Client-Library or the drivers
themselves.

### Data Source Names

#### URI DSN

The URI DSN is a common URI like `ase://user:pass@host:port/?prop1=val1&prop2=val2`.

DSNs in this form are parsed using `url.Parse`.

#### Simple DSN

The simple DSN is a key/value string: `username=user password=pass host=hostname port=4901`

Values with spaces must be quoted using single or double quotes.

Each member of `dblib.dsn.DsnInfo` can be set using any of their
possible json tags. E.g. `.Host` will receive the values from the keys
`host` and `hostname`.

Additional properties are set as key/value pairs as well: `...
prop1=val1 prop2=val2`. If the parser doesn't recognize a string as
a json tag it assumes that the key/value pair is a property and its
value.

Similar to the URI DSN those property/value pairs are purely additive.
Any property that only recognizes a single argument (e.g. a boolean)
will only honour the last given value for a property.

#### Connector

As an alternative to the string DSN `ase.NewConnector` accept a `dsn.DsnInfo`
directly and return a `driver.Connector`, which can be passed to `sql.OpenDB`:

```go
package main

import (
    "database/sql"

    "github.com/SAP/cgo-ase"
)

func main() {
    info := ase.NewInfo()
    info.Host = "hostname"
    info.Port = "4901"
    info.Username = "user"
    info.Password = "pass"

    connector, err := ase.NewConnector(info)
    if err != nil {
        log.Printf("Failed to create connector: %v", err)
        return
    }

    db, err := sql.OpenDB(connector)
    if err != nil {
        log.Printf("Failed to open database: %v", err)
        return
    }
    defer db.Close()

    if err := db.Ping(); err != nil {
        log.Printf("Failed to ping ASE: %v", err)
    }
}
```

### Properties

##### AppName / app-name

Recognized values: string

When set overrides Client-Libraries default application name sent to the
ASE server.

##### Userstorekey / userstorekey

Recognized values: string

When set uses the ASE userstore instead of username/password to
authenticate with ASE.

##### TLSHostname / tls-hostname

Recognized values: string

Expected server TLS hostname to pass to Client-Library for validation.

##### LogClientMsgs / log-client-msgs

Recognized values: `true` or `false`

When set to `true` all client messages will be printed to stderr.

Please note that this is a debug property - for logging you should
register your own message handler with the `GlobalClientMessageBroker`.

When unset the callback will not bet set.

##### LogServerMsgs / log-server-msgs

Recognized values: `true` or `false`

When set to `true` all server messages will be printed to stderr.

Please note that this is a debug property - for logging you should
register your own message handler with the `GlobalServerMessageBroker`.

When unset the callback will not bet set.

## Limitations

### Prepared statements

Regarding the limitations of prepared statements/dynamic SQL please see
[the Client-Library documentation](https://help.sap.com/viewer/71b47f4a8269411da6d15ed25f5d39b3/LATEST/en-US/bfc531e46db61014bf8f040071e613d7.html).

### Unsupported ASE data types

Currently the following data types are not supported:

- Timestamp
- Univarchar

### Null types

Due to the limitations of the Client-Library it is not possible to
support null types.

Additionally columns of the following data types must be nullable:

- Image
- Binary

## Known Issues

The list of known issues is available [here][issues].

## How to obtain support

Feel free to open issues for feature requests, bugs or general feedback [here][issues].

## Contributing

Any help to improve this package is highly appreciated.

For details on how to contribute please see the
[contributing](CONTRIBUTING.md) file.

## License

Copyright (c) 2019-2020 SAP SE or an SAP affiliate company. All rights reserved.
This file is licensed under the Apache License 2.0 except as noted otherwise in the [LICENSE file](LICENSES).

[cgo]: https://golang.org/cmd/cgo
[cl-sdk-install-guide]: https://help.sap.com/viewer/882ef48c7e9c4d6e845d98f34378db40/16.0.3.2/en-US
[go]: https://golang.org/
[go-dblib]: https://www.github.com/SAP/go-dblib
[issues]: https://github.com/SAP/cgo-ase/issues
[pkg-database-sql]: https://golang.org/pkg/database/sql
[purego]: https://github.com/SAP/go-ase
[sap-ase]: https://www.sap.com/products/sybase-ase.html
