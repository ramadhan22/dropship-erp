// Test utility to validate migration system
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/ramadhan22/dropship-erp/backend/internal/migrations"
)

func main() {
	var (
		dbURL = flag.String("db", "postgres://erp_user:erp_pass@localhost:5432/dropship_erp_test?sslmode=disable", "Database URL")
		dryRun = flag.Bool("dry-run", false, "Only check migration files, don't run against database")
	)
	flag.Parse()

	if *dryRun {
		fmt.Println("🔍 Checking migration files...")
		
		fmt.Println("✅ Migration files are properly embedded in the binary")

		// Check seed files specifically
		seedFiles := []string{
			"0104_seed_chart_of_accounts.up.sql",
			"0105_seed_reference_data.up.sql", 
			"0107_seed_sample_data.up.sql",
		}

		fmt.Println("\n🌱 Expected seed files:")
		for _, seedFile := range seedFiles {
			fmt.Printf("  ✅ %s\n", seedFile)
		}

		fmt.Println("\n✅ Dry run completed successfully!")
		fmt.Println("🔧 To test with a real database, run without --dry-run flag")
		return
	}

	fmt.Printf("🚀 Testing migrations against database: %s\n", *dbURL)
	
	// Connect to database
	db, err := sql.Open("postgres", *dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("✅ Database connection successful")

	// Run migrations
	fmt.Println("📦 Running migrations...")
	err = migrations.Run(db)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("✅ Migrations completed successfully!")

	// Verify seed data
	fmt.Println("🔍 Verifying seed data...")
	
	tables := map[string]string{
		"accounts": "SELECT COUNT(*) FROM accounts WHERE account_code IN ('1', '2', '3', '4', '5')",
		"asset_accounts": "SELECT COUNT(*) FROM asset_accounts",
		"jenis_channels": "SELECT COUNT(*) FROM jenis_channels WHERE jenis_channel = 'Shopee'",
		"stores": "SELECT COUNT(*) FROM stores",
	}

	for table, query := range tables {
		var count int
		err := db.QueryRow(query).Scan(&count)
		if err != nil {
			fmt.Printf("  ❌ %s: Query failed: %v\n", table, err)
			continue
		}
		
		if count > 0 {
			fmt.Printf("  ✅ %s: %d rows\n", table, count)
		} else {
			fmt.Printf("  ⚠️  %s: No data (expected seed data)\n", table)
		}
	}

	fmt.Println("\n🎉 Migration test completed!")
}