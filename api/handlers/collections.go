package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"api/models"
	"api/services"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// get all collections + photos
func GetCollectionsHandler(w http.ResponseWriter, r *http.Request) {
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

	rows, err := services.DB.Query(context.Background(), query)
	if err != nil {
		http.Error(w, "Failed to get all collections: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var fullResponse []models.CollectionResponse
	for rows.Next() {
		var response models.CollectionResponse
		var photosJSON []byte

		if err := rows.Scan(&response.Collection.ID, &response.Collection.Title, &response.Collection.Precedence, &photosJSON); err != nil {
			http.Error(w, "Failed to get collection: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.Unmarshal(photosJSON, &response.Photos); err != nil {
			http.Error(w, "Failed to get photos from collection: "+err.Error(), http.StatusInternalServerError)
			return
		}

		fullResponse = append(fullResponse, response)
	}

	json.NewEncoder(w).Encode(fullResponse)
}

// get collection with photos
func GetCollectionByIDHandler(w http.ResponseWriter, r *http.Request) {
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

	var response models.CollectionResponse
	var photosJSON []byte

	if err := services.DB.QueryRow(context.Background(), query, chi.URLParam(r, "collectionId")).Scan(
		&response.Collection.ID,
		&response.Collection.Title,
		&response.Collection.Precedence,
		&photosJSON,
	); err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Collection not found: "+err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get collection: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(photosJSON, &response.Photos); err != nil {
		http.Error(w, "Failed to get photos from collection: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(response)
}

// post new collection with list of existing photo ids
func PostCollectionHandler(w http.ResponseWriter, r *http.Request) {
	query := "INSERT INTO collections (id, title, precedence) VALUES ($1, $2, $3)"

	collectionID := uuid.NewString()

	_, err := services.DB.Exec(context.Background(), query, collectionID, r.FormValue("title"), r.FormValue("precedence"))
	if err != nil {
		http.Error(w, "Failed to insert collection: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// insert photos
	photoIDs := r.Form["photos[]"]
	batch := &pgx.Batch{}
	for idx, photoID := range photoIDs {
		batch.Queue(
			"INSERT INTO collection_photos (collection_id, photo_id, precedence) VALUES ($1, $2, $3)",
			collectionID, photoID, idx,
		)
	}
	photosResults := services.DB.SendBatch(context.Background(), batch)
	defer photosResults.Close()
	for range batch.Len() {
		_, err := photosResults.Exec()
		if err != nil {
			http.Error(w, "Failed to insert collection photos: "+err.Error(), http.StatusBadRequest)
		}
	}

	if err := services.InvalidateClientCache(); err != nil {
		http.Error(w, "Failed to clear client cache: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(models.RestResponse{ // TODO: update this to include created photo junctions?
		ID:      collectionID,
		Success: true,
	})
}

// edit collection (name, add/remove photos)
func PatchCollectionByIDHandler(w http.ResponseWriter, r *http.Request) { // TODO: updating photo precedence
	query := "UPDATE collections SET title = $1, precedence = $2 WHERE id = $3 RETURNING *"

	var id string
	if err := services.DB.QueryRow(context.Background(), query, r.FormValue("title"), r.FormValue("precedence"), chi.URLParam(r, "collectionId")).Scan(&id, nil, nil); err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Collection not found: "+err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to update collection: "+err.Error(), http.StatusInternalServerError)
		return
	}

	photoIDs := r.Form["photos[]"] // TODO: better way of doing this?
	batch := &pgx.Batch{}
	batch.Queue(
		"DELETE FROM collection_photos WHERE collection_id=$1",
		chi.URLParam(r, "collectionId"),
	)
	for idx, photoID := range photoIDs {
		batch.Queue(
			"INSERT INTO collection_photos (collection_id, photo_id, precedence) VALUES ($1, $2, $3)",
			chi.URLParam(r, "collectionId"), photoID, idx,
		)
	}
	photosResults := services.DB.SendBatch(context.Background(), batch)
	defer photosResults.Close()
	for range batch.Len() {
		_, err := photosResults.Exec()
		if err != nil {
			http.Error(w, "Failed to modify collection photos: "+err.Error(), http.StatusBadRequest)
		}
	}

	if err := services.InvalidateClientCache(); err != nil {
		http.Error(w, "Failed to clear client cache: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(models.RestResponse{ // TODO: same as post TODO
		ID:      id,
		Success: true,
	})
}

// delete collection
func DeleteCollectionByIDHandler(w http.ResponseWriter, r *http.Request) {
	query := "DELETE FROM collections WHERE id=$1 RETURNING *"

	var id string
	if err := services.DB.QueryRow(context.Background(), query, chi.URLParam(r, "collectionId")).Scan(&id, nil, nil); err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Collection not found: "+err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete collection: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := services.InvalidateClientCache(); err != nil {
		http.Error(w, "Failed to clear client cache: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(models.RestResponse{
		ID:      id,
		Success: true,
	})
}
