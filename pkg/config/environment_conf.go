package config

import (
	"os"
)

var (
	EnvironmentConf = map[string]string{}
	ECCPublicKey    string
)

func init() {
	var val string
	var ok bool

	//app config
	val, ok = os.LookupEnv("CLOUD_API_SERVER_PORT")
	if !ok {
		EnvironmentConf["CLOUD_API_SERVER_PORT"] = "6000"
	} else {
		EnvironmentConf["CLOUD_API_SERVER_PORT"] = val
	}

	val, ok = os.LookupEnv("ENVIRONMENT")
	if !ok {
		EnvironmentConf["ENVIRONMENT"] = "development"
	} else {
		EnvironmentConf["ENVIRONMENT"] = val
	}

	val, ok = os.LookupEnv("LOG_OUTPUT")
	if !ok {
		EnvironmentConf["LOG_OUTPUT"] = "stdout"
	} else {
		EnvironmentConf["LOG_OUTPUT"] = val
	}

	val, ok = os.LookupEnv("LOG_LEVEL")
	if !ok {
		EnvironmentConf["LOG_LEVEL"] = "info"
	} else {
		EnvironmentConf["LOG_LEVEL"] = val
	}

	val, ok = os.LookupEnv("SERVER_READ_TIMEOUT")
	if !ok {
		EnvironmentConf["SERVER_READ_TIMEOUT"] = "60"
	} else {
		EnvironmentConf["SERVER_READ_TIMEOUT"] = val
	}

	//Jwt configs
	val, ok = os.LookupEnv("ACCESS_SECRET")
	if !ok {
		EnvironmentConf["ACCESS_SECRET"] = "secret"
	} else {
		EnvironmentConf["ACCESS_SECRET"] = val
	}

	val, ok = os.LookupEnv("JWT_SECRET_KEY_EXPIRE_HOURS_COUNT")
	if !ok {
		EnvironmentConf["JWT_SECRET_KEY_EXPIRE_HOURS_COUNT"] = "24"
	} else {
		EnvironmentConf["JWT_SECRET_KEY_EXPIRE_HOURS_COUNT"] = val
	}

	val, ok = os.LookupEnv("JWT_SECRET_KEY_EXPIRE_HOURS_COUNT_REMEMBER_ME")
	if !ok {
		EnvironmentConf["JWT_SECRET_KEY_EXPIRE_HOURS_COUNT_REMEMBER_ME"] = "168"
	} else {
		EnvironmentConf["JWT_SECRET_KEY_EXPIRE_HOURS_COUNT_REMEMBER_ME"] = val
	}

	//Db configs
	EnvironmentConf["DB_SERVER_URL"] = os.Getenv("DB_SERVER_URL")

	val, ok = os.LookupEnv("DB_TESTING_SERVER_URL")
	if !ok {
		EnvironmentConf["DB_TESTING_SERVER_URL"] = "host=localhost port=5432 user=postgres password=somePassword dbname=testing-cloud-api sslmode=disable"
	} else {
		EnvironmentConf["DB_TESTING_SERVER_URL"] = val
	}

	val, ok = os.LookupEnv("DB_MAX_CONNECTIONS")
	if !ok {
		EnvironmentConf["DB_MAX_CONNECTIONS"] = "100"
	} else {
		EnvironmentConf["DB_MAX_CONNECTIONS"] = val
	}

	val, ok = os.LookupEnv("DB_MAX_IDLE_CONNECTIONS")
	if !ok {
		EnvironmentConf["DB_MAX_IDLE_CONNECTIONS"] = "100"
	} else {
		EnvironmentConf["DB_MAX_IDLE_CONNECTIONS"] = val
	}

	val, ok = os.LookupEnv("DB_MAX_LIFETIME_CONNECTIONS")
	if !ok {
		EnvironmentConf["DB_MAX_LIFETIME_CONNECTIONS"] = "100"
	} else {
		EnvironmentConf["DB_MAX_LIFETIME_CONNECTIONS"] = val
	}

	val, ok = os.LookupEnv("VERIFICATION_TOKEN_LENGTH")
	if !ok {
		EnvironmentConf["VERIFICATION_TOKEN_LENGTH"] = "80"
	} else {
		EnvironmentConf["VERIFICATION_TOKEN_LENGTH"] = val
	}

	val, ok = os.LookupEnv("VERIFICATION_TOKEN_EXPIRY_HOURS")
	if !ok {
		EnvironmentConf["VERIFICATION_TOKEN_EXPIRY_HOURS"] = "24"
	} else {
		EnvironmentConf["VERIFICATION_TOKEN_EXPIRY_HOURS"] = val
	}

	val, ok = os.LookupEnv("SEND_GRID_SENDER_NAME")
	if !ok {
		EnvironmentConf["SEND_GRID_SENDER_NAME"] = "Mohamed"
	} else {
		EnvironmentConf["SEND_GRID_SENDER_NAME"] = val
	}
	val, ok = os.LookupEnv("SEND_GRID_SENDER_EMAIL")
	if !ok {
		EnvironmentConf["SEND_GRID_SENDER_EMAIL"] = "m.abdelrhman@kotal.co"
	} else {
		EnvironmentConf["SEND_GRID_SENDER_EMAIL"] = val
	}

	EnvironmentConf["SEND_GRID_API_KEY"] = os.Getenv("SEND_GRID_API_KEY")

	//tfa configs
	val, ok = os.LookupEnv("2_FACTOR_SECRET")
	if !ok {
		EnvironmentConf["2_FACTOR_SECRET"] = "secret"
	} else {
		EnvironmentConf["2_FACTOR_SECRET"] = val
	}

	//rate limiting
	val, ok = os.LookupEnv("RATE_LIMITER_PER_MINUTE")
	if !ok {
		EnvironmentConf["RATE_LIMITER_PER_MINUTE"] = "100"
	} else {
		EnvironmentConf["RATE_LIMITER_PER_MINUTE"] = val
	}

	//LICENSE
	val, ok = os.LookupEnv("SUBSCRIPTION_API_BASE_URL")
	if !ok {
		EnvironmentConf["SUBSCRIPTION_API_BASE_URL"] = "http://localhost:8081"
	} else {
		EnvironmentConf["SUBSCRIPTION_API_BASE_URL"] = val
	}
	//ECC
	if EnvironmentConf["ENVIRONMENT"] == "development" {
		EnvironmentConf["ECC_PUBLIC_KEY"] = os.Getenv("ECC_PUBLIC_KEY")
	} else {
		EnvironmentConf["ECC_PUBLIC_KEY"] = ECCPublicKey
	}

	//TRAEFIK
	EnvironmentConf["DOMAIN_MATCH_BASE_URL"] = os.Getenv("DOMAIN_MATCH_BASE_URL")
}
