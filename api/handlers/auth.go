package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"time"
    "log"

	"api/models"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

// health check
func RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := models.Response{
		Message: "hey, get out of here",
		Success: true,
	}
	json.NewEncoder(w).Encode(response)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	password := r.FormValue("password")

	if err := bcrypt.CompareHashAndPassword([]byte(os.Getenv("PASSWORD_HASH")), []byte(password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(time.Hour * 48)
	claims := &models.Claims{
		Authenticated: true,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		http.Error(w, "Failed to generate token: "+err.Error(), http.StatusInternalServerError)
		return
	}
    log.Println(tokenString)

	cookie := http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Expires:  expirationTime,
		HttpOnly: true,
		Secure:   false, // enable in prod
		Path:     "/",
	}
	http.SetCookie(w, &cookie)

	json.NewEncoder(w).Encode(models.Response{
		Message: "Success",
		Success: true,
	})
}

func AuthMiddleware(next http.Handler) http.Handler {
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

		claims := &models.Claims{}
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
