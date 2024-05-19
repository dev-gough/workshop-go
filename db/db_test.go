package db

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestCreateCard(t *testing.T) {
	t.Run("Successful creation", func(t *testing.T) {
		card, err := CreateCard(1, "front", "back", 1, 5)
		assert.Nil(t, err)
		assert.Equal(t, 1, card.ID)
		assert.Equal(t, "front", card.Front)
		assert.Equal(t, "back", card.Back)
		assert.Equal(t, int64(1), card.Reviewed)
		assert.Equal(t, 5, card.Difficulty)
	})

	t.Run("Erroneous creation", func(t *testing.T) {
		card, err := CreateCard(1, "", "back", 1, 5)
		assert.NotNil(t, err)
		assert.EqualError(t, err, "invalid card data")
		assert.Equal(t, Card{}, card)
	})
}

func TestCreateDeck(t *testing.T) {
	t.Run("Valid inputs", func(t *testing.T) {
		expectedDeck := Deck{
			ID:   1,
			Name: "Test Deck",
		}
		actualDeck, err := CreateDeck(1, "Test Deck")
		assert.NoError(t, err)
		assert.Equal(t, expectedDeck, actualDeck)
	})

	t.Run("Invalid id", func(t *testing.T) {
		_, err := CreateDeck(0, "Test Deck")
		assert.Error(t, err)
	})

	t.Run("Invalid name", func(t *testing.T) {
		_, err := CreateDeck(1, "")
		assert.Error(t, err)
	})
}

func TestDropAllTables(t *testing.T) {
	tests := []struct {
		name    string
		tables  []string
		dropErr error
		wantErr bool
	}{
		{
			name:    "Success",
			tables:  []string{"card_deck", "cards", "decks"},
			dropErr: nil,
			wantErr: false,
		},
		{
			name:    "Error dropping table",
			tables:  []string{"card_deck", "cards", "decks"},
			dropErr: fmt.Errorf("error dropping table"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("error creating mock database: %v", err)
			}
			defer db.Close()

			// Loop through each table and set expectations
			for _, table := range tt.tables {
				// For a successful drop, set a result. For error cases, set the error.
				if tt.dropErr == nil {
					mock.ExpectExec("DROP TABLE IF EXISTS " + table + " CASCADE;").WillReturnResult(sqlmock.NewResult(0, 1))
				}
			}

			err = DropAllTables(db) // Adjusted to assume DropAllTables accepts table names
			if (err != nil) != tt.wantErr {
				t.Errorf("DropAllTables() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check if all expected queries were executed
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}

}

func TestConnectToDB(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, err := ConnectToDB()
		if err != nil {
			t.Fatalf("error connecting to database: %v", err)
		}
		defer db.Close()

		errPing := db.Ping()

		assert.NoError(t, err)
		assert.NoError(t, errPing)
		assert.NotNil(t, db)
	})

	t.Run("Success using options", func(t *testing.T) {
		psqlInfo := "host=localhost port=5432 user=devon password=root dbname=flashcard sslmode=disable"
		db, err := ConnectToDB(WithPsqlInfo(psqlInfo))

		if err != nil {
			t.Fatalf("error connecting to database: %v", err)
		}
		defer db.Close()

		errPing := db.Ping()
		assert.NoError(t, err)
		assert.NoError(t, errPing)
		assert.NotNil(t, db)
	})

	t.Run("Error using options", func(t *testing.T) {
		psqlInfo := "host=localhost port=5432 user=devon password=wrongpassword dbname=flashcard sslmode=disablee"
		db, err := ConnectToDB(WithPsqlInfo(psqlInfo))

		assert.Error(t, err)
		assert.Nil(t, db)
	})
}

func TestCreateTableFromSchema(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("error creating mock database: %v", err)
		}
		defer db.Close()

		mock.ExpectExec("CREATE TABLE IF NOT EXISTS cards").WillReturnResult(sqlmock.NewResult(0, 1))

		err = CreateTableFromSchema(db, CardsTable)
		assert.NoError(t, err)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %v", err)
		}
	})

	t.Run("Error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("error creating mock database: %v", err)
		}
		defer db.Close()

		mock.ExpectExec("CREATE TABLE IF NOT EXISTS cards").WillReturnError(fmt.Errorf("error creating table"))

		err = CreateTableFromSchema(db, CardsTable)
		assert.Error(t, err)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %v", err)
		}
	})
}

