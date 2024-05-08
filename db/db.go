package db

import (
	"database/sql"
	"errors"
	"fmt"

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

var CardDeckTable = TableSchema{
	Name: "card_deck",
	CreateSQL: `CREATE TABLE IF NOT EXISTS card_deck (
    card_id INT,
    deck_id INT,
    PRIMARY KEY (card_id, deck_id),
    FOREIGN KEY (card_id) REFERENCES cards (id) ON DELETE CASCADE,
    FOREIGN KEY (deck_id) REFERENCES decks (id) ON DELETE CASCADE
);`,
}

var CurrentTables = []TableSchema{CardsTable, DecksTable, CardDeckTable}

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

	fmt.Printf("Connecting to database with options: %v\n", defaultOptions)

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
		}}
	return nil
}

func InsertCards(db *sql.DB, cards []Card) error {
	for _, card := range cards {
		_, err := db.Exec("INSERT INTO cards (front, back, recency, prevdifficulty) VALUES ($1, $2, $3, $4)", card.Front, card.Back, card.Reviewed, card.Difficulty)
		if err != nil {
			return err
		}}
	return nil
}

func AddCardToDeck(db *sql.DB, cardID, deckID int) error {
	// SQL statement to insert a new relation into the card_deck table
	query := `INSERT INTO card_deck (card_id, deck_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;`

	// Execute the query with the provided cardID and deckID
	_, err := db.Exec(query, cardID, deckID)
	if err != nil {
		return fmt.Errorf("failed to add card %d to deck %d: %v", cardID, deckID, err)
	}

	return nil
}

func InsertDeck(db *sql.DB, deck Deck) error {
	_, err := db.Exec("INSERT INTO decks (name) VALUES ($1)", deck.Name)
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

// NEW (req. relation func)
func PrintCardsFromDeck(db *sql.DB, deckID int) error {
	// Query to select all cards from the provided deck_id
	query := `
    SELECT c.id, c.front, c.back, c.recency, c.prevdifficulty
    FROM cards AS c
    JOIN card_deck AS cd ON c.id = cd.card_id
    WHERE cd.deck_id = $1;
    `


	// Prepare the query
	stmt, err := db.Prepare(query)
	if err != nil {
		return fmt.Errorf("error preparing query: %v", err)
	}
	defer stmt.Close()

	// Execute the query with the provided deck_id
	rows, err := stmt.Query(deckID)
	if err != nil {
		return fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	fmt.Printf("Rows %s\n", rows.Next())
	// Loop through the returned rows and print the card information
	for rows.Next() {
		fmt.Printf("Got here")
		var cardID, recency, prevdifficulty int
		var front, back string
		if err := rows.Scan(&cardID, &front, &back, &recency, &prevdifficulty); err != nil {
			return fmt.Errorf("error scanning row: %v", err)
		}
		fmt.Printf("Card ID: %d\n", cardID)
		fmt.Printf("Front: %s\n", front)
		fmt.Printf("Back: %s\n", back)
		fmt.Printf("Recency: %d\n", recency)
		fmt.Printf("Previous Difficulty: %d\n", prevdifficulty)
		fmt.Println("-------------------")
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating through rows: %v", err)
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
