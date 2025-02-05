package main

import (
	"social/internal/auth"
	"social/internal/db"
	"social/internal/env"
	"social/internal/mailer"
	"social/internal/store"

	"log"
)

func main() {

	// database
	db, err := db.New(
		env.Envs.DbConfig.Addr,
		env.Envs.DbConfig.MaxOpenConns,
		env.Envs.DbConfig.MaxIdleConns,
		env.Envs.DbConfig.MaxIdleTime,
	)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	var (
		store     = store.NewStorage(db)
		ApiKey    = env.Envs.Mail.SendGrid.ApiKey
		FromEmail = env.Envs.Mail.SendGrid.FromEmail
		mailer    = mailer.NewSendgrid(ApiKey, FromEmail)
	)

	jwtAuthenticator := auth.NewJWTAuthenticator(
		env.Envs.JWTConfig.Secret,
		env.Envs.JWTConfig.Issuer,
		env.Envs.JWTConfig.Issuer,
	)

	app := &application{
		config:        env.Envs,
		store:         store,
		db:            db,
		mailer:        mailer,
		authenticator: jwtAuthenticator,
	}

	mux := app.mount()

	log.Fatal(app.run(mux))

}
