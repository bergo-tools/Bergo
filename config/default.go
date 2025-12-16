package config

import (
	_ "embed"
)

//go:embed default.toml
var defaultToml string
