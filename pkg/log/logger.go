package log

import (
	"github.com/shyimo/kubeobserver/pkg/config"

	"github.com/rs/zerolog"
)

func Initialize() {
	zerolog.SetGlobalLevel(config.LogLevel())
}
