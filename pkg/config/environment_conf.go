package config

var EnvironmentConf = map[string]string{
	"CLOUD_API_SERVER_PORT":             ":5000",
	"ENVIRONMENT":                       "development",
	"LOG_OUTPUT":                        "stdout",
	"LOG_LEVEL":                         "info",
	"SERVER_READ_TIMEOUT":               "60",
	"ACCESS_SECRET":                     "secret",
	"JWT_SECRET_KEY_EXPIRE_HOURS_COUNT": "24",
	"JWT_SECRET_KEY_EXPIRE_HOURS_COUNT_REMEMBER_ME": "168",
	"DB_SERVER_URL":                   "host=localhost port=5432 user=postgres password=somePassword dbname=cloud-api sslmode=disable",
	"DB_MAX_CONNECTIONS":              "100",
	"DB_MAX_IDLE_CONNECTIONS":         "100",
	"DB_MAX_LIFETIME_CONNECTIONS":     "100",
	"VERIFICATION_TOKEN_LENGTH":       "80", //if changed =>change at => verification_dto length role
	"VERIFICATION_TOKEN_EXPIRY_HOURS": "24",
	"SEND_GRID_API_KEY":               "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",
	"EMAIL_VERIFICATION_BASE_URL":     "http://localhost:3000/",
	"SEND_GRID_SENDER_NAME":           "Mohamed",
	"SEND_GRID_SENDER_EMAIL":          "m.abdelrhman@kotal.co",
}
