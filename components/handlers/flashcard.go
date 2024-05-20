package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"learn_go/db"
	"net/http"
	"strconv"
	"strings"
)

// RandomFlashcardHandler handles a GET request to /api/flashcard,
// returning a random flashcard from all decks
func RandomFlashcardHandler(data *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		card, err := db.GetRandomCard(data)
		if err != nil {
			http.Error(w, "Error fetching card", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json") // Set JSON content type

		if err := json.NewEncoder(w).Encode(card); err != nil { // Encode card as JSON
			http.Error(w, "Error encoding card", http.StatusInternalServerError)
			return
		}
	}
}

func GetCardsForDeckHandler(data *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		parts := strings.Split(r.URL.Path, "/")
		id := parts[4]
		deckID, err := strconv.Atoi(id)
		if err != nil {
			http.Error(w, "Invalid deck ID", http.StatusBadRequest)
			return
		}

		cards, err := db.GetCardsFromDeck(data, deckID)
		if err != nil {
			http.Error(w, "Error fetching cards", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json") // Set JSON content type

		if err := json.NewEncoder(w).Encode(cards); err != nil { // Encode cards as JSON
			http.Error(w, "Error encoding cards", http.StatusInternalServerError)
			return
		}
	}
}

func GetDecksHandler(data *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		decks, err := db.GetDecksData(data)
		if err != nil {
			http.Error(w, "Error fetching decks", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json") // Set JSON content type

		if err := json.NewEncoder(w).Encode(decks); err != nil { // Encode decks as JSON
			http.Error(w, "Error encoding decks", http.StatusInternalServerError)
			return
		}
	}
}

// RateFlashcardHandler handles POST requests to /api/flashcard/rate
func RateFlashcardHandler(data *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form data", http.StatusBadRequest)
			return
		}

		// Access form values
		idStr := r.FormValue("ID")
		ratingStr := r.FormValue("Rating")

		// Convert form values to appropriate types
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		rating, err := strconv.Atoi(ratingStr)
		if err != nil {
			http.Error(w, "Invalid Rating", http.StatusBadRequest)
			return
		}

		// Print form values
		fmt.Printf("ID: %d, Rating: %d\n", id, rating)

	}
}

func DeckHandler(data *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// Decode the deck name from the request body
			var deckName struct {
				Name string `json:"name"`
			}
			if err := json.NewDecoder(r.Body).Decode(&deckName); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			// Check if the deck name is empty (add validation as needed)
			if deckName.Name == "" {
				http.Error(w, "Deck name cannot be empty", http.StatusBadRequest)
				return
			}

			// Insert the deck into the database
			err := db.InsertDeck(data, deckName.Name)
			if err != nil {
				http.Error(w, "Error creating deck", http.StatusInternalServerError)
				return
			}

			// Optionally: You could return the ID of the newly created deck
			response := struct {
				DeckName string
			}{
				DeckName: deckName.Name,
			}
			w.Header().Set("Content-Type", "application/json")

			if err := json.NewEncoder(w).Encode(response); err != nil { // Encode decks as JSON
				http.Error(w, "Error encoding decks", http.StatusInternalServerError)
				return
			}
		} else if r.Method == http.MethodDelete {
			// Get the deck ID from the URL parameters (e.g., /api/flashcard/decks/123)
			parts := strings.Split(r.URL.Path, "/")
			id := parts[4]
			deckID, err := strconv.Atoi(id)
			if err != nil {
				http.Error(w, "Invalid deck ID", http.StatusBadRequest)
				return
			}

			// Delete the deck from the database
			err = db.DeleteDeckByID(data, deckID)
			if err != nil {
				http.Error(w, "Error deleting deck", http.StatusInternalServerError)
				return
			}

			// Respond with a success message or status
			w.WriteHeader(http.StatusNoContent) // 204 No Content is common for successful DELETE
		}
	}
}
