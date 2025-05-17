package main

import (
    "fmt"
    "os"
    "net/http"

    "github.com/go-chi/chi/v5"
)

func homePageHandler(w http.ResponseWriter, r *http.Request) {
    serveStaticTemplate(w, "templates/index.tmpl.html")
}

func photoAllPageHandler(w http.ResponseWriter, r *http.Request) {
    sendTemplateWithAPIData[[]Photo](
        w, 
        fmt.Sprintf("%s/photos", os.Getenv("API_BASE_URL")), 
        "templates/photos-all.tmpl.html",
        )
}

func photoHomePageHandler(w http.ResponseWriter, r *http.Request) {
    sendTemplateWithAPIData[[]CollectionResponse](
        w, 
        fmt.Sprintf("%s/collections", os.Getenv("API_BASE_URL")), 
        "templates/photos-home.tmpl.html",
        )
}

func photoCollectionPageHandler(w http.ResponseWriter, r *http.Request) {
    collectionID := chi.URLParam(r, "collectionId")
    sendTemplateWithAPIData[CollectionResponse](
        w, 
        fmt.Sprintf("%s/collections/%s", os.Getenv("API_BASE_URL"), collectionID), 
        "templates/photos-collection.tmpl.html",
        )
}

func photoPageHandler(w http.ResponseWriter, r *http.Request) {
    photoID := chi.URLParam(r, "photoId")
    sendTemplateWithAPIData[PhotoResponse](
        w, 
        fmt.Sprintf("%s/photos/%s", os.Getenv("API_BASE_URL"), photoID), 
        "templates/photos-photo.tmpl.html",
        )
}

func photoFromAllPageHandler(w http.ResponseWriter, r *http.Request) {
}

func photoFromCollectionPageHandler(w http.ResponseWriter, r *http.Request) {
}
