package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/lib/pq"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func CreateDatabaseContainer(instanceKey string) (*sqlstore.Container, error) {
	dbDriver := os.Getenv("DB_DRIVER")
	dbURL := os.Getenv("DB_URL")

	// Create a new database for this instance
	dbName := "whatsapp_" + instanceKey
	// The DB_URL should point to a maintenance database (e.g., "postgres")
	db, err := sql.Open(dbDriver, dbURL)
	if err != nil {
		return nil, fmt.Errorf("error opening maintenance database: %w", err)
	}
	defer db.Close()

	// Using fmt.Sprintf because CREATE DATABASE doesn't support parameterized queries for the db name.
	// instanceKey is a hex string, so it's safe from SQL injection.
	_, err = db.Exec(fmt.Sprintf(`CREATE DATABASE "%s"`, dbName))
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "42P04" { // 42P04 is duplicate_database
			log.Printf("Database %s already exists, proceeding.", dbName)
		} else {
			return nil, fmt.Errorf("error creating database %s: %w", dbName, err)
		}
	} else {
		log.Printf("Successfully created database %s", dbName)
	}

	// Construct the new DB URL for this instance
	parsedURL, err := url.Parse(dbURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing DB URL: %w", err)
	}
	parsedURL.Path = "/" + dbName
	instanceDbURL := parsedURL.String()

	// Setup database for this instance
	dbLog := waLog.Stdout(fmt.Sprintf("Database-%s", instanceKey), "DEBUG", true)

	container, err := sqlstore.New(context.Background(), dbDriver, instanceDbURL, dbLog)
	if err != nil {
		return nil, fmt.Errorf("error creating database container for instance %s: %w", instanceKey, err)
	}
	return container, nil
}

func DropDatabase(instanceKey string) {
	dbDriver := os.Getenv("DB_DRIVER")
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open(dbDriver, dbURL)
	if err != nil {
		log.Printf("Warning: Error opening maintenance database to drop instance db: %v", err)
	} else {
		defer db.Close()
		dbName := "whatsapp_" + instanceKey
		// Using fmt.Sprintf because DROP DATABASE doesn't support parameterized queries for the db name.
		_, err = db.Exec(fmt.Sprintf(`DROP DATABASE "%s"`, dbName))
		if err != nil {
			log.Printf("Warning: Error dropping database %s: %v", dbName, err)
		} else {
			log.Printf("Successfully dropped database %s", dbName)
		}
	}
}
