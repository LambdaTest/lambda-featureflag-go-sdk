package rootOrg

import (
	"time"
)

type Config struct {
	ServerUrl                      string
	FlagConfigPollerInterval       time.Duration
	FlagConfigPollerRequestTimeout time.Duration
}

var DefaultConfig = &Config{
	ServerUrl:                      "https://api.lambdatest.com",
	FlagConfigPollerInterval:       120 * time.Second,
	FlagConfigPollerRequestTimeout: 10 * time.Second,
}
