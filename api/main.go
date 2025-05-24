package main

import (
    "os"
    "log"
    "time"
    "net/http"

    "api/handlers"
    "api/services"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/joho/godotenv"
)

func main() {
    log.Print("Starting API server")

    if err := godotenv.Load(); err != nil {
        log.Fatal(err)
    }
    services.InitializePhotosDB();

    r := chi.NewRouter()

    r.Use(middleware.Throttle(1000))
	r.Use(middleware.Timeout(60 * time.Second))
    r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.SetHeader("Content-Type", "application/json"))

    r.Get("/", handlers.RootHandler)

    r.Route("/v1", func(r chi.Router) {
        r.Post("/login", handlers.LoginHandler)

        r.Route("/collections", func(r chi.Router) {
            r.Get("/", handlers.GetCollectionsHandler)
            r.Get("/{collectionId}", handlers.GetCollectionByIDHandler)

            r.Group(func(r chi.Router) {
                r.Use(handlers.AuthMiddleware)

                r.Post("/", handlers.PostCollectionHandler)
                r.Patch("/{collectionId}", handlers.PatchCollectionByIDHandler)
                r.Delete("/{collectionId}", handlers.DeleteCollectionByIDHandler)
            })
        })

        r.Route("/photos", func(r chi.Router) {
            r.Get("/", handlers.GetPhotosHandler)
            r.Get("/{photoId}", handlers.GetPhotoByIDHandler)

            r.Group(func(r chi.Router) {
                r.Use(handlers.AuthMiddleware)

                r.Post("/", handlers.PostPhotoHandler)
                r.Patch("/{photoId}", handlers.PatchPhotoByIDHandler)
                r.Delete("/{photoId}", handlers.DeletePhotoByIDHandler)
            })
        })
    })

    log.Printf("Listening on port %s", os.Getenv("PORT"))
    http.ListenAndServe(":"+os.Getenv("PORT"), r)
}
