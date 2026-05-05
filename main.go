package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	app := &App{}
	_ = godotenv.Load()
	var DBName = os.Getenv("DBNAME")
	var DBUser = os.Getenv("DBUSER")
	var DBPassword = os.Getenv("DBPASSWORD")
	var Host = os.Getenv("DBHOST")
	if err := app.Initialise(DBUser, DBPassword, DBName, Host); err != nil {
    log.Fatal("Failed to initialize the application: ", err)
}
	app.Run("localhost:8080")
}