package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"learn_go/db"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type DeleteDeckReq struct {
	ID int `json:"id"`
}

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

// RateFlashcardHandler handles POST requests to /api/flashcard/rate
func RateFlashcardHandler(data *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form data", http.StatusBadRequest)
			return
		}

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

		// TODO: Apply when rating system is done (NYI)
		fmt.Printf("ID: %d, Rating: %d\n", id, rating)

	}
}

// Handles GET/POST/DELETE to /api/flashcard/decks
func DeckHandler(data *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
			case http.MethodGet:
				decks, err := db.GetDecksData(data)
				if err != nil {
					http.Error(w, "Error fetching decks", http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")

				if err := json.NewEncoder(w).Encode(decks); err != nil {
					http.Error(w, "Error encoding decks", http.StatusInternalServerError)
					return
				}
			case http.MethodPost:
				// Decode the deck name from the request body
				var deckName struct {
					Name string `json:"name"`
				}
				if err := json.NewDecoder(r.Body).Decode(&deckName); err != nil {
					http.Error(w, "Invalid request body", http.StatusBadRequest)
					return
				}

				// Check if the deck name is empty (TODO: add more validation)
				if deckName.Name == "" {
					http.Error(w, "Deck name cannot be empty", http.StatusBadRequest)
					return
				}

				_, err := db.InsertDeck(data, deckName.Name)
				if err != nil {
					http.Error(w, "Error creating deck", http.StatusInternalServerError)
					return
				}
				response := struct {
					DeckName string
				}{
					DeckName: deckName.Name,
				}
				w.Header().Set("Content-Type", "application/json")

				if err := json.NewEncoder(w).Encode(response); err != nil {
					http.Error(w, "Error encoding decks", http.StatusInternalServerError)
					return
				}
			case http.MethodDelete:
				body, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "Error reading request body", http.StatusBadRequest)
					return
				}

				var req DeleteDeckReq
				err = json.Unmarshal(body, &req)

				if err != nil {
					http.Error(w, "Error unmarshaling req body data", http.StatusBadRequest)
					return
				}

				if req.ID == 0 {
					http.Error(w, "Missing deck ID", http.StatusBadRequest)
					return
				}

				err = db.DeleteDeckByID(data, req.ID)

				if err != nil {
					http.Error(w, "Error deleting deck", http.StatusBadRequest)
					return
				}

				w.WriteHeader(http.StatusNoContent)
		}
	}
}

// Handles GET/POST/DELETE/PUT to /api/flashcard/cards
func CardHandler(data *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
			//TODO: test this implementation (for unifying cards and cards/ endpoints)
			case http.MethodGet:
				var deckID struct {
					ID int `json:"id"`
				}

				if err := json.NewDecoder(r.Body).Decode(&deckID); err != nil {
					http.Error(w, "Invalid request body", http.StatusBadRequest)
					return
				}

				// validation
				if deckID.ID < 0 {
					http.Error(w, "Invalid deck ID", http.StatusBadRequest)
					return
				}

				cards, err := db.GetCardsFromDeck(data, deckID.ID)

				if err == nil {
					http.Error(w, "Error getting cards from deck", http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")

				if err := json.NewEncoder(w).Encode(cards); err != nil {
					http.Error(w, "Error encoding cards", http.StatusInternalServerError)
					return
				}
			case http.MethodPost:
				var cardData struct {
					Front string `json:"front"`
					Back  string `json:"back"`
				}
				if err := json.NewDecoder(r.Body).Decode(&cardData); err != nil {
					http.Error(w, "Invalid request body", http.StatusConflict)
					fmt.Printf("Error decoding request body: %v\n", err)
					return
				}

				// Validate card data (TODO: add more checks)
				if cardData.Front == "" || cardData.Back == "" {
					http.Error(w, "Front and back content cannot be empty", http.StatusBadRequest)
					fmt.Printf("Front and back content cannot be empty\n")
					return
				}

				newCard, err := db.CreateCard(0, cardData.Front, cardData.Back, int64(time.Now().Nanosecond()), 0)
				if err != nil {
					http.Error(w, "Error creating card", http.StatusInternalServerError)
					return
				}

				// Insert the card and get the ID
				insertedIDs, err := db.InsertCards(data, []db.Card{newCard})
				if err != nil {
					http.Error(w, "Error inserting card", http.StatusInternalServerError)
					return
				}

				// Get the deck ID from the URL path (assuming the path is /projects/flashcard/edit/{deck_id})
				parts := strings.Split(r.Header.Get("Referer"), "/")
				deckID, err := strconv.Atoi(parts[6])

				if err != nil {
					http.Error(w, "Invalid deck ID in URL", http.StatusBadRequest)
					return
				}

				// Add the card to the deck
				if len(insertedIDs) > 0 { // Check if we got an ID back
					cardID := insertedIDs[0]
					log.Printf("Adding card with ID %d to deck %d\n", cardID, deckID)
					err = db.AddCardToDeck(data, int(cardID), deckID)
					if err != nil {
						http.Error(w, "Error adding card to deck", http.StatusInternalServerError)
						log.Print(err)
						return
					}
				} else {
					// Handle the case where no ID was returned (this shouldn't happen if InsertCards is working correctly)
					http.Error(w, "Card created but ID not found", http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				response := struct {
					Message string  `json:"message"`
					Card    db.Card `json:"card"`
				}{
					Message: "Card created and added to deck successfully",
					Card:    newCard,
				}
				if err := json.NewEncoder(w).Encode(response); err != nil {
					http.Error(w, "Error encoding response", http.StatusInternalServerError)
					return
				}
			case http.MethodDelete:
				var cardData struct {
					ID int `json:"id"`
				}
				if err := json.NewDecoder(r.Body).Decode(&cardData); err != nil {
					http.Error(w, "Invalid request body", http.StatusBadRequest)
					return
				}

				// Validate card ID (TODO: add more checks)
				if cardData.ID <= 0 {
					http.Error(w, "Invalid card ID", http.StatusBadRequest)
					return
				}

				err := db.DeleteCardByID(data, cardData.ID)
				if err != nil {
					http.Error(w, "Error deleting card", http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				response := struct {
					Message string `json:"message"`
				}{
					Message: "Card deleted successfully",
				}
				if err := json.NewEncoder(w).Encode(response); err != nil {
					http.Error(w, "Error encoding decks", http.StatusInternalServerError)
					return
				}
			case http.MethodPut:
				var updatedCard db.Card
				if err := json.NewDecoder(r.Body).Decode(&updatedCard); err != nil {
					http.Error(w, "Invalid request body", http.StatusBadRequest)
					return
				}
				// Validate card data (similar to how you validate in POST)
				// TODO: Think abt extracting validation as a separate method
				if updatedCard.Front == "" || updatedCard.Back == "" || updatedCard.ID == 0 {
					http.Error(w, "Front, back, and ID are required", http.StatusBadRequest)
					return
				}

				err := db.UpdateCard(data, updatedCard)
				if err != nil {
					http.Error(w, "Error updating card", http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				response := struct {
					Message string  `json:"message"`
					Card    db.Card `json:"card"`
				}{
					Message: "Card updated successfully",
					Card:    updatedCard,
				}
				if err := json.NewEncoder(w).Encode(response); err != nil {
					http.Error(w, "Error encoding response", http.StatusInternalServerError)
					return
				}
		}
	}
}
