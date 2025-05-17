package main

import (
	"encoding/json"
	"net/http"

	// "github.com/google/uuid"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)


type Response struct {
    Message string `json:"message"`
}
type PhotoResponse struct {
    Photo       Photo           `json:"photo"`
    Collections []Collection    `json:"collections"`
}
type CollectionResponse struct {
    Collection  Collection    `json:"collection"`
    Photos      []Photo       `json:"photos"`
}


// health check
func rootHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    response := Response{ Message: "hey, get out of here" }
    json.NewEncoder(w).Encode(response)
}


// get all collections + photos
func getCollectionsHandler(w http.ResponseWriter, r *http.Request) {
    query := `
        SELECT
            c.id,
            c.title,
            c.precedence,
            COALESCE(
                json_agg(
                    json_build_object(
                        'id', p.id,
                        'title', p.title,
                        'precedence', cp.precedence
                    ) ORDER BY cp.precedence
                ) FILTER (WHERE p.id IS NOT NULL),
                '[]'::json
            ) as photos
        FROM collections c
        LEFT JOIN collection_photos cp ON c.id = cp.collection_id
        LEFT JOIN photos p ON cp.photo_id = p.id
        GROUP BY c.id
        ORDER BY c.precedence
    `

    rows, err := db.Query(ctx, query)
    if err != nil {
        http.Error(w, "Unable to get all collections: " + err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var fullResponse []CollectionResponse
    for rows.Next() {
        var response CollectionResponse
        var photosJSON []byte

        if err := rows.Scan(&response.Collection.ID, &response.Collection.Title, &response.Collection.Precedence, &photosJSON); err != nil {
            http.Error(w, "Unable to get collection: " + err.Error(), http.StatusInternalServerError)
            return
        }

        if err := json.Unmarshal(photosJSON, &response.Photos); err != nil {
            http.Error(w, "Unable to get photos from collection: " + err.Error(), http.StatusInternalServerError)
            return
        }

        fullResponse = append(fullResponse, response)
    }

    json.NewEncoder(w).Encode(fullResponse)
}

// get collection with photos
func getCollectionByIDHandler(w http.ResponseWriter, r *http.Request) {
    query := `
        SELECT
            c.id,
            c.title,
            c.precedence,
            COALESCE(
                json_agg(
                    json_build_object(
                        'id', p.id,
                        'title', p.title,
                        'precedence', cp.precedence
                    ) ORDER BY cp.precedence
                ) FILTER (WHERE p.id IS NOT NULL),
                '[]'::json
            ) as photos
        FROM collections c
        LEFT JOIN collection_photos cp ON c.id = cp.collection_id
        LEFT JOIN photos p ON cp.photo_id = p.id
        WHERE c.id = $1
        GROUP BY c.id
    `

    collectionID := chi.URLParam(r, "collectionId")

    var response CollectionResponse
    var photosJSON []byte

    if err := db.QueryRow(ctx, query, collectionID).Scan(
        &response.Collection.ID,
        &response.Collection.Title,
        &response.Collection.Precedence,
        &photosJSON,
    ); err != nil {
        if err == pgx.ErrNoRows {
            http.Error(w, "Collection not found: " + err.Error(), http.StatusNotFound)
            return
        }
        http.Error(w, "Unable to get collection: " + err.Error(), http.StatusInternalServerError)
        return
    }

    if err := json.Unmarshal(photosJSON, &response.Photos); err != nil {
        http.Error(w, "Unable to get photos from collection: " + err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(response)
}

// post new collection with list of existing photo ids
func postCollectionHandler(w http.ResponseWriter, r *http.Request) {
}

// edit collection (name, add/remove photos)
func patchCollectionByIDHandler(w http.ResponseWriter, r *http.Request) {
}

// delete collection
func deleteCollectionByIDHandler(w http.ResponseWriter, r *http.Request) {
    query := "DELETE FROM collections WHERE id=$1"
    var response Response

    if err := db.QueryRow(ctx, query).Scan(&response.Message); err != nil {
        if err == pgx.ErrNoRows {
            http.Error(w, "Could not find collection to delete: " + err.Error(), http.StatusNotFound)
            return
        }
        http.Error(w, "Unable to delete collection: " + err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(response)
}


// get all photos
func getPhotosHandler(w http.ResponseWriter, r *http.Request) {
    query := `
        SELECT
            p.id,
            p.title,
            p.timestamp
        FROM photos p
        ORDER BY p.timestamp DESC
    ` // most recent first

    rows, err := db.Query(ctx, query)
    if err != nil {
        http.Error(w, "Unable to get all photos: " + err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var fullResponse []Photo
    for rows.Next() {
        var photo Photo

        if err := rows.Scan(&photo.ID, &photo.Title, &photo.Timestamp); err != nil {
            http.Error(w, "Unable to get photo: " + err.Error(), http.StatusInternalServerError)
            return
        }

        fullResponse = append(fullResponse, photo)
    }
    json.NewEncoder(w).Encode(fullResponse)
}

// get photo with collections
func getPhotoByIDHandler(w http.ResponseWriter, r *http.Request) { 
    query := `
        SELECT
            p.id,
            p.title,
            p.timestamp,
            COALESCE(
                json_agg(
                    json_build_object(
                        'id', c.id,
                        'title', c.title,
                        'precedence', c.precedence
                    ) ORDER BY c.precedence
                ) FILTER (WHERE c.id IS NOT NULL),
                '[]'::json
            ) as collections
        FROM photos p
        LEFT JOIN collection_photos cp ON p.id = cp.photo_id
        LEFT JOIN collections c ON cp.collection_id = c.id
        WHERE p.id = $1
        GROUP BY p.id
    `

    collectionID := chi.URLParam(r, "photoId")

    var response PhotoResponse
    var collectionsJSON []byte

    if err := db.QueryRow(ctx, query, collectionID).Scan(
        &response.Photo.ID,
        &response.Photo.Title,
        &response.Photo.Timestamp,
        &collectionsJSON,
    ); err != nil {
        if err == pgx.ErrNoRows {
            http.Error(w, "Photo not found: " + err.Error(), http.StatusNotFound)
            return
        }
        http.Error(w, "Unable to get photo: " + err.Error(), http.StatusInternalServerError)
        return
    }

    if err := json.Unmarshal(collectionsJSON, &response.Collections); err != nil {
        http.Error(w, "Unable to get collections with photo: " + err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(response)
}

// post photo to postgres and cloudflare
func postPhotoHandler(w http.ResponseWriter, r *http.Request) {
    // generate id
    // uuid := uuid.New()

    // concurrently post to cloudflare and postgres
}

// patch photo (postgres only)
func patchPhotoByIDHandler(w http.ResponseWriter, r *http.Request) {
    // have to work with form data request body
}

// delete photo
func deletePhotoByIDHandler(w http.ResponseWriter, r *http.Request) {
    query := "DELETE FROM photos WHERE id=$1"
    var response Response

    if err := db.QueryRow(ctx, query).Scan(&response.Message); err != nil {
        if err == pgx.ErrNoRows {
            http.Error(w, "Could not find photo to delete: " + err.Error(), http.StatusNotFound)
            return
        }
        http.Error(w, "Unable to delete photo: " + err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(response)
}
