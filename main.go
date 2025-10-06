package main

// main.go
import (
	"log"
	"sample_server/db"
	"sample_server/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	
)

func main() {

	// if _, err := os.Stat(".env"); err == nil {
	// 	if err := godotenv.Load(); err != nil {
	// 		log.Println("Warning: could not load .env file:", err)
	// 	} else {
	// 		log.Println(".env file loaded")
	// 	}
	// } else {
	// 	log.Println("No .env file found, relying on environment variables")
	// }

	NEO4J_URI := "neo4j://host.docker.internal:7687"
	NEO4J_USER := "neo4j"
	NEO4J_PASSWORD:="sample-db-password"
	// NEO4J_DB := "neo4j"

	db.InitDB(NEO4J_URI, NEO4J_USER, NEO4J_PASSWORD)
	defer db.CloseDB()

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://10.8.12.212:5173"},
		AllowMethods:     []string{"GET"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))
	// Example: trusting only local network proxies
	err := r.SetTrustedProxies([]string{"127.0.0.1", "0.0.0.0/0"})
	if err != nil {
		log.Fatalf("failed to set trusted proxies: %v", err)
	}
	routes.RegisterRoutes(r)

	log.Println("Server is running on port 8080")
	r.Run("0.0.0.0:8080") // Start server
}
