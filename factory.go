package traces_to_logs

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/consumer"
)

const (
	typeStr = "traces_to_logs"
)

func NewFactory() connector.Factory {
	return connector.NewFactory(
		component.MustNewType(typeStr),
		CreateDefaultConfig,
		connector.WithTracesToLogs(CreateTraceToLogConnector, component.StabilityLevelAlpha),
	)
}

func CreateTraceToLogConnector(_ context.Context, set connector.Settings, cfg component.Config, nextConsumer consumer.Logs) (connector.Traces, error) {
	t, err := NewTraceToLogsConnector(set.Logger, cfg, nextConsumer)
	if err != nil {
		return nil, err
	}
	t.logsConsumer = nextConsumer
	return t, nil
}

func CreateDefaultConfig() component.Config {
	return &Config{}
}
