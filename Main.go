package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// Favorite represents a struct with the specified attributes.
type Favorite struct {
	FavoriteID           int    `json:"favorite_id"`
	SessionID            string `json:"session_id"`
	UserName             string `json:"user_name"`
	Name                 string `json:"name"`
	ProfessionalHeadline string `json:"professional_headline"`
	ImgURL               string `json:"img_url"`
}

// Database connection string for PostgreSQL.
const connectionString = "host=torre-assessment.cpmmssgwhxrp.us-east-2.rds.amazonaws.com port=5432 user=postgres  password=postgres dbname=torreassessment sslmode=disable"

func main() {
	r := mux.NewRouter()
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Routes
	r.HandleFunc("/favorites/by-session/{session_id}", getFavoritesBySessionHandler(db)).Methods("GET")
	r.HandleFunc("/favorites", getAllFavoritesHandler(db)).Methods("GET")
	r.HandleFunc("/favorites", createFavoriteHandler(db)).Methods("POST")
	r.HandleFunc("/favorites/{id:[0-9]+}", updateFavoriteHandler(db)).Methods("PUT")
	r.HandleFunc("/favorites/{id:[0-9]+}", deleteFavoriteHandler(db)).Methods("DELETE")

	http.Handle("/", r)
	fmt.Println("Server is running on :8080")
	http.ListenAndServe(":8080", nil)
}

func getAllFavoritesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM favorites")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var favorites []Favorite
		for rows.Next() {
			var f Favorite
			err := rows.Scan(&f.FavoriteID, &f.SessionID, &f.UserName, &f.Name, &f.ProfessionalHeadline, &f.ImgURL)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			favorites = append(favorites, f)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(favorites)
	}
}

func createFavoriteHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var fav Favorite
		if err := json.NewDecoder(r.Body).Decode(&fav); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		insertSQL := `
		INSERT INTO favorites (session_id, user_name, name, professional_headline, img_url)
		VALUES ($1, $2, $3, $4, $5) RETURNING favorite_id`
		err := db.QueryRow(insertSQL, fav.SessionID, fav.UserName, fav.Name, fav.ProfessionalHeadline, fav.ImgURL).
			Scan(&fav.FavoriteID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(fav)
	}
}
func getFavoritesBySessionHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		sessionID := vars["session_id"]

		rows, err := db.Query("SELECT * FROM favorites WHERE session_id = $1", sessionID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var favorites []Favorite
		for rows.Next() {
			var f Favorite
			err := rows.Scan(&f.FavoriteID, &f.SessionID, &f.UserName, &f.Name, &f.ProfessionalHeadline, &f.ImgURL)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			favorites = append(favorites, f)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(favorites)
	}
}
func updateFavoriteHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var fav Favorite
		if err := json.NewDecoder(r.Body).Decode(&fav); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		updateSQL := `
		UPDATE favorites
		SET session_id = $1, user_name = $2, name = $3, professional_headline = $4, img_url = $5
		WHERE favorite_id = $6`
		_, err = db.Exec(updateSQL, fav.SessionID, fav.UserName, fav.Name, fav.ProfessionalHeadline, fav.ImgURL, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func deleteFavoriteHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		deleteSQL := "DELETE FROM favorites WHERE favorite_id = $1"
		_, err = db.Exec(deleteSQL, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
