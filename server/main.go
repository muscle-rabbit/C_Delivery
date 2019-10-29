package main

import (
	"fmt"
	"os"

	"golang.org/x/net/context"

	firebase "firebase.google.com/go"
	// "firebase.google.com/go/auth"
	"github.com/joho/godotenv"

	"google.golang.org/api/option"
)

func main() {
	fmt.Println(initializeAppDefault())
}
func initializeAppDefault() (*firebase.App, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error get env: %v\n", err)
	}

	// CREDENTIAL 
	opt := option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %v\n", err)
	}

	return app, nil
}
