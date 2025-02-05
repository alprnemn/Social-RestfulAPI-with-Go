package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"social/internal/store"
	"social/internal/types"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type userKey string

const userCtx userKey = "user"

func (app *application) getUserHandler(w http.ResponseWriter, req *http.Request) {

	user := getUserFromCtx(req)

	if err := app.JsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, req, err)
		return
	}
}

func (app *application) followUserHandler(w http.ResponseWriter, req *http.Request) {

	user := getUserFromCtx(req)

	var payload types.FollowUserPayload
	if err := ParseJSON(w, req, &payload); err != nil {
		app.badRequestResponse(w, req, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, req, err)
		return
	}

	ctx := req.Context()

	err := app.store.Users.Follow(ctx, user.ID, payload.UserID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrConflict):
			app.conflictError(w, req, err)
			return

		default:
			app.internalServerError(w, req, err)
			return
		}

	}

	if err := app.JsonResponse(w, http.StatusNoContent, map[string]string{"message": "success"}); err != nil {
		app.internalServerError(w, req, err)
		return
	}
}

func (app *application) unfollowUserHandler(w http.ResponseWriter, req *http.Request) {

	user := getUserFromCtx(req)

	var payload types.FollowUserPayload

	if err := ParseJSON(w, req, &payload); err != nil {
		app.internalServerError(w, req, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, req, err)
		return
	}

	ctx := req.Context()

	err := app.store.Users.Unfollow(ctx, user.ID, payload.UserID)

	if err != nil {
		switch {
		case errors.Is(err, store.ErrConflict):
			app.conflictError(w, req, err)
			return

		default:
			app.internalServerError(w, req, err)
			return
		}
	}

	if err := app.JsonResponse(w, http.StatusNoContent, map[string]string{"message": "success"}); err != nil {
		app.internalServerError(w, req, err)
		return
	}
}

func (app *application) activateUserHandler(w http.ResponseWriter, req *http.Request) {

	token := chi.URLParam(req, "token")

	ctx := req.Context()

	err := app.store.Users.Activate(ctx, token)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundError(w, req, err)
		default:
			app.internalServerError(w, req, err)
		}
		return
	}

	if err := app.JsonResponse(w, http.StatusNoContent, map[string]string{"message": "user activated successfully"}); err != nil {
		app.internalServerError(w, req, err)
		return
	}

}

func (app *application) userContextMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		userIdParam := chi.URLParam(req, "userId")
		userId, err := strconv.ParseInt(userIdParam, 10, 64)
		if err != nil {
			app.internalServerError(w, req, err)
			return
		}

		ctx := req.Context()

		user, err := app.store.Users.GetById(ctx, userId)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundError(w, req, err)
				return
			default:
				app.internalServerError(w, req, err)
			}
		}

		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, req.WithContext(ctx))

	})
}

func getUserFromCtx(req *http.Request) *types.User {
	user, ok := req.Context().Value(userCtx).(*types.User)
	if !ok {
		log.Println("error: failed to get user from context")
		return nil
	}
	return user
}
