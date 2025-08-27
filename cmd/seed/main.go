package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/Code-byme/e-commerce/internal/database"
	"github.com/Code-byme/e-commerce/internal/services"
)

func main() {
	// Parse command line flags
	var (
		limit = flag.Int("limit", 20, "Number of products to fetch from API")
		stats = flag.Bool("stats", false, "Show database statistics")
		help  = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *help {
		fmt.Println("Database Seeding Tool")
		fmt.Println("=====================")
		fmt.Println("This tool fetches products from FakeStore API and stores them in your database.")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  go run cmd/seed/main.go [options]")
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  go run cmd/seed/main.go                    # Seed with 20 products (default)")
		fmt.Println("  go run cmd/seed/main.go -limit 50          # Seed with 50 products")
		fmt.Println("  go run cmd/seed/main.go -stats             # Show database statistics")
		fmt.Println("  go run cmd/seed/main.go -limit 10 -stats   # Seed 10 products and show stats")
		return
	}

	// Initialize database
	if err := database.InitDatabase(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// Run migrations
	if err := database.RunMigrations(); err != nil {
		log.Fatal("Failed to run database migrations:", err)
	}

	// Create seed service
	seedService := services.NewSeedService()

	// Show stats if requested
	if *stats {
		fmt.Println("Current Database Statistics:")
		fmt.Println("============================")
		if err := seedService.GetDatabaseStats(); err != nil {
			log.Fatal("Failed to get database stats:", err)
		}
		fmt.Println()
	}

	// Seed database if limit > 0
	if *limit > 0 {
		fmt.Printf("Seeding database with %d products from FakeStore API...\n", *limit)
		fmt.Println("==================================================")

		if err := seedService.SeedProducts(*limit); err != nil {
			log.Fatal("Failed to seed database:", err)
		}

		fmt.Println("\nSeeding completed successfully!")
	}

	// Show final stats
	if *limit > 0 || *stats {
		fmt.Println("\nFinal Database Statistics:")
		fmt.Println("==========================")
		if err := seedService.GetDatabaseStats(); err != nil {
			log.Fatal("Failed to get final database stats:", err)
		}
	}

	fmt.Println("\n✅ Database seeding tool completed!")
}
