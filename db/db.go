package db

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"

	_ "github.com/lib/pq"
)

type Card struct {
	ID         int    `json:"id"`
	Front      string `json:"front"`
	Back       string `json:"back"`
	Reviewed   int64  `json:"reviewed"`
	Difficulty int    `json:"difficulty"`
}

type Deck struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Option func(*dbOptions)

type dbOptions struct {
	psqlInfo string
}

type TableSchema struct {
	Name      string
	CreateSQL string
}

var CardsTable = TableSchema{
	Name: "cards",
	CreateSQL: `CREATE TABLE IF NOT EXISTS cards (
        id SERIAL PRIMARY KEY,
        front TEXT NOT NULL,
        back TEXT NOT NULL,
        recency BIGINT NOT NULL,
        prevdifficulty INT NOT NULL
    );`,
}

var DecksTable = TableSchema{
	Name: "decks",
	CreateSQL: `CREATE TABLE IF NOT EXISTS decks (
        id SERIAL PRIMARY KEY,
        name TEXT NOT NULL
		);`,
}

var DeckCardsTable = TableSchema{
	Name: "deck_cards",
	CreateSQL: `CREATE TABLE IF NOT EXISTS deck_cards (
        card_id INT NOT NULL,
        deck_id INT NOT NULL,
        PRIMARY KEY (card_id, deck_id),
        FOREIGN KEY (card_id) REFERENCES cards(id) ON DELETE CASCADE,
        FOREIGN KEY (deck_id) REFERENCES decks(id) ON DELETE CASCADE
    );`,
}

var CurrentTables = []TableSchema{CardsTable, DecksTable, DeckCardsTable}

func CreateCard(id int, front string, back string, reviewed int64, difficulty int) (Card, error) {
	card := Card{
		ID:         id,
		Front:      front,
		Back:       back,
		Reviewed:   reviewed,
		Difficulty: difficulty,
	}

	if card.ID < 0 || card.Front == "" || card.Back == "" {
		return Card{}, errors.New("invalid card data")
	}

	return card, nil
}

func CreateDeck(id int, name string) (Deck, error) {
	if id <= 0 || name == "" {
		return Deck{}, fmt.Errorf("invalid id or name")
	}
	return Deck{
		ID:   id,
		Name: name,
	}, nil
}

func WithPsqlInfo(psqlInfo string) Option {
	return func(opts *dbOptions) {
		opts.psqlInfo = psqlInfo
	}
}

func ConnectToDB(opts ...Option) (*sql.DB, error) {
	defaultOptions := &dbOptions{
		psqlInfo: fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			"localhost", 5432, "devon", "root", "flashcard"),
	}

	for _, opt := range opts {
		opt(defaultOptions)
	}

	//fmt.Printf("Connecting to database with options: %v\n", defaultOptions)

	// I have removed the error handling for the sake of simplicity
	db, _ := sql.Open("postgres", defaultOptions.psqlInfo)

	// Force a connection check to ensure DB exists and is connected.
	err := db.Ping()
	if err != nil {
		fmt.Printf("Error pinging database: %v\n", err)
		return nil, err
	}

	return db, nil
}

func CreateTableFromSchema(db *sql.DB, schema TableSchema) error {
	_, err := db.Exec(schema.CreateSQL)

	if err != nil {
		return fmt.Errorf("error creating table %s: %w", schema.Name, err)
	}

	return nil
}

func CreateAllTables(db *sql.DB, tables []TableSchema) error {
	for _, table := range tables {
		err := CreateTableFromSchema(db, table)
		if err != nil {
			return err
		}
	}

	return nil
}

func DropTable(db *sql.DB, tableName string) error {
	// Create the SQL statement to drop the table. Use parameter substitution for safety.
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE;", tableName)

	// Execute the query
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to drop table %s: %v", tableName, err)
	}

	fmt.Printf("Table %s dropped successfully.\n", tableName)
	return nil
}

func DropAllTables(db *sql.DB) error {
	tables := []string{"card_deck", "cards", "decks"}

	for _, table := range tables {
		if err := DropTable(db, table); err != nil {
			return fmt.Errorf("error dropping table %s: %v", table, err)
		}
	}
	return nil
}

func InsertCards(db *sql.DB, cards []Card) error {
	for _, card := range cards {
		_, err := db.Exec("INSERT INTO cards (front, back, recency, prevdifficulty) VALUES ($1, $2, $3, $4)", card.Front, card.Back, card.Reviewed, card.Difficulty)
		if err != nil {
			return err
		}
	}
	return nil
}

