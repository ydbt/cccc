package config

import "context"

var Default *ConfigServer

func init() {
	Default = NewConfigServer(context.Background())
}
