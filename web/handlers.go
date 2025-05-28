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

func photoPhotoPageHandler(w http.ResponseWriter, r *http.Request) {
    photoID := chi.URLParam(r, "photoId")
    sendTemplateWithAPIData[PhotoResponse](
        w,
        r.URL.Path,
        fmt.Sprintf("%s/photos/%s", os.Getenv("API_BASE_URL"), photoID),
        "templates/photos-photo.tmpl.html",
        )
}

func photoCollectionPageHandler(w http.ResponseWriter, r *http.Request) { // next collection hotkey
    collectionID := chi.URLParam(r, "collectionId")

    w.Header().Set("Content-Type", "text/html")

    if sendPageIfCached(w, r.URL.Path) {
        return
    }

    collectionResponse, err := fetchData[CollectionResponse](fmt.Sprintf("%s/collections/%s", os.Getenv("API_BASE_URL"), collectionID))
    if err != nil {
        http.Error(w, "Failed to fetch collection: "+err.Error(), http.StatusInternalServerError)
        return
    }
    allCollectionResponse, err := fetchData[[]CollectionResponse](fmt.Sprintf("%s/collections", os.Getenv("API_BASE_URL")))
    if err != nil {
        http.Error(w, "Failed to fetch collection: "+err.Error(), http.StatusInternalServerError)
        return
    }

    var prevID, nextID string
    for idx, cr := range allCollectionResponse {
        if cr.Collection.ID == collectionResponse.Collection.ID {
            if idx > 0 {
                prevID = allCollectionResponse[idx-1].Collection.ID
            }
            if idx < len(allCollectionResponse) - 1 {
                nextID = allCollectionResponse[idx+1].Collection.ID
            }
        }
    }

    data := struct {
        Collection  Collection
        Photos      []Photo
        PrevID      string
        NextID      string
    }{
        Collection: collectionResponse.Collection,
        Photos: collectionResponse.Photos,
        PrevID: prevID,
        NextID: nextID,
    }

    pageBuf, err := fillTemplateWithData("templates/photos-collection.tmpl.html", data)
    if err != nil {
        http.Error(w, "Failed to fill template: "+err.Error(), http.StatusInternalServerError)
        return
    }

    writePageToCache(pageBuf, r.URL.Path)
    w.Write(pageBuf.Bytes())
}

func photoCollectionPhotoPageHandler(w http.ResponseWriter, r *http.Request) {
    collectionID := chi.URLParam(r, "collectionId")
    photoID := chi.URLParam(r, "photoId")

    w.Header().Set("Content-Type", "text/html")

    if sendPageIfCached(w, r.URL.Path) {
        return
    }

    collectionResponse, err := fetchData[CollectionResponse](fmt.Sprintf("%s/collections/%s", os.Getenv("API_BASE_URL"), collectionID))
    if err != nil {
        http.Error(w, "Failed to fetch collection: "+err.Error(), http.StatusInternalServerError)
        return
    }
    photoResponse, err := fetchData[PhotoResponse](fmt.Sprintf("%s/photos/%s", os.Getenv("API_BASE_URL"), photoID))
    if err != nil {
        http.Error(w, "Failed to fetch photo: "+err.Error(), http.StatusInternalServerError)
        return
    }

    var prevID, nextID string
    for idx, photo := range collectionResponse.Photos {
        if photo.ID == photoResponse.Photo.ID {
            if idx > 0 {
                prevID = collectionResponse.Photos[idx-1].ID
            }
            if idx < len(collectionResponse.Photos) - 1 {
                nextID = collectionResponse.Photos[idx+1].ID
            }
        }
    }

    timeString := Photo.Timestamp.Format("2006-01-02 03:04:05")

    data := struct { // build this directly into api??
        Photo               Photo
        Collection          Collection
	TimeString	    string
        PrevID              string
        NextID              string
        OtherPhotos         []Photo
        OtherCollections    []Collection
    }{
        Photo: photoResponse.Photo,
        Collection: collectionResponse.Collection,
	TimeString: timeString,
        PrevID: prevID,
        NextID: nextID,
        OtherPhotos: collectionResponse.Photos,
        OtherCollections: photoResponse.Collections,
    }

    pageBuf, err := fillTemplateWithData("templates/photos-collection-photo.tmpl.html", data)
    if err != nil {
        http.Error(w, "Failed to fill template: "+err.Error(), http.StatusInternalServerError)
        return
    }

    writePageToCache(pageBuf, r.URL.Path)
    w.Write(pageBuf.Bytes())
}
