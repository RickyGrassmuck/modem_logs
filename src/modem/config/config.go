package config

import (
	"net/http"
)

type Config struct {
	CustomOpts any
	ModemModel string
	Client     *http.Client
}