func TestCreateAllTables(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
		}
		defer db.Close()

		// Set expectations for SQL commands
		mock.ExpectExec("CREATE TABLE IF NOT EXISTS cards").WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec("CREATE TABLE IF NOT EXISTS decks").WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec("CREATE TABLE IF NOT EXISTS deck_cards").WillReturnResult(sqlmock.NewResult(0, 0))

		// Call the function that executes the SQL
		tables := []TableSchema{CardsTable, DecksTable, DeckCardsTable}
		err = CreateAllTables(db, tables)
		assert.NoError(t, err)
		// Verify that all expectations were met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err, "there were unfulfilled expectations")
	})
	t.Run("Error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
		}
		defer db.Close()

		// Set expectations for SQL commands
		mock.ExpectExec("CREATE IFNOT EXISTS ").WillReturnError(fmt.Errorf("error creating table"))

		// Call the function that executes the SQL
		tables := []TableSchema{{Name: "invalid", CreateSQL: "CREATE IFNOT EXISTS "}}
		err = CreateAllTables(db, tables)
		assert.Error(t, err)

		// Verify that all expectations were met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err, "there were unfulfilled expectations")
	})
}

func TestInsertCards(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("error creating mock database: %v", err)
		}
		defer db.Close()

		mock.ExpectExec("INSERT INTO cards").WillReturnResult(sqlmock.NewResult(0, 1))

		err = InsertCards(db, []Card{{ID: 1, Front: "Front", Back: "Back", Reviewed: 1, Difficulty: 5}})
		assert.NoError(t, err)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %v", err)
		}
	})
	t.Run("Error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("error creating mock database: %v", err)
		}
		defer db.Close()

		mock.ExpectExec("INSERT INTO cards").WillReturnError(fmt.Errorf("error inserting card"))

		err = InsertCards(db, []Card{{ID: 1, Front: "Front", Back: "Back", Reviewed: 1}})
		assert.Error(t, err)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %v", err)
		}
	})
}

func TestInsertDeck(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("error creating mock database: %v", err)
		}
		defer db.Close()

		mock.ExpectExec("INSERT INTO decks").WillReturnResult(sqlmock.NewResult(0, 1))

		err = InsertDeck(db, "Deck1")
		assert.NoError(t, err)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %v", err)
		}
	})
	t.Run("Error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("error creating mock database: %v", err)
		}
		defer db.Close()

		mock.ExpectExec("INSERT INTO decks").WillReturnError(fmt.Errorf("error inserting deck"))

		_ = InsertDeck(db, "Deck1")
		err = InsertDeck(db, "Deck1")
		assert.Error(t, err)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %v", err)
		}
	})
}

