package database

import (
	"database/sql"
	"log"

	"hyw-webpics/config"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Connect() {
	var err error
	DB, err = sql.Open("sqlite", config.AppConfig.DatabasePath)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	createTables()
	log.Println("Database connected successfully")
}

func createTables() {
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	categoriesTable := `
	CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		slug TEXT NOT NULL UNIQUE
	);`

	imagesTable := `
	CREATE TABLE IF NOT EXISTS images (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		filename TEXT NOT NULL UNIQUE,
		original_name TEXT NOT NULL,
		uploader_id INTEGER NOT NULL,
		category_id INTEGER,
		status TEXT DEFAULT 'pending',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		approved_at DATETIME,
		FOREIGN KEY (uploader_id) REFERENCES users(id),
		FOREIGN KEY (category_id) REFERENCES categories(id)
	);`

	if _, err := DB.Exec(usersTable); err != nil {
		log.Fatal("Failed to create users table:", err)
	}

	if _, err := DB.Exec(categoriesTable); err != nil {
		log.Fatal("Failed to create categories table:", err)
	}

	if _, err := DB.Exec(imagesTable); err != nil {
		log.Fatal("Failed to create images table:", err)
	}

	// Migration: Check if category_id exists in images table
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM pragma_table_info('images') WHERE name='category_id'").Scan(&count)
	if err == nil && count == 0 {
		log.Println("Migrating: Adding category_id column to images table...")
		_, err = DB.Exec("ALTER TABLE images ADD COLUMN category_id INTEGER REFERENCES categories(id)")
		if err != nil {
			log.Printf("Warning: Failed to add category_id column: %v", err)
		}
	}

	seedCategories()
}

func seedCategories() {
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count)
	if count == 0 {
		categories := []struct {
			Name string
			Slug string
		}{
			{"全部", "all"}, // Virtual category usually, but can be useful
			{"梗图", "meme"},
			{"表情包", "reaction"},
			{"可爱", "cute"},
			{"怪图", "cursed"},
			{"二次元", "anime"},
		}

		for _, c := range categories {
			if c.Slug == "all" {
				continue
			}
			DB.Exec("INSERT INTO categories (name, slug) VALUES (?, ?)", c.Name, c.Slug)
		}
		log.Println("Seeded default categories")
	}
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
