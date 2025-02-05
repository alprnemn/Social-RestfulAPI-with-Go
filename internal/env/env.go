package env

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr        string
	DbConfig    dbConfig
	Mail        MailConfig
	Env         string
	FrontendURL string
	JWTConfig   JWTConfig
}

type dbConfig struct {
	Addr         string
	Port         string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
}

type MailConfig struct {
	SendGrid SendGridConfig
	Exp      time.Duration
}

type SendGridConfig struct {
	ApiKey    string
	FromEmail string
}

type JWTConfig struct {
	Secret string
	Exp    time.Duration
	Issuer string
}

var Envs = initConfig()

func initConfig() Config {

	godotenv.Load()

	return Config{
		Addr:        GetString("ADDR", ":8080"),
		Env:         "deployment",
		FrontendURL: "http://localhost:3000",
		DbConfig: dbConfig{
			Addr:         GetString("DB_ADDR", "postgres://user:adminpassword@localhost/social?sslmode=disable"),
			MaxOpenConns: GetInt("DB_MAXOPENCONNS", 3),
			MaxIdleConns: GetInt("DB_MAXIDLECONNS", 3),
			MaxIdleTime:  GetString("DB_MAXIDLETIME", "15min"),
		},
		Mail: MailConfig{
			SendGrid: SendGridConfig{
				ApiKey:    "as485qtng8a43q9jfq2382",
				FromEmail: "alprnemn@hotmail.com",
			},
			Exp: time.Hour * 24 * 3,
		},
		JWTConfig: JWTConfig{
			Secret: GetString("JWT_SECRET", "AASDHJGOAWEURHG"),
			Exp:    time.Hour * 24 * 3,
			Issuer: "gophersocial",
		},
	}
}

func GetString(key, fallback string) string {

	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	return value
}

func GetInt(key string, fallback int) int {

	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	valInt, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return valInt
}
