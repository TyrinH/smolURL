package main

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/TyrinH/smolURL/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type websiteRedirect struct {
	ReceivedUrl string `json:"website"`
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

	r.POST("/website", func(c *gin.Context) {
		websiteRedirect := websiteRedirect{}
		c.ShouldBind(&websiteRedirect)
		if err != nil {
			log.Println(err)
		}
		if !strings.Contains(websiteRedirect.ReceivedUrl, "http://") && !strings.Contains(websiteRedirect.ReceivedUrl, "https://") {
			websiteRedirect.originalUrl = fmt.Sprintf("http://%s", websiteRedirect.ReceivedUrl)
		} else {
			websiteRedirect.originalUrl = websiteRedirect.ReceivedUrl
		}
		data := []byte(websiteRedirect.originalUrl)
		encodedString := base64.StdEncoding.EncodeToString(data)
		encodedUrlRedirect := strings.ReplaceAll(encodedString[len(encodedString)-9:], "=", "")
		websiteRedirect.redirectUrl = encodedUrlRedirect
		createdRedirect, createWebsiteErr := createWebsiteRedirect(db, websiteRedirect)
		if createWebsiteErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "This redirect already exist.",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":     "received",
			"redirectUrl": createdRedirect.Redirecturl,
		})
	})
	r.GET("/:redirect", func(c *gin.Context) {
		redirect, err := fetchWebsiteUrl(db, c.Param("redirect"))
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
			   "message": "This website does not exist.",
		   })
		   return
	   }
		if err != nil {
			 c.JSON(http.StatusInternalServerError, gin.H{
				"message": err,
			})
			return
		}
		c.Redirect(http.StatusMovedPermanently, redirect)
	})
	r.Run()

}

type User struct {
	ID   int
	Name string
}

//go:embed sql/schema.sql
var ddl string

func createWebsiteUrlTable(db *sql.DB) error {
	ctx := context.Background()

	// create tables
	if _, err := db.ExecContext(ctx, ddl); err != nil {
		return err
	}
	return nil
}

func createWebsiteRedirect(db *sql.DB, redirect websiteRedirect) (database.WebsiteRedirect, error) {
	ctx := context.Background()
	queries := database.New(db)

	foundRedirect, _ := queries.CheckIfWebsiteRedirectExists(ctx, database.CheckIfWebsiteRedirectExistsParams{
		Originalurl: redirect.originalUrl,
		Redirecturl: redirect.redirectUrl,
	})
	if foundRedirect.ID != 0 {
		err := errors.New("duplicate redirect attempted to be added")
		return database.WebsiteRedirect{}, err
	}

	insertedRedirect, err := queries.CreateWebsiteRedirect(ctx, database.CreateWebsiteRedirectParams{
		Originalurl: redirect.originalUrl,
		Redirecturl: redirect.redirectUrl,
	})
	return insertedRedirect, err

}

func fetchWebsiteUrl(db *sql.DB, redirect string) (string, error) {
	ctx := context.Background()

	queries := database.New(db)

	fetchedUrl, err := queries.GetWebsiteRedirectByRedirectUrl(ctx, redirect)
	if err != nil {
		log.Println(err)
	}
	return fetchedUrl, err
}