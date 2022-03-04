package config

var EnvironmentConf = map[string]string{
	"CLOUD_API_SERVER_PORT": ":5000",
	"ENVIRONMENT":           "development",
	"SERVER_READ_TIMEOUT":   "60",
	"LOG_OUTPUT":            "stdout",
	"LOG_LEVEL":             "info",
}
