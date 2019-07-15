package jaeger

import (
	"time"

	"github.com/bilibili/kratos/pkg/net/trace"

	jaeger "github.com/uber/jaeger-client-go"
)

type report struct {
	rpt jaeger.Reporter
}

func newReport(c *Config) *report {
	tr, err := jaeger.NewUDPTransport(c.UDPAddr, c.MaxPacketSize)
	if err != nil {
		panic(err)
	}
	return &report{
		rpt: jaeger.NewRemoteReporter(tr,
			jaeger.ReporterOptions.BufferFlushInterval(time.Duration(c.BufferFlushInterval)),
			jaeger.ReporterOptions.QueueSize(c.QueueSize),
		),
	}
}

// WriteSpan write a trace span to queue.
func (r *report) WriteSpan(raw *trace.Span) (err error) {
	span := new(jaeger.Span)
	span.SetOperationName(raw.Name())
	r.rpt.Report(span)
	return
}

// Close close the report.
func (r *report) Close() (err error) {
	r.rpt.Close()
	return
}
