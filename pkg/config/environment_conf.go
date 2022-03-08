package config

import "os"

var EnvironmentConf = map[string]string{
	"CLOUD_API_SERVER_PORT":             os.Getenv("CLOUD_API_SERVER_PORT"),
	"ENVIRONMENT":                       os.Getenv("ENVIRONMENT"),
	"LOG_OUTPUT":                        os.Getenv("LOG_OUTPUT"),
	"LOG_LEVEL":                         os.Getenv("LOG_LEVEL"),
	"SERVER_READ_TIMEOUT":               os.Getenv("SERVER_READ_TIMEOUT"),
	"ACCESS_SECRET":                     os.Getenv("ACCESS_SECRET"),
	"JWT_SECRET_KEY_EXPIRE_HOURS_COUNT": os.Getenv("JWT_SECRET_KEY_EXPIRE_HOURS_COUNT"),
	"JWT_SECRET_KEY_EXPIRE_HOURS_COUNT_REMEMBER_ME": os.Getenv("JWT_SECRET_KEY_EXPIRE_HOURS_COUNT_REMEMBER_ME"),
	"DB_SERVER_URL":                   os.Getenv("DB_SERVER_URL"),
	"DB_MAX_CONNECTIONS":              os.Getenv("DB_MAX_CONNECTIONS"),
	"DB_MAX_IDLE_CONNECTIONS":         os.Getenv("DB_MAX_IDLE_CONNECTIONS"),
	"DB_MAX_LIFETIME_CONNECTIONS":     os.Getenv("DB_MAX_LIFETIME_CONNECTIONS"),
	"VERIFICATION_TOKEN_LENGTH":       os.Getenv("VERIFICATION_TOKEN_LENGTH"), //80 => if changed =>change at => verification_dto length role
	"VERIFICATION_TOKEN_EXPIRY_HOURS": os.Getenv("VERIFICATION_TOKEN_EXPIRY_HOURS"),
	"SEND_GRID_API_KEY":               os.Getenv("SEND_GRID_API_KEY"),
	"EMAIL_VERIFICATION_BASE_URL":     os.Getenv("EMAIL_VERIFICATION_BASE_URL"),
	"SEND_GRID_SENDER_NAME":           os.Getenv("SEND_GRID_SENDER_NAME"),
	"SEND_GRID_SENDER_EMAIL":          os.Getenv("SEND_GRID_SENDER_EMAIL"),
}
