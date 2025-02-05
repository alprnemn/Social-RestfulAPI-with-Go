package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"social/internal/types"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func (app *application) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		authHeader := req.Header.Get("Authorization")

		if authHeader == "" {
			app.unAuthorizedError(w, req, fmt.Errorf("authorization header is missing"))
			return
		}

		parts := strings.Split(authHeader, " ") // authorization: Bearer eQ83DQ3RFasdfWDFWDadfas.f214f2ef21r.f1c21f

		if len(parts) != 2 || parts[0] != "Bearer" {
			app.unAuthorizedError(w, req, fmt.Errorf("authorization error is malformed"))
			return
		}

		token := parts[1]

		jwtToken, err := app.authenticator.ValidateToken(token)
		if err != nil {
			app.unAuthorizedError(w, req, err)
			return
		}

		claims := jwtToken.Claims.(jwt.MapClaims)

		userID, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)
		if err != nil {
			app.unAuthorizedError(w, req, err)
			return
		}

		ctx := req.Context()

		user, err := app.store.Users.GetById(ctx, userID)
		if err != nil {
			app.unAuthorizedError(w, req, err)
			return
		}

		ctx = context.WithValue(ctx, userCtx, user)

		next.ServeHTTP(w, req.WithContext(ctx))

	})
}

func (app *application) BasciAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// read the auth header
			authHeader := req.Header.Get("Authorization")
			if authHeader == "" {
				app.unAuthorizedError(w, req, fmt.Errorf("authorization header is missing"))
				return
			}
			// parse it -> get the base64
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Basic" {
				app.unAuthorizedError(w, req, fmt.Errorf("authorization header is malformed"))
				return
			}

			// decode it
			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				app.unAuthorizedError(w, req, err)
				return
			}

			username := "asdgfasdgasd"
			pass := "asdgfasdg"

			creds := strings.SplitN(string(decoded), ":", 2)
			if len(creds) != 2 || creds[0] != username || creds[1] != pass {
				app.unAuthorizedError(w, req, fmt.Errorf("invalid credentials"))
				return
			}

			// check the credentials

			next.ServeHTTP(w, req)
		})
	}
}

func (app *application) checkPostOwnership(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		user := getUserFromCtx(req)
		post := getPostFromCtx(req)

		// check if its users post
		if post.UserId == user.ID {
			next.ServeHTTP(w, req)
			return
		}

		allowed, err := app.checkRolePrecedence(req.Context(), user, requiredRole)
		if err != nil {
			app.internalServerError(w, req, err)
			return
		}

		if !allowed {
			app.forbiddenResponse(w, req)
		}

		next.ServeHTTP(w, req)
	})
}

func (app *application) checkRolePrecedence(ctx context.Context, user *types.User, roleName string) (bool, error) {
	role, err := app.store.Roles.GetByName(ctx, roleName)
	if err != nil {
		return false, err
	}

	return user.Role.Level >= role.Level, nil

}
