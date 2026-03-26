package eslgo

import "time"

type JobUUID string

type Option func(*Config)

type Config struct {
	connectPassword string
	connectTimeOut  time.Duration
	commandTimeOut  time.Duration
	eventChanCap    uint
}

func DefaultConfig() *Config {
	return &Config{
		connectPassword: OPT_CONNECT_PASSWORD,
		connectTimeOut:  OPT_CONNECT_TIMEOUT * time.Second,
		commandTimeOut:  OPT_COMMAND_TIMEOUT * time.Second,
		eventChanCap:    OPT_EVENT_CHANNEL_CAPACITY,
	}
}

func WithConnectPassword(password string) Option {
	return func(cfg *Config) {
		cfg.connectPassword = password
	}
}

func WithConnectTimeout(second uint) Option {
	return func(cfg *Config) {
		if second == 0 {
			cfg.connectTimeOut = time.Duration(OPT_CONNECT_TIMEOUT) * time.Second
			return
		}
		cfg.connectTimeOut = time.Duration(second) * time.Second
	}
}

func WithCommandTimeout(second uint) Option {
	return func(cfg *Config) {
		if second == 0 {
			cfg.commandTimeOut = time.Duration(OPT_COMMAND_TIMEOUT) * time.Second
			return
		}
		cfg.commandTimeOut = time.Duration(second) * time.Second
	}
}

func WithEventChanCap(num uint) Option {
	return func(cfg *Config) {
		if num == 0 {
			cfg.eventChanCap = OPT_EVENT_CHANNEL_CAPACITY
			return
		}
		cfg.eventChanCap = num
	}
}
