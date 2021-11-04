package main // Add your code to a main package so you can execute it independently.

import (
	"database/sql" // In most cases clients will use the database/sql package instead of using pq package directly.
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq" //Import the Postgres Driver．With the driver imported, you’ll start writing code to access the database.
)

type Album struct { // use this to hold row data returned from the query
	ID     int64
	Title  string
	Artist string
	Price  float32
}

var db *sql.DB

func main() {
	// connStr := "host=localhost user=ichi dbname=ichi sslmode=disable"
	var err error
	// db, err = sql.Open("postgres", connStr)
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	fmt.Println(os.Getenv("DATABASE_URL"))
	fmt.Println(db)
	if err != nil {
		log.Fatal(err) // In production code, you’ll want to handle errors in a more graceful way.
	}

	pingErr := db.Ping() // You’re using Ping here to confirm that the database/sql package can connect when it needs to.
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected!")

	albums, err := albumsByArtist("John Coltrane")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Albums found: %v\n", albums)

	// Hard-code ID 2 here to test the query.
	alb, err := albumByID(2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Album found: %v\n", alb)

	albID, err := addAlbum(Album{
		Title:  "The Modern Sound of Betty Carter",
		Artist: "Betty Carter",
		Price:  49.99,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ID of added album: %v\n", albID)
}

// albumsByArtist queries for albums that have the specified artist name.
func albumsByArtist(name string) ([]Album, error) {
	// An albums slice to hold data from returned rows.
	var albums []Album

	rows, err := db.Query("SELECT * FROM album WHERE artist = $1", name) // By separating the SQL statement from parameter values (rather than concatenating them with, say, fmt.Sprintf), you enable the database/sql package to send the values separate from the SQL text, removing any SQL injection risk.すべての検索クエリにはなるべくデータベースが提供するパラメータ化検索インターフェースを使用する。
	if err != nil {
		return nil, fmt.Errorf("albumsByArtist %q: %v", name, err)
	}
	defer rows.Close()
	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var alb Album
		if err := rows.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil { //Scan takes a list of pointers to Go values
			return nil, fmt.Errorf("albumsByArtist %q: %v", name, err)
		}
		albums = append(albums, alb)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("albumsByArtist %q: %v", name, err)
	}
	return albums, nil
}

// albumByID queries for the album with the specified ID.
func albumByID(id int64) (Album, error) {
	// An album to hold data from the returned row.
	var alb Album

	row := db.QueryRow("SELECT * FROM album WHERE id = $1", id) //QueryRow executes a query that is expected to return at most one row. QueryRow always returns a non-nil value.Errors are deferred until Row's Scan method is called.
	if err := row.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
		if err == sql.ErrNoRows { //sql.ErrNoRows indicates that the query returned no rows
			return alb, fmt.Errorf("albumsById %d: no such album", id)
		}
		return alb, fmt.Errorf("albumsById %d: %v", id, err)
	}
	return alb, nil
}

// addAlbum adds the specified album to the database,
// returning the album ID of the new entry
func addAlbum(alb Album) (int64, error) {
	var lastInsertId int64
	if err := db.QueryRow("INSERT INTO album (title, artist, price) VALUES ($1, $2, $3) RETURNING id", alb.Title, alb.Artist, alb.Price).Scan(&lastInsertId); err != nil {
		return 0, fmt.Errorf("addAlbum: %v", err)
	}
	return lastInsertId, nil
}
