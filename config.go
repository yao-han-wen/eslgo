package eslgo

import "time"

type JobUUID string

type Option func(*Config)

type Config struct {
	eventChanCap uint
	cmdTimeOut   time.Duration
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

func WithCmdTimeout(second uint) Option {
	return func(cfg *Config) {
		if second == 0 {
			cfg.cmdTimeOut = time.Duration(OPT_CMD_TIMEOUT) * time.Second
			return
		}
		cfg.cmdTimeOut = time.Duration(second) * time.Second
	}
}
