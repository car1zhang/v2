package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
)

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie("token")
		if err != nil {
			if err == http.ErrNoCookie {
				http.Error(w, "Unauthorized request", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Failed to read token: "+err.Error(), http.StatusBadRequest)
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenCookie.Value, claims, func(token *jwt.Token) (any, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid || !claims.Authenticated {
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	log.Print("Starting admin template server")

	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()

	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.Throttle(1000))
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	r.Get("/login", loginHandler)

	r.Route("/collections", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Get("/", collectionsPageHandler)
        r.Get("/new", newCollectionPageHandler)
        r.Get("/{collectionId}", editCollectionPageHandler)
	})

	r.Route("/photos", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Get("/", photosPageHandler)
        r.Get("/new", newPhotoPageHandler)
        r.Get("/{photoId}", editPhotoPageHandler)
	})

	log.Printf("Listening on port %s", os.Getenv("PORT"))
	http.ListenAndServe(":"+os.Getenv("PORT"), r)
}
