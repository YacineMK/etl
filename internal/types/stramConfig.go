package types

type StreamConfig struct {
	Subject       string
	ConsumerGroup string
}

var StreamsList = []StreamConfig{
	{
		Subject:       "etl:jobs",
		ConsumerGroup: "etl-processors",
	},
	{
		Subject:       "etl:write-jobs",
		ConsumerGroup: "postgres-writers",
	},
	{
		Subject:       "etl:retry",
		ConsumerGroup: "retry-processors",
	},
	{
		Subject:       "etl:dlq",
		ConsumerGroup: "dlq-processors",
	},
}
