package traces_to_logs

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

type traceToLogConnector struct {
	config Config

	component.StartFunc
	component.ShutdownFunc
	logsConsumer consumer.Logs

	logger *zap.Logger
}

func NewTraceToLogsConnector(logger *zap.Logger, cfg component.Config, logsConsumer consumer.Logs) (*traceToLogConnector, error) {
	logger.Info("Creating new trace to log connector")
	conf := cfg.(*Config)
	return &traceToLogConnector{
		logger:       logger,
		config:       *conf,
		logsConsumer: logsConsumer,
	}, nil
}

func (t *traceToLogConnector) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

func (c *traceToLogConnector) ConsumeTraces(ctx context.Context, traces ptrace.Traces) error {
	newLogs := plog.NewLogs()
	for i := 0; i < traces.ResourceSpans().Len(); i++ {
		rs := traces.ResourceSpans().At(i)
		newResourceLog := newLogs.ResourceLogs().AppendEmpty()
		rs.Resource().CopyTo(newResourceLog.Resource())
		ils := rs.ScopeSpans()
		for j := 0; j < ils.Len(); j++ {
			newScopeLog := newResourceLog.ScopeLogs().AppendEmpty()
			ils.At(j).Scope().CopyTo(newScopeLog.Scope())
			spans := ils.At(j).Spans()
			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)
				logRecord := newScopeLog.LogRecords().AppendEmpty()
				logRecord.SetTraceID(span.TraceID())
				logRecord.SetSpanID(span.SpanID())
				logRecord.SetTimestamp(span.StartTimestamp())
				span.Attributes().CopyTo(logRecord.Attributes())
				// logRecord.Body().SetStr("Converted from trace")
			}
		}
	}

	c.logger.Info("Converted traces to logs")
	return c.ExportLogs(ctx, newLogs)
}

func (c *traceToLogConnector) ExportLogs(ctx context.Context, ld plog.Logs) error {
	if err := c.logsConsumer.ConsumeLogs(ctx, ld); err != nil {
		c.logger.Error("Failed to export logs", zap.Error(err))
		return err
	}
	return nil
}