func AddCardToDeck(db *sql.DB, cardID int, deckID int) error {
	// SQL statement to insert a new relation into the card_deck table
	query := `INSERT INTO deck_cards (card_id, deck_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;`

	// Execute the query with the provided cardID and deckID
	_, err := db.Exec(query, cardID, deckID)
	if err != nil {
		return fmt.Errorf("failed to add card %d to deck %d: %v", cardID, deckID, err)
	}

	return nil
}

func InsertDeck(db *sql.DB, deckName string) error {
	_, err := db.Exec("INSERT INTO decks (name) VALUES ($1)", deckName)
	if err != nil {
		return err
	}
	return nil
}

func PrintCards(db *sql.DB) error {
	rows, err := db.Query("SELECT * FROM cards")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var front, back string
		var id int
		var reviewed int64
		var difficulty int

		rows.Scan(&id, &front, &back, &reviewed, &difficulty)
		fmt.Printf("Front: %s, Back: %s, ID: %d, Reviewed: %d, Difficulty: %d\n", front, back, id, reviewed, difficulty)
	}
	return nil
}

func PrintCardsInDeck(db *sql.DB, deckID int) error {
	query := `
        SELECT cards.* FROM cards
		JOIN deck_cards ON cards.id = deck_cards.card_id
		WHERE deck_cards.deck_id = $1;`

	rows, err := db.Query(query, deckID)
	if err != nil {
		return fmt.Errorf("error querying cards: %w", err)
	}
	defer rows.Close()

	fmt.Printf("Cards in deck %d:\n", deckID)
	for rows.Next() {
		var card Card
		err := rows.Scan(&card.ID, &card.Front, &card.Back, &card.Reviewed, &card.Difficulty)
		if err != nil {
			return fmt.Errorf("error scanning card: %w", err)
		}
		fmt.Printf("- ID: %d, Front: %s, Back: %s\n", card.ID, card.Front, card.Back)
	}

	return nil
}

func DeleteCard(db *sql.DB, card Card) error {
	_, err := db.Exec("DELETE FROM cards WHERE id = $1", card.ID)
	if err != nil {
		return err
	}
	return nil
}

func DeleteCardByID(db *sql.DB, cardID int) error {
	_, err := db.Exec("DELETE FROM cards WHERE id = $1", cardID)
	if err != nil {
		return err
	}
	return nil
}

func GetRandomCard(db *sql.DB) (*Card, error) {
	// 1. Get the total number of cards
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM cards").Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("error getting card count: %v", err)
	}

	// 2. Generate a random ID within the valid range
	randomID := rand.Intn(count) + 1

	// 3. Fetch the card with the random ID
	var card Card
	err = db.QueryRow("SELECT id, front, back FROM cards WHERE id = $1", randomID).Scan(&card.ID, &card.Front, &card.Back)
	if err != nil {
		return nil, fmt.Errorf("error getting card: %v", err)
	}

	return &card, nil
}

func GetCardsFromDeck(db *sql.DB, deckID int) (*[]Card, error) {
	// 1. Fetch the cards associated with the deck
	rows, err := db.Query(`
        SELECT c.*
        FROM cards c
        JOIN deck_cards dc ON c.id = dc.card_id
        WHERE dc.deck_id = $1
    `, deckID)
	if err != nil {
		return nil, fmt.Errorf("error getting cards for deck: %v", err)
	}
	defer rows.Close()

	cards := []Card{}

	// 2. Populate the deck's Cards slice
	for rows.Next() {
		var card Card
		err := rows.Scan(&card.ID, &card.Front, &card.Back, &card.Reviewed, &card.Difficulty)
		if err != nil {
			return nil, fmt.Errorf("error scanning card: %v", err)
		}
		cards = append(cards, card)
	}

	return &cards, nil
}

func GetDecksData(db *sql.DB) (*[]Deck, error) {
	// 1. Fetch all decks
	rows, err := db.Query("SELECT * FROM decks")
	if err != nil {
		return nil, fmt.Errorf("error getting decks: %v", err)
	}
	defer rows.Close()


	decks := []Deck{}

	// 2. Populate the decks slice
	for rows.Next() {
		var deck Deck
		err := rows.Scan(&deck.ID, &deck.Name)
		if err != nil {
			return nil, fmt.Errorf("error scanning deck: %v", err)
		}
		decks = append(decks, deck)
	}

	return &decks, nil
}

/* Get all cards from given deck

SELECT c.*
FROM cards c
JOIN deck_cards dc ON c.id = dc.card_id
WHERE dc.deck_id = $1;

*/

/* Get all decks a card is in

SELECT d.*
FROM decks d
JOIN deck_cards dc ON d.id = dc.deck_id
WHERE dc.card_id = $1;
*/

/*  Insert into deck_cards
INSERT INTO deck_cards (card_id, deck_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;
*/
