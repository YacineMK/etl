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
}
