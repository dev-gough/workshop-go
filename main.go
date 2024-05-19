package main

import (
	"fmt"
	"learn_go/components"
	"learn_go/components/handlers"
	"learn_go/db"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/a-h/templ"
)

type UpdateDifficultyRequest struct {
	ID     int `json:"id"`
	Rating int `json:"rating"`
}

// Create a couple of mock flashcards
var mockFlashcards = []db.Card{
	{
		ID:         1,
		Front:      "What is the capital of France?",
		Back:       "Paris",
		Reviewed:   1625179200, // Timestamp for July 1, 2021, 12:00 AM UTC
		Difficulty: 3,
	},
	{
		ID:         2,
		Front:      "What is the largest country in the world?",
		Back:       "Russia",
		Reviewed:   1625265600, // Timestamp for July 2, 2021, 12:00 AM UTC
		Difficulty: 4,
	},
	{
		ID:         3,
		Front:      "What is the capital of Germany?",
		Back:       "Berlin",
		Reviewed:   1625352000, // Timestamp for July 3, 2021, 12:00 AM UTC
		Difficulty: 2,
	},
	{
		ID:         4,
		Front:      "What is the capital of Spain?",
		Back:       "Madrid",
		Reviewed:   1625438400, // Timestamp for July 4, 2021, 12:00 AM UTC
		Difficulty: 5,
	},
	{
		ID:         5,
		Front:      "What is the population of the United States?",
		Back:       "331,449,281",
		Reviewed:   1625524800, // Timestamp for July 5, 2021, 12:00 AM UTC
		Difficulty: 3,
	},
	{
		ID:         6,
		Front:      "What is the currency of Japan?",
		Back:       "Japanese yen",
		Reviewed:   1625611200, // Timestamp for July 6, 2021, 12:00 AM UTC
		Difficulty: 4,
	},
	{
		ID:         7,
		Front:      "What is the capital of China?",
		Back:       "Beijing",
		Reviewed:   1625697600, // Timestamp for July 7, 2021, 12:00 AM UTC
		Difficulty: 2,
	},
	{
		ID:         8,
		Front:      "What is the largest city in the United States?",
		Back:       "New York City",
		Reviewed:   1625784000, // Timestamp for July 8, 2021, 12:00 AM UTC
		Difficulty: 5,
	},
	{
		ID:         9,
		Front:      "What is the area of the United States?",
		Back:       "9,833,520 km²",
		Reviewed:   1625870400, // Timestamp for July 9, 2021, 12:00 AM UTC
		Difficulty: 3,
	},
	{
		ID:         10,
		Front:      "What is the capital of Australia?",
		Back:       "Canberra",
		Reviewed:   1625956800, // Timestamp for July 10, 2021, 12:00 AM UTC
		Difficulty: 4,
	},
}


func updateDifficulty(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	// Now you can use `id` and `rating` in your logic
	fmt.Printf("ID: %d, Rating: %d\n", id, rating)

	// TODO: Push this new data to database

	// Respond to the request
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Difficulty updated successfully"))
}

func setHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request is for a JavaScript file
		if strings.HasSuffix(r.URL.Path, ".js") {
			// Set the Content-Type header
			w.Header().Set("Content-Type", "application/javascript")
		}
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func main() {
	database, _ := db.ConnectToDB()
	defer database.Close()

	// Drop all tables
	db.DropAllTables(database)
	db.CreateAllTables(database, db.CurrentTables)
	db.InsertCards(database, mockFlashcards)
	db.InsertDeck(database, "Countries and Capitals")
	db.InsertDeck(database, "Geography")
	db.AddCardToDeck(database, 1, 1)
	db.AddCardToDeck(database, 2, 1)
	db.AddCardToDeck(database, 3, 2)

	http.Handle("/projects/gol", templ.Handler(components.GOLPage()))
	http.Handle("/home", templ.Handler(components.Home()))
	http.Handle("/projects/flashcard", templ.Handler(components.Decks()))
	http.Handle("/projects/flashcard/random", templ.Handler(components.Flashcard()))

	http.HandleFunc("/api/flashcard", handlers.RandomFlashcardHandler(database))
	http.HandleFunc("/api/flashcard/rate", handlers.RateFlashcardHandler(database))
	http.HandleFunc("/api/flashcard/cards/", handlers.GetCardsForDeckHandler(database))
	http.HandleFunc("/api/flashcard/decks", handlers.GetDecksHandler(database))
	http.HandleFunc("/api/flashcard/createdeck", handlers.NewDeckHandler(database))

	fs := http.FileServer(http.Dir("public"))
	http.Handle("/static/", setHeaderMiddleware(http.StripPrefix("/static/", fs)))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
