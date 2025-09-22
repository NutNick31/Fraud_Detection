package routes

import (
	"sample_server/db"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func RegisterRoutes(r *gin.Engine) {
	
	r.DELETE("/reset", resetDatabase)
	r.GET("/components", getComponents)
}

func resetDatabase(c *gin.Context) {

	session := db.Driver.NewSession(c, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(c)

	// delete database
	query := "MATCH (n) DETACH DELETE n"
	_, err := session.ExecuteWrite(c, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		_, err := tx.Run(c, query, nil)
		return nil, err
	})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "Database reset successful"})
}

func getComponents(c *gin.Context) {

	session := db.Driver.NewSession(c, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(c)

	results, err := session.ExecuteWrite(c, func(tx neo4j.ManagedTransaction) (interface{}, error){
		records, err := tx.Run(c, db.QueryComponents, nil)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return nil, err
		}
		return records, err
	})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return 
	}
	c.JSON(200, results)
}


