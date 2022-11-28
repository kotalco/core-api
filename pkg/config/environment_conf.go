package config

import (
	"os"
)

// getenv returns environment variable by name or default value
func getenv(name, defaultValue string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return defaultValue
}

var (
	ECCPublicKey   string
	SendgridAPIKey string
	Environment    = struct {
		ServerPort                             string
		Environment                            string
		LogOutput                              string
		LogLevel                               string
		ServerReadTimeout                      string
		AccessSecret                           string
		JwtSecretKeyExpireHoursCount           string
		JwtSecretKeyExpireHoursCountRememberMe string
		DatabaseServerURL                      string
		DatabaseTestingServerURL               string
		DatabaseMaxConnections                 string
		DatabaseMaxIdleConnections             string
		DatabaseMaxLifetimeConnections         string
		DatabaseMaxIdleLifetimeConnections     string
		VerificationTokenLength                string
		VerificationTokenExpiryHours           string
		SendgridSenderName                     string
		SendgridsenderEmail                    string
		SendgridAPIKey                         string
		TwoFactorSecret                        string
		RatelimiterPerMinute                   string
		SubscriptionAPIBaseURL                 string
		ECCPublicKey                           string
		DomainMatchBaseURL                     string
	}{
		ServerPort:                             getenv("CLOUD_API_SERVER_PORT", "6000"),
		Environment:                            getenv("ENVIRONMENT", "development"),
		LogOutput:                              getenv("LOG_OUTPUT", "stdout"),
		LogLevel:                               getenv("LOG_LEVEL", "info"),
		ServerReadTimeout:                      getenv("SERVER_READ_TIMEOUT", "60"),
		AccessSecret:                           getenv("ACCESS_SECRET", "secret"), // TODO: change access secret default value
		JwtSecretKeyExpireHoursCount:           getenv("JWT_SECRET_KEY_EXPIRE_HOURS_COUNT", "24"),
		JwtSecretKeyExpireHoursCountRememberMe: getenv("JWT_SECRET_KEY_EXPIRE_HOURS_COUNT_REMEMBER_ME", "168"),
		DatabaseServerURL:                      os.Getenv("DB_SERVER_URL"),
		DatabaseTestingServerURL:               getenv("DB_TESTING_SERVER_URL", "postgres://postgres:somePassword@localhost:5432/testing-cloud-api?sslmode=disable"),
		DatabaseMaxConnections:                 getenv("DB_MAX_CONNECTIONS", "100"),
		DatabaseMaxIdleConnections:             getenv("DB_MAX_IDLE_CONNECTIONS", "100"),
		DatabaseMaxIdleLifetimeConnections:     getenv("DB_MAX_IDLE_CONNECTIONS", "10"),
		DatabaseMaxLifetimeConnections:         getenv("DB_MAX_LIFETIME_CONNECTIONS", "10"),
		VerificationTokenLength:                getenv("VERIFICATION_TOKEN_LENGTH", "80"),
		VerificationTokenExpiryHours:           getenv("VERIFICATION_TOKEN_EXPIRY_HOURS", "24"),
		SendgridSenderName:                     getenv("SEND_GRID_SENDER_NAME", "Kotal Notifications"),
		SendgridsenderEmail:                    getenv("SEND_GRID_SENDER_EMAIL", "notifications@kotal.co"),
		SendgridAPIKey:                         SendgridAPIKey,
		TwoFactorSecret:                        getenv("2_FACTOR_SECRET", "secret"), // TODO: change 2fa secret default value
		RatelimiterPerMinute:                   getenv("RATE_LIMITER_PER_MINUTE", "100"),
		SubscriptionAPIBaseURL:                 getenv("SUBSCRIPTION_API_BASE_URL", "http://localhost:8081"),
		ECCPublicKey:                           ECCPublicKey,
		DomainMatchBaseURL:                     os.Getenv("DOMAIN_MATCH_BASE_URL"),
	}
)

func init() {

	// If runtime environment is developlent
	// Set public key from the environment variable
	if Environment.Environment == "development" {
		Environment.ECCPublicKey = os.Getenv("ECC_PUBLIC_KEY")
		Environment.SendgridAPIKey = os.Getenv("SEND_GRID_API_KEY")
	}

}
