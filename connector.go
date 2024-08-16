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

func NewTraceToLogsConnector(logger *zap.Logger, cfg component.Config) (*traceToLogConnector, error) {
	logger.Info("Creating new trace to log connector")
	conf := cfg.(*Config)
	return &traceToLogConnector{
		logger: logger,
		config: *conf,
	}, nil
}

func (t *traceToLogConnector) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

func (c *traceToLogConnector) ConsumeTraces(ctx context.Context, traces ptrace.Traces) error {
	allowedTraceIDs := make(map[string]bool)
	newLogs := plog.NewLogs()
	for i := 0; i < traces.ResourceSpans().Len(); i++ {
		rs := traces.ResourceSpans().At(i)
		ils := rs.ScopeSpans()
		for j := 0; j < ils.Len(); j++ {
			newScopeLog := c.newScopeLogs(newLogs)
			ils.At(j).Scope().CopyTo(newScopeLog.Scope())
			spans := ils.At(j).Spans()
			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)
				traceID := span.TraceID().String()
				allowedTraceIDs[traceID] = true
			}
		}
	}
	c.logger.Info("Allowing logs with trace IDs", zap.Any("traceIDs", allowedTraceIDs))
	return c.ExportLogs(ctx, newLogs)
}

func (c *traceToLogConnector) newScopeLogs(newLogsMap plog.Logs) plog.ScopeLogs {
	newScopeLog := newLogsMap.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty()
	return newScopeLog
}

func (c *traceToLogConnector) ExportLogs(ctx context.Context, ld plog.Logs) error {
	if err := c.logsConsumer.ConsumeLogs(ctx, ld); err != nil {
		c.logger.Error("Failed to export logs", zap.Error(err))
		return err
	}
	return nil
}