func TestPrintCards(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("error creating mock database: %v", err)
		}
		defer db.Close()

		// Set expectations for the SQL query
		rows := sqlmock.NewRows([]string{"id", "front", "back", "reviewed", "difficulty"}).
			AddRow(1, "Front of card 1", "Back of card 1", 1234567890, 5).
			AddRow(2, "Front of card 2", "Back of card 2", 1234567891, 6)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM cards")).WillReturnRows(rows)

		old := os.Stdout // keep backup of the real stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Call the function
		err = PrintCards(db)
		assert.NoError(t, err)

		// Close the writer and restore stdout
		w.Close()
		os.Stdout = old

		// Read the output
		var buf bytes.Buffer
		_, err = buf.ReadFrom(r)
		output := buf.String()

		// Check for errors in database operations
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())

		// Check the output text
		expectedOutput := "Front: Front of card 1, Back: Back of card 1, ID: 1, Reviewed: 1234567890, Difficulty: 5\n" +
			"Front: Front of card 2, Back: Back of card 2, ID: 2, Reviewed: 1234567891, Difficulty: 6\n"
		assert.Equal(t, expectedOutput, output)
	})
	t.Run("Error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("error creating mock database: %v", err)
		}
		defer db.Close()

		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM cards")).WillReturnError(fmt.Errorf("error selecting cards"))

		err = PrintCards(db)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPrintCardsInDeck(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("error creating mock database: %v", err)
		}
		defer db.Close()

		// Set up expected query and results for deck ID 1
		expectedDeckID := 1
		rows := sqlmock.NewRows([]string{"id", "front", "back", "reviewed", "difficulty"}).
			AddRow(3, "Front 3", "Back 3", 9876543210, 3).
			AddRow(5, "Front 5", "Back 5", 9876543211, 4)

		// Expect a specific query with the deck ID
		mock.ExpectQuery(regexp.QuoteMeta(`
            SELECT cards.* FROM cards
            JOIN deck_cards ON cards.id = deck_cards.card_id
            WHERE deck_cards.deck_id = $1;
        `)).
			WithArgs(expectedDeckID).
			WillReturnRows(rows)

		// Capture output to test it later
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Call the function under test
		err = PrintCardsInDeck(db, expectedDeckID)
		assert.NoError(t, err)

		// Clean up and restore stdout
		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, err = buf.ReadFrom(r)
		assert.NoError(t, err)
		output := buf.String()

		// Verify database interactions and output
		assert.NoError(t, mock.ExpectationsWereMet())

		expectedOutput := fmt.Sprintf("Cards in deck %d:\n- ID: 3, Front: Front 3, Back: Back 3\n- ID: 5, Front: Front 5, Back: Back 5\n", expectedDeckID)
		assert.Equal(t, expectedOutput, output)
	})

	t.Run("Error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("error creating mock database: %v", err)
		}
		defer db.Close()

		// Expect the query and return an error
		mock.ExpectQuery(regexp.QuoteMeta(`
            SELECT cards.* FROM cards
            JOIN deck_cards ON cards.id = deck_cards.card_id
            WHERE deck_cards.deck_id = $1;
        `)).
			WithArgs(1). // Assuming deck ID 1 for the error case as well
			WillReturnError(fmt.Errorf("query error"))

		// Call the function and assert an error is returned
		err = PrintCardsInDeck(db, 1)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
