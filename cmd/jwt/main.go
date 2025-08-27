package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Code-byme/e-commerce/pkg/utils"
)

func main() {
	var (
		action = flag.String("action", "generate", "Action to perform: generate, validate, or decode")
		userID = flag.Uint("user-id", 1, "User ID for token generation")
		email  = flag.String("email", "test@example.com", "Email for token generation")
		role   = flag.String("role", "customer", "Role for token generation")
		token  = flag.String("token", "", "JWT token to validate or decode")
		pretty = flag.Bool("pretty", false, "Pretty print JSON output")
	)
	flag.Parse()

	generator := utils.NewJWTGenerator()

	switch *action {
	case "generate":
		generateToken(generator, *userID, *email, *role, *pretty)
	case "validate":
		validateToken(generator, *token, *pretty)
	case "decode":
		decodeToken(generator, *token, *pretty)
	default:
		fmt.Println("Invalid action. Use: generate, validate, or decode")
		flag.Usage()
		os.Exit(1)
	}
}

func generateToken(generator *utils.JWTGenerator, userID uint, email, role string, pretty bool) {
	token, err := generator.GenerateTestToken(userID, email, role)
	if err != nil {
		log.Fatalf("Failed to generate token: %v", err)
	}

	output := map[string]interface{}{
		"token": token,
		"claims": map[string]interface{}{
			"user_id": userID,
			"email":   email,
			"role":    role,
		},
	}

	printJSON(output, pretty)
}

func validateToken(generator *utils.JWTGenerator, token string, pretty bool) {
	if token == "" {
		log.Fatal("Token is required for validation")
	}

	claims, err := generator.ValidateToken(token)
	if err != nil {
		log.Fatalf("Token validation failed: %v", err)
	}

	output := map[string]interface{}{
		"valid":  true,
		"claims": claims,
	}

	printJSON(output, pretty)
}

func decodeToken(generator *utils.JWTGenerator, token string, pretty bool) {
	if token == "" {
		log.Fatal("Token is required for decoding")
	}

	claims, err := generator.DecodeToken(token)
	if err != nil {
		log.Fatalf("Token decoding failed: %v", err)
	}

	output := map[string]interface{}{
		"claims": claims,
	}

	printJSON(output, pretty)
}

func printJSON(data interface{}, pretty bool) {
	var jsonData []byte
	var err error

	if pretty {
		jsonData, err = json.MarshalIndent(data, "", "  ")
	} else {
		jsonData, err = json.Marshal(data)
	}

	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	fmt.Println(string(jsonData))
}
