// SPDX-FileCopyrightText: 2020 - 2025 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package ase

import (
	"flag"
	"fmt"

	"github.com/SAP/go-dblib/dsn"
)

type Info struct {
	dsn.Info

	AppName string `json:"appname" doc:"Application Name to transmit to ASE"`

	Userstorekey string `json:"key" multiref:"userstorekey" doc:"Key of userstore data to use for login"`

	TLSHostname string `json:"tls-hostname" doc:"Expected server TLS hostname to pass to C driver"`

	LogClientMsgs bool `json:"log-client-msgs" doc:"Log client messages"`
	LogServerMsgs bool `json:"log-server-msgs" doc:"Log server messages"`
}

// NewInfo returns a bare Info for github.com/SAP/go-dblib/dsn with defaults.
func NewInfo() (*Info, error) {
	info := new(Info)

	return info, nil
}

// NewInfoWithEnv is a convenience function returning an Info with
// values filled from the environment with the prefix 'ASE'.
func NewInfoWithEnv() (*Info, error) {
	info, err := NewInfo()
	if err != nil {
		return nil, err
	}

	if err := dsn.FromEnv("ASE", info); err != nil {
		return nil, fmt.Errorf("ase: error setting environment values on info: %w", err)
	}

	return info, nil
}

// NewInfoFlags is a convenience function returning an Info filled with
// defaults and a flagset with flags bound to the members of the
// returned info.
func NewInfoWithFlags() (*Info, *flag.FlagSet, error) {
	info, err := NewInfo()
	if err != nil {
		return nil, nil, err
	}

	flagset, err := dsn.FlagSet("", flag.ContinueOnError, info)
	if err != nil {
		return nil, nil, fmt.Errorf("ase: error creating flagset: %w", err)
	}

	return info, flagset, nil
}
