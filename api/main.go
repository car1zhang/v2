package main

import (
    "os"
    "log"
    "time"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/joho/godotenv"
)

func main() {
    log.Print("Starting API server")

    if err := godotenv.Load(); err != nil {
        log.Fatal(err)
    }
    if err := initializePhotosDB(); err != nil {
        log.Fatal(err)
    }

    r := chi.NewRouter()

    r.Use(middleware.Throttle(1000))
	r.Use(middleware.Timeout(60 * time.Second))
    r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.SetHeader("Content-Type", "application/json"))

    r.Get("/", rootHandler)

    r.Route("/v1", func(r chi.Router) {
        r.Route("/collections", func(r chi.Router) {
            r.Get("/", getCollectionsHandler)
            r.Get("/{collectionId}", getCollectionByIDHandler) // edit id for more descriptive (match web)
            r.Post("/", postCollectionHandler)
            r.Patch("/{collectionId}", patchCollectionByIDHandler)
            r.Delete("/{collectionId}", deleteCollectionByIDHandler)
        })

        r.Route("/photos", func(r chi.Router) {
            r.Get("/", getPhotosHandler)
            r.Get("/{photoId}", getPhotoByIDHandler)
            r.Post("/", postPhotoHandler)
            r.Patch("/{photoId}", patchPhotoByIDHandler)
            r.Delete("/{photoId}", deletePhotoByIDHandler)
        })
    })

    log.Printf("Listening on port %s", os.Getenv("PORT"))
    http.ListenAndServe(":"+os.Getenv("PORT"), r)
}
