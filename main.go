package main

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/TyrinH/smolURL/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type websiteRedirect struct {
	originalUrl string
	redirectUrl string
}

func main() {
	godotenv.Load()
	token := os.Getenv("TURSO_AUTH_TOKEN")
	dbUrl := os.Getenv("TURSO_DATABASE_URL")
	url := fmt.Sprintf("%s?authToken=%s", dbUrl, token)

  db, err := sql.Open("libsql", url)
  if err != nil {
    fmt.Fprintf(os.Stderr, "failed to open db %s: %s", url, err)
    os.Exit(1)
  }
	defer db.Close()
	createErr := createWebsiteUrlTable(db)
	if createErr != nil {
		log.Println(createErr)
	}

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.POST("/website/:url", func(c *gin.Context) {
		websiteRedirect := websiteRedirect{}
		websiteRedirect.originalUrl = c.Param("url")
		data := []byte(websiteRedirect.originalUrl)
		encodedString := base64.StdEncoding.EncodeToString(data)
		encodedUrlRedirect := encodedString[:7]
		websiteRedirect.redirectUrl = fmt.Sprintf("localhost:8080/%s", encodedUrlRedirect)
		createdRedirect, err := createWebsiteRedirect(db, websiteRedirect)
		if err != nil {
			log.Println(err)
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "received",
			"redirectUrl": createdRedirect.Redirecturl,
		})
	})
	r.Run()

}

type User struct {
	ID   int
	Name string
}

//go:embed sql/schema.sql
var ddl string

func createWebsiteUrlTable(db * sql.DB) error{
	ctx := context.Background()

		// create tables
		if _, err := db.ExecContext(ctx, ddl); err != nil {
			return err
		}
		return nil
}

func createWebsiteRedirect(db * sql.DB, redirect websiteRedirect)( database.WebsiteRedirect, error) {
	ctx := context.Background()
	queries := database.New(db)

	insertedRedirect, err := queries.CreateWebsiteRedirect(ctx, database.CreateWebsiteRedirectParams{
		Originalurl: redirect.originalUrl,
		Redirecturl: redirect.redirectUrl,
	})
	return insertedRedirect, err

}