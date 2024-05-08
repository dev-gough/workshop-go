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
		mock.ExpectExec("CREATE TABLE IF NOT EXISTS card_deck").WillReturnResult(sqlmock.NewResult(0, 0))

		// Call the function that executes the SQL
		tables := []TableSchema{CardsTable, DecksTable, CardDeckTable}
		CreateAllTables(db, tables) // This function needs to be implemented in your actual code

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
		CreateAllTables(db, tables) // This function needs to be implemented in your actual code

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

func TestInsertDecek(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("error creating mock database: %v", err)
		}
		defer db.Close()

		mock.ExpectExec("INSERT INTO decks").WillReturnResult(sqlmock.NewResult(0, 1))

		err = InsertDeck(db, Deck{ID: 1, Name: "Deck"})
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

		err = InsertDeck(db, Deck{ID: 1})
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

		// Close the writer and restore stdout
		w.Close()
		os.Stdout = old

		// Read the output
		var buf bytes.Buffer
		buf.ReadFrom(r)
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

// TODO: Add tests for PrintCardsFromDeck

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

		// Assuming InsertCards is another function that inserts 'card' into the database
		InsertCards(db, []Card{card})  // Make sure this is mocked if it's supposed to interact with the database

		// Call DeleteCard with the 'card' instance
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

		InsertCards(db, []Card{card})
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

		InsertCards(db, []Card{card})
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

		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO card_deck")).WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(0, 1))

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

		mock.ExpectExec("INSERT INTO card_deck").WillReturnError(fmt.Errorf("error adding card to deck"))

		err = AddCardToDeck(db, 1, 1)
		assert.Error(t, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
