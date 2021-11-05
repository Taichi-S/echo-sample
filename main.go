package main // Add your code to a main package so you can execute it independently.

import (
	"database/sql" // In most cases clients will use the database/sql package instead of using pq package directly.
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq" //Import the Postgres Driver．With the driver imported, you’ll start writing code to access the database.
)

type Album struct { // use this to hold row data returned from the query
	ID     int64
	Title  string
	Artist string
	Price  float32
}

var db *sql.DB // productionではグローバル変数は非推奨

func main() {
	// connStr := "host=localhost user=ichi dbname=ichi sslmode=disable"
	var err error
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	// db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err) // In production code, you’ll want to handle errors in a more graceful way.
	}

	pingErr := db.Ping() // You’re using Ping here to confirm that the database/sql package can connect when it needs to.
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected!")

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.GET("/albums", getAllAlbums)
	e.POST("/albums", createAlbum)
	e.GET("/album/:id", getAlbum)
	e.PUT("/albums/:id", updateAlbum)
	e.DELETE("/albums/:id", deleteAlbum)

	e.Logger.Fatal(e.Start(":" + os.Getenv("PORT")))
}

func getAllAlbums(c echo.Context) error {
	var albums []Album

	rows, err := db.Query("SELECT * FROM album")
	if err != nil {
		return fmt.Errorf("getAllAlbums: %v", err)
	}

	defer rows.Close()
	for rows.Next() {
		var alb Album
		if err := rows.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil { //Scan takes a list of pointers to Go values
			return fmt.Errorf("getAllAlbums: %v", err)
		}
		albums = append(albums, alb)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("getAllAlbums: %v", err)
	}

	return c.JSON(http.StatusOK, albums)
}

// albumByID queries for the album with the specified ID.
func getAlbum(c echo.Context) error {
	// An album to hold data from the returned row.
	var alb Album
	id, _ := strconv.Atoi(c.Param("id"))

	row := db.QueryRow("SELECT * FROM album WHERE id = $1", id) //QueryRow executes a query that is expected to return at most one row. QueryRow always returns a non-nil value.Errors are deferred until Row's Scan method is called.
	if err := row.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
		if err == sql.ErrNoRows { //sql.ErrNoRows indicates that the query returned no rows
			return fmt.Errorf("albumsById %d: no such album", id)
		}
		return fmt.Errorf("albumsById %d: %v", id, err)
	}

	return c.JSON(http.StatusOK, alb)
}

func createAlbum(c echo.Context) error {
	var lastInsertId int64
	a := new(Album)
	if err := c.Bind(a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := db.QueryRow("INSERT INTO album (title, artist, price) VALUES ($1, $2, $3) RETURNING id", a.Title, a.Artist, a.Price).Scan(&lastInsertId); err != nil {
		return fmt.Errorf("createAlbum: %v", err)
	}
	a.ID = lastInsertId

	return c.JSON(http.StatusCreated, a)
}

func updateAlbum(c echo.Context) error {
	a := new(Album)
	if err := c.Bind(a); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))

	if _, err := db.Exec("UPDATE album SET title=$1, artist=$2, price=$3 WHERE id=$4", a.Title, a.Artist, a.Price, id); err != nil {
		return fmt.Errorf("updateAlbum: %v", err)
	}
	a.ID = int64(id)

	return c.JSON(http.StatusOK, a)
}

func deleteAlbum(c echo.Context) error {
	a := new(Album)
	if err := c.Bind(a); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))

	if _, err := db.Exec("DELETE FROM album WHERE id=$1", id); err != nil {
		return fmt.Errorf("deleteAlbum: %v", err)
	}

	return c.NoContent(http.StatusNoContent)
}
