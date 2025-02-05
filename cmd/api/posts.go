package main

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"net/http"
	"social/internal/store"
	"social/internal/types"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type postKey string

const postCtx postKey = "post"

func (app *application) createPostHandler(w http.ResponseWriter, req *http.Request) {

	var payload types.CreatePostPayload

	if err := ParseJSON(w, req, &payload); err != nil {
		app.badRequestResponse(w, req, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, req, err)
		return
	}

	post := &types.Post{
		Title:   payload.Title,
		Content: payload.Content,
		UserId:  1,
		Tags:    payload.Tags,
	}

	ctx := req.Context()

	if err := app.store.Posts.Create(ctx, post); err != nil {
		app.internalServerError(w, req, err)
		return
	}

	if err := app.JsonResponse(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, req, err)
		return
	}
}

func (app *application) deletePostHandler(w http.ResponseWriter, req *http.Request) {

	post := getPostFromCtx(req)

	if err := app.store.Posts.Delete(req.Context(), post.ID); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, req, err)
		default:
			app.internalServerError(w, req, err)
		}
		return
	}

	if err := app.JsonResponse(w, http.StatusOK, map[string]string{"message": "post deleted successfully"}); err != nil {
		app.internalServerError(w, req, err)
		return
	}
}

func (app *application) getPostHandler(w http.ResponseWriter, req *http.Request) {

	post := getPostFromCtx(req)

	comments, err := app.store.Comments.GetByPostId(req.Context(), post.ID)
	if err != nil {
		app.internalServerError(w, req, err)
		return
	}

	post.Comments = comments

	if err := app.JsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, req, err)
		return
	}
}

func (app *application) updatePostHandler(w http.ResponseWriter, req *http.Request) {

	post := getPostFromCtx(req)

	var payload types.UpdatePostPayload

	if err := ParseJSON(w, req, &payload); err != nil {
		app.badRequestResponse(w, req, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, req, err)
		return
	}

	if payload.Content != nil {
		post.Content = *payload.Content
	}
	if payload.Title != nil {
		post.Title = *payload.Title
	}

	if err := app.store.Posts.Update(req.Context(), post); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.notFoundError(w, req, err)
			return
		default:
			app.internalServerError(w, req, err)
		}
		return
	}

	if err := app.JsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, req, err)
		return
	}
}

func (app *application) postsContextMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		postIdParam := chi.URLParam(req, "postId")
		postId, err := strconv.ParseInt(postIdParam, 10, 64)
		if err != nil {
			app.internalServerError(w, req, err)
			return
		}

		ctx := req.Context()

		post, err := app.store.Posts.GetPostById(ctx, postId)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundError(w, req, err)
				return
			default:
				app.internalServerError(w, req, err)
			}
			return
		}

		ctx = context.WithValue(ctx, postCtx, post)
		next.ServeHTTP(w, req.WithContext(ctx))

	})
}

func getPostFromCtx(req *http.Request) *types.Post {
	post, ok := req.Context().Value(postCtx).(*types.Post)
	if !ok {
		log.Println("error: failed to retrieve post from context")
		return nil
	}
	return post
}
