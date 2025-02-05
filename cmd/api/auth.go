package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"social/internal/env"
	"social/internal/store"
	"social/internal/types"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func (app *application) registerUserHandler(w http.ResponseWriter, req *http.Request) {

	// parse payload that come from user
	var payload types.RegisterUserPayload
	if err := ParseJSON(w, req, &payload); err != nil {
		app.badRequestResponse(w, req, err)
		return
	}

	// validate payload
	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, req, err)
		return
	}

	// hash and store password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		app.internalServerError(w, req, err)
		return
	}

	// create user object
	user := &types.User{
		Username: payload.Username,
		Email:    payload.Email,
		Password: string(hashedPassword),
	}

	ctx := req.Context()

	inv_token := generateInvToken()

	hash := sha256.Sum256([]byte(inv_token))
	hashToken := hex.EncodeToString(hash[:])

	// create and invites join into the data layer
	err1 := app.store.Users.CreateAndInvite(ctx, user, hashToken, env.Envs.Mail.Exp)
	if err1 != nil {
		switch err1 {
		case store.ErrDuplicateEmail:
			app.badRequestResponse(w, req, err1)
			return
		case store.ErrDuplicateUsername:
			app.badRequestResponse(w, req, err1)
			return
		default:
			app.internalServerError(w, req, err1)
		}
		return
	}

	userWithToken := types.UserWithToken{
		User:  user,
		Token: inv_token,
	}

	// isProdEnv := env.Envs.Env == "production"
	// activationURL := fmt.Sprintf("%s/confirm/%s", env.Envs.FrontendURL, inv_token)
	// vars := struct {
	// 	Username      string
	// 	ActivationURL string
	// }{
	// 	Username:      user.Username,
	// 	ActivationURL: activationURL,
	// }

	// send email
	// err = app.mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, !isProdEnv)

	// if err != nil {
	// 	log.Println("error sending welcome email", "error", err.Error())

	// 	if err := app.store.Users.Delete(ctx, user.ID); err != nil {
	// 		log.Println("error sending welcome email", " error: ", err.Error())
	// 	}

	// 	app.internalServerError(w, req, err)
	// 	return
	// }

	// return json response
	if err := app.JsonResponse(w, http.StatusCreated, userWithToken); err != nil {
		app.internalServerError(w, req, err)
		return
	}
}

func (app *application) createTokenHandler(w http.ResponseWriter, req *http.Request) {

	// parse login payload
	var payload types.CreateUserTokenPayload
	if err := ParseJSON(w, req, &payload); err != nil {
		app.badRequestResponse(w, req, err)
		return
	}

	// validate
	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, req, err)
		return
	}

	ctx := req.Context()

	// check user if exist
	user, err := app.store.Users.GetByEmail(ctx, payload.Email)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.unAuthorizedError(w, req, err)
			return
		default:
			app.internalServerError(w, req, err)
		}
		return
	}

	// check password is valid
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
		app.badRequestResponse(w, req, fmt.Errorf("invalid password"))
		return
	}

	// generate token
	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(env.Envs.JWTConfig.Exp).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": env.Envs.JWTConfig.Issuer,
		"aud": env.Envs.JWTConfig.Issuer,
	}

	token, err := app.authenticator.GenerateToken(claims)
	if err != nil {
		app.internalServerError(w, req, err)
		return
	}

	// return json response
	if err := app.JsonResponse(w, http.StatusCreated, token); err != nil {
		app.internalServerError(w, req, err)
		return
	}

}

func generateInvToken() string {
	inv_token := fmt.Sprintf("D26O43JY78-%s", time.Now().Format("15:04:05.000000000"))
	return inv_token
}
