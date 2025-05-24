package main

import (
	"html/template"
	"net/http"
	"os"
    "fmt"

    "github.com/go-chi/chi/v5"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/layout.tmpl.html", "templates/login.tmpl.html"))

	data := struct {
		APIBaseURL  string
	}{
		APIBaseURL:  os.Getenv("API_BASE_URL"),
	}

	tmpl.Execute(w, data)
}

func photosPageHandler(w http.ResponseWriter, r *http.Request) {
    sendTemplateWithAPIData[[]Photo](
        w,
        fmt.Sprintf("%s/photos", os.Getenv("API_BASE_URL")),
        "templates/photos.tmpl.html",
    )
}

func collectionsPageHandler(w http.ResponseWriter, r *http.Request) {
    sendTemplateWithAPIData[[]CollectionResponse](
        w,
        fmt.Sprintf("%s/collections", os.Getenv("API_BASE_URL")),
        "templates/collections.tmpl.html",
    )
}

func newPhotoPageHandler(w http.ResponseWriter, r *http.Request) {
    tmpl := template.Must(template.ParseFiles("templates/layout.tmpl.html", "templates/new-photo.tmpl.html"))

    data := struct {
        APIBaseURL string
    }{
        APIBaseURL: os.Getenv("API_BASE_URL"),
    }

    tmpl.Execute(w, data)
}

func newCollectionPageHandler(w http.ResponseWriter, r *http.Request) {
    tmpl := template.Must(template.ParseFiles("templates/layout.tmpl.html", "templates/new-collection.tmpl.html"))

    photos, err := fetchData[[]Photo](fmt.Sprintf("%s/photos", os.Getenv("API_BASE_URL")))
	if err != nil {
		http.Error(w, "Failed to fetch data: "+err.Error(), http.StatusInternalServerError)
		return
    }

    data := struct {
        APIBaseURL string
        Photos []Photo
    }{
        APIBaseURL: os.Getenv("API_BASE_URL"),
        Photos: photos,
    }

    tmpl.Execute(w, data)
}

func editPhotoPageHandler(w http.ResponseWriter, r *http.Request) {
    photoID := chi.URLParam(r, "photoId")

	w.Header().Set("Content-Type", "text/html")

	data, err := fetchData[PhotoResponse](fmt.Sprintf("%s/photos/%s", os.Getenv("API_BASE_URL"), photoID))
	if err != nil {
		http.Error(w, "Failed to fetch data: "+err.Error(), http.StatusInternalServerError)
		return
	}

    pageBuf, err := fillTemplateWithData("templates/edit-photo.tmpl.html", struct {
        Photo Photo
        Collections []Collection
        APIBaseURL string
    }{
            Photo: data.Photo,
            Collections: data.Collections,
            APIBaseURL: os.Getenv("API_BASE_URL"),
        })
    if err != nil {
        http.Error(w, "Failed to fill template: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Write(pageBuf.Bytes())
}

func editCollectionPageHandler(w http.ResponseWriter, r *http.Request) {

    collectionID := chi.URLParam(r, "collectionId")

	w.Header().Set("Content-Type", "text/html")

	collectionData, err := fetchData[CollectionResponse](fmt.Sprintf("%s/collections/%s", os.Getenv("API_BASE_URL"), collectionID))
	if err != nil {
		http.Error(w, "Failed to fetch data: "+err.Error(), http.StatusInternalServerError)
		return
	}
    allPhotos, err := fetchData[[]Photo](fmt.Sprintf("%s/photos", os.Getenv("API_BASE_URL")))
	if err != nil {
		http.Error(w, "Failed to fetch data: "+err.Error(), http.StatusInternalServerError)
		return
    }

    var nonPhotos []Photo
    for _, photo := range allPhotos {
        found := false
        for _, photoB := range collectionData.Photos {
            if photoB.ID == photo.ID {
                found = true
                break
            }
        }
        if !found {
            nonPhotos = append(nonPhotos, photo)
        }
    }

    pageBuf, err := fillTemplateWithData("templates/edit-collection.tmpl.html", struct {
        Collection Collection
        Photos []Photo
        NonPhotos []Photo
        APIBaseURL string
    }{
            Collection: collectionData.Collection,
            Photos: collectionData.Photos,
            NonPhotos: nonPhotos,
            APIBaseURL: os.Getenv("API_BASE_URL"),
        })
    if err != nil {
        http.Error(w, "Failed to fill template: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Write(pageBuf.Bytes())
}
