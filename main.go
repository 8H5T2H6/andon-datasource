package main

import (
	"DataSource/config"
	"DataSource/router"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func initGlobalVar() {
	err := godotenv.Load(config.EnvName)
	if err != nil {
		log.Fatalf("Error loading env file")
	}

	config.MongodbURL = os.Getenv("MONGODB_URL")
	config.MongodbDatabase = os.Getenv("MONGODB_DATABASE")
	config.MongodbUsername = os.Getenv("MONGODB_USERNAME")
	config.MongodbPassword = os.Getenv("MONGODB_PASSWORD")

	config.TaipeiTimeZone, _ = time.LoadLocation("Asia/Taipei")
	config.UTCTimeZone, _ = time.LoadLocation("UTC")
}

func main() {
	initGlobalVar()
	log.Println("Activation")
	fmt.Println("Version ->", "1.0.0")
	fmt.Println("MongoDB ->", "URL:", config.MongodbURL, " Database:", config.MongodbDatabase)
	r := router.Router()
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	})
	handler := c.Handler(r)
	log.Fatal(http.ListenAndServe(":8080", handler))
}
