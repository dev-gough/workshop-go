package main

import (
	"encoding/json"
	"learn_go/components"
	"learn_go/components/handlers"
	"learn_go/db"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

func ListPatternFiles(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir("./static/patterns")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var names []string

	for _, file := range files {
		if !file.IsDir() {
			names = append(names, file.Name())
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(names); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// TODO: move to handlers package
func GetFileContents(w http.ResponseWriter, r *http.Request) {
	const patternDir = "./static/patterns/"
	prefix := "/api/gol/patterns/"
	fileName := strings.TrimPrefix(r.URL.Path, prefix)
	if fileName == "" {
		http.Error(w, "No file name provided", http.StatusBadRequest)
		return
	}

	path := filepath.Join(patternDir, fileName)

	// check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		http.Error(w, "File does not exist", http.StatusNotFound)
		return
	}

	// convert file to json
	contents, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := map[string]string{
		"filename": fileName,
		"contents": string(contents),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
	http.HandleFunc("/api/gol/patterns", ListPatternFiles)
	http.HandleFunc("/api/gol/patterns/", GetFileContents)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)
	http.Handle("/static/", setHeaderMiddleware(http.StripPrefix("/static/", fs)))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
