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
    log.Print("Starting template server")

    if err := godotenv.Load(); err != nil {
        log.Fatal(err)
    }

    go listenCacheInvalidation()

    r := chi.NewRouter()

	r.Use(middleware.Timeout(60 * time.Second))
    r.Use(middleware.Throttle(1000))
    r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

    r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

    r.Get("/", homePageHandler)

    r.Route("/photos", func (r chi.Router) {
        r.Get("/", photoHomePageHandler) // maybe move to subdomain instead of routes (when blog/emulator is done too)
        r.Get("/{photoId}", photoPageHandler) // error template page instead of sending errors directly back to browser

        r.Get("/all", photoAllPageHandler)
        r.Get("/all/{photoId}", photoFromAllPageHandler)

        r.Get("/collection/{collectionId}", photoCollectionPageHandler)
        r.Get("/collection/{collectionId}/{photoId}", photoFromCollectionPageHandler)
    })


    log.Printf("Listening on port %s", os.Getenv("PORT"))
    http.ListenAndServe(":" + os.Getenv("PORT"), r)
}