// working
func TestGetRandomCard(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("error creating mock database: %v", err)
		}
		defer db.Close()

		// 1. Mock the count query
		expectedCount := 10
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM cards").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedCount))

		// 2. Mock the card retrieval query
		expectedCard := Card{ID: 5, Front: "Front 5", Back: "Back 5"} // Example card
		mock.ExpectQuery("SELECT id, front, back FROM cards WHERE id = \\$1").
			WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id", "front", "back"}).
				AddRow(expectedCard.ID, expectedCard.Front, expectedCard.Back))

		// Call the function under test
		card, err := GetRandomCard(db)

		// Verify results and database interactions
		assert.NoError(t, err)
		assert.NotNil(t, card)
		assert.Equal(t, expectedCard, *card)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ErrorGettingCount", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("error creating mock database: %v", err)
		}
		defer db.Close()

		// Mock an error when getting the count
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM cards").
			WillReturnError(fmt.Errorf("count query error"))

		// Call the function and expect an error
		card, err := GetRandomCard(db)

		assert.Error(t, err)
		assert.Nil(t, card)
		assert.EqualError(t, err, "error getting card count: count query error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ErrorGettingCard", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("error creating mock database: %v", err)
		}
		defer db.Close()

		// Mock success for count, error for card retrieval
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM cards").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
		mock.ExpectQuery("SELECT id, front, back FROM cards WHERE id = \\$1").
			WithArgs(sqlmock.AnyArg()). // Assuming the random ID will be 1
			WillReturnError(fmt.Errorf("card retrieval error"))

		// Call the function and expect an error
		card, err := GetRandomCard(db)

		assert.Error(t, err)
		assert.Nil(t, card)
		assert.EqualError(t, err, "error getting card: card retrieval error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetCardsFromDeck(t *testing.T) {
    t.Run("Success", func(t *testing.T) {
        db, mock, err := sqlmock.New()
        if err != nil {
            t.Fatalf("error creating mock database: %v", err)
        }
        defer db.Close()

        // 1. Mock successful query with expected deck ID and cards
        deckID := 123 // Example deck ID
        expectedCards := []Card{
            {ID: 1, Front: "Front 1", Back: "Back 1", Reviewed: 1234567890, Difficulty: 5},
            {ID: 2, Front: "Front 2", Back: "Back 2", Reviewed: 9876543210, Difficulty: 4},
        }

        rows := sqlmock.NewRows([]string{"id", "front", "back", "reviewed", "difficulty"})
        for _, card := range expectedCards {
            rows.AddRow(card.ID, card.Front, card.Back, card.Reviewed, card.Difficulty)
        }
        mock.ExpectQuery(regexp.QuoteMeta(`
            SELECT c.*
            FROM cards c
            JOIN deck_cards dc ON c.id = dc.card_id
            WHERE dc.deck_id = $1
        `)).WithArgs(deckID).WillReturnRows(rows)

        // 2. Call the function
        cards, err := GetCardsFromDeck(db, deckID)

        // 3. Assert expected results
        assert.NoError(t, err)
        assert.NotNil(t, cards)
        assert.Equal(t, expectedCards, *cards)
        assert.NoError(t, mock.ExpectationsWereMet())
    })

    t.Run("QueryError", func(t *testing.T) {
        db, mock, err := sqlmock.New()
        if err != nil {
            t.Fatalf("error creating mock database: %v", err)
        }
        defer db.Close()

        // Mock an error when fetching cards
        mock.ExpectQuery(regexp.QuoteMeta(`
            SELECT c.*
            FROM cards c
            JOIN deck_cards dc ON c.id = dc.card_id
            WHERE dc.deck_id = $1
        `)).WillReturnError(fmt.Errorf("query error"))

        // Call the function and expect an error
        cards, err := GetCardsFromDeck(db, 123)
        assert.Error(t, err)
        assert.Nil(t, cards)
        assert.EqualError(t, err, "error getting cards for deck: query error")
        assert.NoError(t, mock.ExpectationsWereMet())
    })

    t.Run("ScanError", func(t *testing.T) {
        db, mock, err := sqlmock.New()
        if err != nil {
            t.Fatalf("error creating mock database: %v", err)
        }
        defer db.Close()

        // Mock invalid data returned from the database that would fail Scan()
        mock.ExpectQuery(regexp.QuoteMeta(`
            SELECT c.*
            FROM cards c
            JOIN deck_cards dc ON c.id = dc.card_id
            WHERE dc.deck_id = $1
        `)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("invalid"))

        // Call the function and expect an error
        cards, err := GetCardsFromDeck(db, 123)
        assert.Error(t, err)
        assert.Nil(t, cards)
        assert.Contains(t, err.Error(), "error scanning card:")
        assert.NoError(t, mock.ExpectationsWereMet())
    })
}

