package main

import (
	"net/http"
	"social/internal/store"
)

func (app *application) getUserFeedHandler(w http.ResponseWriter, req *http.Request) {

	// pagination, filters, sort
	fq := store.PaginatedFeedQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
	}

	fq, err := fq.Parse(req)
	if err != nil {
		app.badRequestResponse(w, req, err)
		return
	}

	if err := Validate.Struct(fq); err != nil {
		app.badRequestResponse(w, req, err)
		return
	}

	ctx := req.Context()

	feed, err := app.store.Posts.GetUserFeed(ctx, int64(1), fq)
	if err != nil {
		app.internalServerError(w, req, err)
		return
	}

	if err := app.JsonResponse(w, http.StatusOK, feed); err != nil {
		app.internalServerError(w, req, err)
		return
	}

}
