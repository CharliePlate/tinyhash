package main

import (
	"log"
	"os"

	"github.com/charlieplate/TinyHash/ui/internal/view"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			e.Logger.Fatal("Error loading .env file")
		}
		log.Println(".env file loaded")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Starting server on port", port)
	e.GET("/", view.Home)
	e.GET("/static/*", view.ServeStaticFiles)

	log.Printf("Server is running on port %s :)", port)
	e.Logger.Fatal(e.Start("0.0.0.0:" + port))
}
