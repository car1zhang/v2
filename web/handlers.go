package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

func homePageHandler(w http.ResponseWriter, r *http.Request) {
	serveStaticTemplate(w, "templates/index.tmpl.html")
}

func photoHomePageHandler(w http.ResponseWriter, r *http.Request) {
	sendTemplateWithAPIData[[]CollectionResponse](
		w,
		r.URL.Path,
		fmt.Sprintf("%s/collections", os.Getenv("API_BASE_URL")),
		"templates/photos-home.tmpl.html",
	)
}

func photoCollectionPageHandler(w http.ResponseWriter, r *http.Request) {
	collectionID := chi.URLParam(r, "collectionId")
	sendTemplateWithAPIData[CollectionResponse](
		w,
		r.URL.Path,
		fmt.Sprintf("%s/collections/%s", os.Getenv("API_BASE_URL"), collectionID),
		"templates/photos-collection.tmpl.html",
	)
}

func photoPhotoPageHandler(w http.ResponseWriter, r *http.Request) {
	photoID := chi.URLParam(r, "photoId")
	sendTemplateWithAPIData[PhotoResponse](
		w,
		r.URL.Path,
		fmt.Sprintf("%s/photos/%s", os.Getenv("API_BASE_URL"), photoID),
		"templates/photos-photo.tmpl.html",
	)
}
