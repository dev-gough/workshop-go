package main

import (
	"learn_go/components"
	"learn_go/components/handlers"
	"learn_go/db"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/a-h/templ"
)

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

type dynamicHandler struct {
	pattern *regexp.Regexp
	handler func(http.ResponseWriter, *http.Request)
}

func (h dynamicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.pattern.MatchString(r.URL.Path) {
		h.handler(w, r)
		return
	}
	http.NotFound(w, r)
}

func StudyHandler(w http.ResponseWriter, r *http.Request) {
	templ.Handler(components.Study()).ServeHTTP(w, r)
}
func EditHandler(w http.ResponseWriter, r *http.Request) {
	templ.Handler(components.EditDeck()).ServeHTTP(w, r)
}

func main() {
	database, _ := db.ConnectToDB()
	defer database.Close()

	_ = db.CreateAllTables(database, db.CurrentTables)

	http.Handle("/projects/gol", templ.Handler(components.GOLPage()))
	http.Handle("/home", templ.Handler(components.Home()))
	http.Handle("/projects/flashcard", templ.Handler(components.Decks()))
	http.Handle("/projects/flashcard/random", templ.Handler(components.Flashcard()))

	http.Handle("/projects/flashcard/decks/", dynamicHandler{
		pattern: regexp.MustCompile(`^/projects/flashcard/decks/(\d+)/study`),
		handler: StudyHandler,
	})

	http.Handle("/projects/flashcard/edit/", dynamicHandler{
		pattern: regexp.MustCompile(`^/projects/flashcard/edit/(\d+)`),
		handler: EditHandler,
	})

	http.HandleFunc("/api/flashcard", handlers.RandomFlashcardHandler(database))
	http.HandleFunc("/api/flashcard/rate", handlers.RateFlashcardHandler(database))
	http.HandleFunc("/api/flashcard/cards/", handlers.GetCardsForDeckHandler(database))
	http.HandleFunc("/api/flashcard/decks", handlers.GetDecksHandler(database))
	http.HandleFunc("/api/flashcard/decks/", handlers.DeckHandler(database))
	http.HandleFunc("/api/flashcard/cards", handlers.CardHandler(database))

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)
	http.Handle("/static/", setHeaderMiddleware(http.StripPrefix("/static/", fs)))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
