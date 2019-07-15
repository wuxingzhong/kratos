package jaeger

import (
	"time"

	"github.com/bilibili/kratos/pkg/conf/env"
	"github.com/bilibili/kratos/pkg/net/trace"
	xtime "github.com/bilibili/kratos/pkg/time"
)

// Config config.
type Config struct {
	UDPAddr             string         `dsn:"udp_addr"`
	MaxPacketSize       int            `dsn:"max_packet_size"`
	BufferFlushInterval xtime.Duration `dsn:"buffer_flush_interval"`
	QueueSize           int            `dsn:"query.queue_size"`
	DisableSample       bool           `dsn:"query.disable_sample"`
}

// Init init trace report.
func Init(c *Config) {
	if c.MaxPacketSize == 0 {
		c.MaxPacketSize = 4096
	}
	if c.BufferFlushInterval == 0 {
		c.BufferFlushInterval = xtime.Duration(time.Second)
	}
	if c.QueueSize == 0 {
		c.QueueSize = 1024
	}
	trace.SetGlobalTracer(trace.NewTracer(env.AppID, newReport(c), c.DisableSample))
}
