package store

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PaginatedFeedQuery struct {
	Limit  int      `json:"limit" validate:"gte=1,lte=20"`
	Offset int      `json:"offset" validate:"gte=0"`
	Sort   string   `json:"sort" validate:"oneof=asc desc"`
	Tags   []string `json:"tags" validate:"max=5"`
	Search string   `json:"search" validate:"max=100"`
	Since  string   `json:"since"`
	Until  string   `json:"until"`
}

func (PaginatedFeedQuery PaginatedFeedQuery) Parse(req *http.Request) (PaginatedFeedQuery, error) {

	// /feed?limit=10&offset=5&sort=asc
	qs := req.URL.Query()

	limit := qs.Get("limit")

	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return PaginatedFeedQuery, nil
		}
		PaginatedFeedQuery.Limit = l
	}

	offset := qs.Get("offset")
	if offset != "" {
		o, err := strconv.Atoi(offset)
		if err != nil {
			return PaginatedFeedQuery, nil
		}
		PaginatedFeedQuery.Offset = o
	}

	sort := qs.Get("sort")
	if sort != "" {
		PaginatedFeedQuery.Sort = sort
	}

	tags := qs.Get("tags")
	if tags != "" {
		PaginatedFeedQuery.Tags = strings.Split(tags, ",")
	}

	search := qs.Get("search")
	if search != "" {
		PaginatedFeedQuery.Search = search
	}

	since := qs.Get("since")
	if since != "" {
		log.Println("since:: ", since)
	}

	until := qs.Get("until")
	if until != "" {
		log.Println("until:: ", until)
	}

	return PaginatedFeedQuery, nil

}

func parseTime(s string) string {
	t, err := time.Parse(time.DateTime, s)
	if err != nil {
		return ""
	}
	return t.Format(time.DateTime)
}

// My solution
// func (PaginatedFeedQuery PaginatedFeedQuery) ParseTagsFromURLQuery(tags string) []string {
// 	tagStrings := []string{}
// 	tagString := ""
// 	for index := 0; index < len(tags); index++ {
// 		if tags[index] == 44 {
// 			tagStrings = append(tagStrings, tagString)
// 			tagString = ""
// 		} else if index != len(tags)-1 {
// 			tagString = tagString + string(tags[index])
// 		} else {
// 			tagString = tagString + string(tags[index])
// 			tagStrings = append(tagStrings, tagString)
// 		}
// 	}
// 	return tagStrings
// }