func TestGetDecksData(t *testing.T) {
    t.Run("Success", func(t *testing.T) {
        db, mock, err := sqlmock.New()
        if err != nil {
            t.Fatalf("error creating mock database: %v", err)
        }
        defer db.Close()

        // 1. Mock successful query with expected deck data
        expectedDecks := []Deck{
            {ID: 1, Name: "Deck 1"},
            {ID: 2, Name: "Deck 2"},
            {ID: 3, Name: "Deck 3"}, // Adding more decks for a thorough test
        }

        rows := sqlmock.NewRows([]string{"id", "name"})
        for _, deck := range expectedDecks {
            rows.AddRow(deck.ID, deck.Name)
        }
        mock.ExpectQuery("SELECT \\* FROM decks").WillReturnRows(rows)

        // 2. Call the function
        decks, err := GetDecksData(db)

        // 3. Assert expected results
        assert.NoError(t, err)
        assert.NotNil(t, decks)
        assert.Equal(t, expectedDecks, *decks)
        assert.NoError(t, mock.ExpectationsWereMet())
    })

    t.Run("QueryError", func(t *testing.T) {
        db, mock, err := sqlmock.New()
        if err != nil {
            t.Fatalf("error creating mock database: %v", err)
        }
        defer db.Close()

        // Mock an error when fetching decks
        mock.ExpectQuery("SELECT \\* FROM decks").WillReturnError(fmt.Errorf("query error"))

        // Call the function and expect an error
        decks, err := GetDecksData(db)
        assert.Error(t, err)
        assert.Nil(t, decks)
        assert.EqualError(t, err, "error getting decks: query error")
        assert.NoError(t, mock.ExpectationsWereMet())
    })

    t.Run("ScanError", func(t *testing.T) {
        db, mock, err := sqlmock.New()
        if err != nil {
            t.Fatalf("error creating mock database: %v", err)
        }
        defer db.Close()

        // Mock invalid data returned from the database to trigger a Scan() error
        mock.ExpectQuery("SELECT \\* FROM decks").
            WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
                AddRow("invalid", 123)) // Inconsistent data types

        // Call the function and expect an error
        decks, err := GetDecksData(db)
        assert.Error(t, err)
        assert.Nil(t, decks)
        assert.Contains(t, err.Error(), "error scanning deck:") // Check for partial error message
        assert.NoError(t, mock.ExpectationsWereMet())
    })
}

func TestDeleteCard(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("error creating mock database: %v", err)
		}
		defer db.Close()

		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM cards WHERE id = $1")).WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))

		// Create a Card instance
		card := Card{ID: 1, Front: "Front", Back: "Back", Reviewed: 1, Difficulty: 5}

		InsertCards(db, []Card{card}) // trunk-ignore(golangci-lint/errcheck)

		err = DeleteCard(db, card)
		assert.NoError(t, err)

		// Verify that all expectations set on the mock database were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("Error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("error creating mock database: %v", err)
		}
		defer db.Close()

		mock.ExpectExec("DELETE FROM cards").WillReturnError(fmt.Errorf("error deleting card"))
		card := Card{ID: 99, Front: "Front", Back: "Back", Reviewed: 1, Difficulty: 5}
		card2 := Card{Front: "Front", Back: "Back", Reviewed: 1, Difficulty: 5}

		InsertCards(db, []Card{card}) // trunk-ignore(golangci-lint/errcheck)
		err = DeleteCard(db, card2)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDeleteCardByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("error creating mock database: %v", err)
		}
		defer db.Close()

		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM cards WHERE id = $1")).WithArgs(99).WillReturnResult(sqlmock.NewResult(0, 1))
		card := Card{ID: 99, Front: "Front", Back: "Back", Reviewed: 1, Difficulty: 5}

		InsertCards(db, []Card{card}) // trunk-ignore(golangci-lint/errcheck)
		err = DeleteCardByID(db, 99)
		assert.NoError(t, err)

		// Verify that all expectations set on the mock database were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("Error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("error creating mock database: %v", err)
		}
		defer db.Close()

		mock.ExpectExec("DELETE FROM cards").WillReturnError(fmt.Errorf("error deleting card"))

		err = DeleteCardByID(db, 99)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAddCardToDeck(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("error creating mock database: %v", err)
		}
		defer db.Close()

		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO deck_cards")).WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(0, 1))

		card := Card{ID: 99, Front: "Front", Back: "Back", Reviewed: 1, Difficulty: 5}
		InsertCards(db, []Card{card})

		err = AddCardToDeck(db, 1, 1)
		assert.NoError(t, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("Error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("error creating mock database: %v", err)
		}
		defer db.Close()

		mock.ExpectExec("INSERT INTO deck_cards").WillReturnError(fmt.Errorf("error adding card to deck"))

		err = AddCardToDeck(db, 1, 1)
		assert.Error(t, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
