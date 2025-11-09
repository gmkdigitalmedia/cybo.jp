package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run gen_password.go <password>")
		os.Exit(1)
	}

	password := os.Args[1]
	
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		fmt.Printf("Error generating hash: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nPassword hash generated successfully!\n\n")
	fmt.Printf("Add this to your config.env file:\n")
	fmt.Printf("PASSWORD_HASH=%s\n\n", string(hash))
}
