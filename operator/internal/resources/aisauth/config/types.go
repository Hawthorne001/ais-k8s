/*
 * Copyright (c) 2026, NVIDIA CORPORATION. All rights reserved.
 */

package config

import "github.com/NVIDIA/aistore/cmn/cos"

// Config is the AuthN server configuration written to the mounted config file, restricted to
// settings AIStoreAuth exposes. Its fields are optional so AuthN applies its own defaults for
// whatever is absent. The AuthN library struct is not used directly because its fields always
// serialize, which writes operator-resolved defaults and limits the AuthN versions able to read
// the result.
type Config struct {
	Server  *ServerConf  `json:"auth,omitempty"`
	Log     *LogConf     `json:"log,omitempty"`
	Net     *NetConf     `json:"net,omitempty"`
	Timeout *TimeoutConf `json:"timeout,omitempty"`
}

// ServerConf configures token issuance, signing, and user storage.
type ServerConf struct {
	Expire      *cos.Duration   `json:"expiration_time,omitempty"`
	MaxTokenAge *cos.Duration   `json:"max_token_age,omitempty"`
	SigningKey  *SigningKeyConf `json:"signing_key,omitempty"`
	DB          *DatabaseConf   `json:"db,omitempty"`
}

// SigningKeyConf configures JWT signing key parameters.
type SigningKeyConf struct {
	Bits *int    `json:"bits,omitempty"`
	Mode *string `json:"mode,omitempty"`
}

// DatabaseConf configures persistent user and role storage.
type DatabaseConf struct {
	Type     *string `json:"type,omitempty"`
	Filepath *string `json:"filepath,omitempty"`
}

// LogConf configures AuthN process logging.
type LogConf struct {
	Level         *string       `json:"level,omitempty"`
	FlushInterval *cos.Duration `json:"flush_interval,omitempty"`
}

// NetConf configures the advertised URL and the HTTP listener.
type NetConf struct {
	ExternalURL *string   `json:"external_url,omitempty"`
	HTTP        *HTTPConf `json:"http,omitempty"`
}

// HTTPConf configures the AuthN HTTP(S) listener.
type HTTPConf struct {
	Certificate *string `json:"server_crt,omitempty"`
	Key         *string `json:"server_key,omitempty"`
	Port        *int    `json:"port,omitempty"`
	UseHTTPS    *bool   `json:"use_https,omitempty"`
}

// TimeoutConf configures AuthN HTTP handler timeouts.
type TimeoutConf struct {
	Default *cos.Duration `json:"default_timeout,omitempty"`
}
