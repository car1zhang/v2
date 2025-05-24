package handlers

import (
    "io"
    "os"
    "log"
    "bytes"
	"context"
	"encoding/json"
	"net/http"

	"api/models"
	"api/services"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
    "github.com/rwcarlsen/goexif/exif"
)

// get all photos
func GetPhotosHandler(w http.ResponseWriter, r *http.Request) {
    query := `
        SELECT
            p.id,
            p.title,
            p.timestamp
        FROM photos p
        ORDER BY p.timestamp DESC
    ` // most recent first

    rows, err := services.DB.Query(context.Background(), query)
    if err != nil {
        http.Error(w, "Failed to get all photos: " + err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var fullResponse []models.Photo
    for rows.Next() {
        var photo models.Photo

        if err := rows.Scan(&photo.ID, &photo.Title, &photo.Timestamp); err != nil {
            http.Error(w, "Failed to get photo: " + err.Error(), http.StatusInternalServerError)
            return
        }

        fullResponse = append(fullResponse, photo)
    }

    json.NewEncoder(w).Encode(fullResponse)
}

// get photo with collections
func GetPhotoByIDHandler(w http.ResponseWriter, r *http.Request) { 
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

    var response models.PhotoResponse
    var collectionsJSON []byte

    if err := services.DB.QueryRow(context.Background(), query, chi.URLParam(r, "photoId")).Scan(
        &response.Photo.ID,
        &response.Photo.Title,
        &response.Photo.Timestamp,
        &collectionsJSON,
    ); err != nil {
        if err == pgx.ErrNoRows {
            http.Error(w, "Photo not found: " + err.Error(), http.StatusNotFound)
            return
        }
        http.Error(w, "Failed to get photo: " + err.Error(), http.StatusInternalServerError)
        return
    }

    if err := json.Unmarshal(collectionsJSON, &response.Collections); err != nil {
        http.Error(w, "Failed to get collections with photo: " + err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(response)
}

// post photo to postgres and cloudflare
func PostPhotoHandler(w http.ResponseWriter, r *http.Request) {
    r.Body = http.MaxBytesReader(w, r.Body, 50 << 40)

    err := r.ParseMultipartForm(50 << 40)
    if err != nil {
        http.Error(w, "Failed to parse request: " + err.Error(), http.StatusBadRequest)
        return
    }

    file, _, err := r.FormFile("image")
    if err != nil {
        http.Error(w, "Failed to get file: " + err.Error(), http.StatusBadRequest)
        return
    }
    defer file.Close()

    // metadata
    fileBytes, err := io.ReadAll(file)
    if err != nil {
        http.Error(w, "Failed to read file: "+err.Error(), http.StatusBadRequest)
        return
    }
    exifReader := bytes.NewReader(fileBytes)
    metadata, err := exif.Decode(exifReader)
    if err != nil {
        http.Error(w,"Failed to decode image metadata: " + err.Error(), http.StatusBadRequest)
    }
    dateTime, err := metadata.DateTime()
    if err != nil {
        http.Error(w, "Failed to get timestamp from EXIF: " + err.Error(), http.StatusBadRequest)
        return
    }

    // temp file
    tempFile, err := os.CreateTemp("", "upload-*.jpg")
    if err != nil {
        http.Error(w, "Failed to create temp file: " + err.Error(), http.StatusInternalServerError)
        return
    }
    defer os.Remove(tempFile.Name())
    defer tempFile.Close()
    _, err = tempFile.Write(fileBytes)
    if err != nil {
        http.Error(w, "Failed to save file: " + err.Error(), http.StatusInternalServerError)
        return
    }

    // upload
    imageID, err := services.UploadImageToCloudflare(tempFile.Name())
    if err != nil {
        http.Error(w, "Failed to upload to Cloudflare: " + err.Error(), http.StatusInternalServerError)
        return
    }

    // postgres
    log.Println(dateTime)
    query := "INSERT INTO photos (id, title, timestamp) VALUES ($1, $2, $3)"
    _, err = services.DB.Exec(context.Background(), query, imageID, r.FormValue("title"), dateTime.Format("2006-01-02 15:04:05"))
    if err != nil {
        http.Error(w, "Failed to insert photo: " + err.Error(), http.StatusInternalServerError)
        return
    }

    if err := services.InvalidateClientCache(); err != nil {
        http.Error(w, "Failed to clear client cache: " + err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(models.RestResponse{
        ID: imageID,
        Success: true,
    })
}

// patch photo (postgres only)
func PatchPhotoByIDHandler(w http.ResponseWriter, r *http.Request) {
    query := "UPDATE photos SET title = $1 WHERE id = $2 RETURNING *"

    var id string
    if err := services.DB.QueryRow(context.Background(), query, r.FormValue("title"), chi.URLParam(r, "photoId")).Scan(&id, nil, nil); err != nil {
        if err == pgx.ErrNoRows {
            http.Error(w, "Photo not found: " + err.Error(), http.StatusNotFound)
            return
        }
        http.Error(w, "Failed to update photo: " + err.Error(), http.StatusInternalServerError)
        return
    }

    if err := services.InvalidateClientCache(); err != nil {
        http.Error(w, "Failed to clear client cache: " + err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(models.RestResponse{
        ID: id,
        Success: true,
    })
}

// delete photo
func DeletePhotoByIDHandler(w http.ResponseWriter, r *http.Request) {
    query := "DELETE FROM photos WHERE id=$1 RETURNING *"

    var id string
    if err := services.DB.QueryRow(context.Background(), query, chi.URLParam(r, "photoId")).Scan(&id, nil, nil); err != nil {
        if err == pgx.ErrNoRows {
            http.Error(w, "Photo not found: " + err.Error(), http.StatusNotFound)
            return
        }
        http.Error(w, "Failed to delete photo: " + err.Error(), http.StatusInternalServerError)
        return
    }

    if err := services.DeleteImageFromCloudflare(chi.URLParam(r, "photoId")); err != nil {
        http.Error(w, "Failed to delete image from Cloudflare: " + err.Error(), http.StatusInternalServerError)
        return
    }

    if err := services.InvalidateClientCache(); err != nil {
        http.Error(w, "Failed to clear client cache: " + err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(models.RestResponse{
        ID: id,
        Success: true,
    })
}
