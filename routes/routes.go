package routes

import (
	"log"
	"net/http"
	"sample_server/db"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func RegisterRoutes(r *gin.Engine) {

	r.DELETE("/reset", resetDatabase)
	r.GET("/components", getAllComponents)
	r.GET("/components/:id", getComponents)
	r.GET("/compdegree", getNodesByIndegree)
	r.GET("/refsimilar", getRefidSimilarities)
	r.GET("/sameop", getUIDOperatorMatches)
}

func getUIDOperatorMatches(c *gin.Context) {
	// open a read session using the gin context as context.Context
	session := db.Driver.NewSession(c, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(c)

	resultObj, err := session.ExecuteRead(c, func(tx neo4j.ManagedTransaction) (any, error) {
		// Cypher: collect distinct nodes and relationships that match the pattern,
		// then return nodes and relationships with safe id fallbacks.
		query := `
MATCH (u:UID)<-[b1:Belongs_To]-(o:Operator)<-[opRel:Operated_By]-(r:Person)-[mr:MATCHES]->(rr:Person)-[b2:Belongs_To]->(u)
WITH collect(DISTINCT u) AS u_nodes,
     collect(DISTINCT o) AS op_nodes,
     collect(DISTINCT r) AS r_nodes,
     collect(DISTINCT rr) AS rr_nodes,
     collect(DISTINCT b1) AS belongs1,
     collect(DISTINCT opRel) AS oper_rels,
     collect(DISTINCT mr) AS matches_rels,
     collect(DISTINCT b2) AS belongs2

// merge node lists and relationship lists
WITH u_nodes + op_nodes + r_nodes + rr_nodes AS all_nodes,
     belongs1 + oper_rels + matches_rels + belongs2 AS all_rels

// build node objects (node_id + labels)
UNWIND all_nodes AS n
WITH collect(DISTINCT { node_id: coalesce(n.id, toString(id(n))), labels: labels(n) }) AS nodes, all_rels AS rels

// build relationship objects with safe node id fallbacks
UNWIND rels AS rel
RETURN nodes,
       collect(DISTINCT {
         start_node_id: coalesce(startNode(rel).id, toString(id(startNode(rel)))),
         end_node_id:   coalesce(endNode(rel).id,   toString(id(endNode(rel)))),
         type:          type(rel)
       }) AS relationships;
		`

		// run
		res, runErr := tx.Run(c, query, nil)
		if runErr != nil {
			return nil, runErr
		}

		// if no row, return empty arrays
		if !res.Next(c) {
			return gin.H{
				"nodes":         []any{},
				"relationships": []any{},
			}, nil
		}

		if iterErr := res.Err(); iterErr != nil {
			return nil, iterErr
		}

		rec := res.Record()
		nodesAny, _ := rec.Get("nodes")
		relsAny, _ := rec.Get("relationships")

		return gin.H{
			"nodes":         nodesAny,
			"relationships": relsAny,
		}, nil
	})

	if err != nil {
		log.Printf("getUIDOperatorMatches: neo4j query failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "database query failed",
			"details": err.Error(), // remove details in production
		})
		return
	}

	c.JSON(http.StatusOK, resultObj)
}



func getRefidSimilarities(c *gin.Context) {
	refid := c.Query("refid")
	if refid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing 'refid' query param (e.g. ?refid=ABC123)"})
		return
	}

	session := db.Driver.NewSession(c, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(c)

	resultObj, err := session.ExecuteRead(c, func(tx neo4j.ManagedTransaction) (any, error) {
		// <-- full Cypher (safe fallbacks for missing node.id)
		query := `
MATCH (target:Person {id: $refid})
OPTIONAL MATCH (r:Person)-[mr:MATCHES]->(target)
OPTIONAL MATCH (r)-[b:Belongs_To]->(uid:UID)
OPTIONAL MATCH (r)-[opRel:Operated_By]->(op:Operator)
WITH target,
     collect(DISTINCT r) AS incoming_nodes,
     [u IN collect(DISTINCT uid) WHERE u IS NOT NULL | u] AS uids,
     [o IN collect(DISTINCT op)  WHERE o IS NOT NULL | o] AS ops,
     [m IN collect(DISTINCT mr)  WHERE m IS NOT NULL | m] AS matches_rels,
     [bb IN collect(DISTINCT b) WHERE bb IS NOT NULL | bb] AS belongs_rels,
     [oo IN collect(DISTINCT opRel) WHERE oo IS NOT NULL | oo] AS oper_rels

WITH target,
     [{ node_id: coalesce(target.id, toString(id(target))), labels: labels(target) }] +
     [ n IN incoming_nodes | { node_id: coalesce(n.id, toString(id(n))), labels: labels(n) } ] +
     [ u IN uids           | { node_id: coalesce(u.id, toString(id(u))), labels: labels(u) } ] +
     [ o IN ops            | { node_id: coalesce(o.id, toString(id(o))), labels: labels(o) } ] AS nodes,
     matches_rels, belongs_rels, oper_rels

WITH nodes,
  [ r IN matches_rels  | { start_node_id: coalesce(startNode(r).id, toString(id(startNode(r)))), end_node_id: coalesce(endNode(r).id, toString(id(endNode(r)))), type: type(r) } ] +
  [ r IN belongs_rels  | { start_node_id: coalesce(startNode(r).id, toString(id(startNode(r)))), end_node_id: coalesce(endNode(r).id, toString(id(endNode(r)))), type: type(r) } ] +
  [ r IN oper_rels     | { start_node_id: coalesce(startNode(r).id, toString(id(startNode(r)))), end_node_id: coalesce(endNode(r).id, toString(id(endNode(r)))), type: type(r) } ] AS relationships

RETURN nodes, relationships;
		`

		params := map[string]any{"refid": refid}
		res, runErr := tx.Run(c, query, params)
		if runErr != nil {
			return nil, runErr
		}

		// If there is no result row, return the target node minimally (client expects arrays)
		if !res.Next(c) {
			// make a best-effort fallback: return empty arrays (or you can return target-only)
			return gin.H{
				"nodes":         []any{},
				"relationships": []any{},
			}, nil
		}

		// check for iteration error
		if iterErr := res.Err(); iterErr != nil {
			return nil, iterErr
		}

		rec := res.Record()
		nodesAny, _ := rec.Get("nodes")
		relsAny, _ := rec.Get("relationships")

		return gin.H{
			"nodes":         nodesAny,
			"relationships": relsAny,
		}, nil
	})

	if err != nil {
		// log full error server-side for debugging
		log.Printf("getRefidSimilarities: neo4j query failed for refid=%s: %v", refid, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "database query failed",
			"details": err.Error(), // remove 'details' in production to avoid leaking internals
		})
		return
	}

	c.JSON(http.StatusOK, resultObj)
}


func getNodesByIndegree(c *gin.Context) {
	minStr := c.Query("min")
	if minStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing 'min' query param (e.g. ?min=3)"})
		return
	}
	min, err := strconv.Atoi(minStr)
	if err != nil || min < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "'min' must be a non-negative integer"})
		return
	}

	// c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()

	// Open a read session
	session := db.Driver.NewSession(c, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(c)

	records, err := session.ExecuteRead(c, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
				MATCH (n)
				OPTIONAL MATCH (m:Person)-[r]->(n:Person)
				WITH n, count(r) AS indegree
				WHERE indegree >= $min
				RETURN n.id AS node_id, indegree
				ORDER BY indegree DESC
			`
		params := map[string]any{"min": min}
		result, err := tx.Run(c, query, params)
		if err != nil {
			return nil, err
		}

		var out []map[string]any
		for result.Next(c) {
			rec := result.Record()
			nodeID, _ := rec.Get("node_id")
			indeg, _ := rec.Get("indegree")

			out = append(out, map[string]any{
				"node_id":    nodeID,
				"indegree":   indeg,
			})
		}
		if err := result.Err(); err != nil {
			return nil, err
		}
		return out, nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, records)
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

	id := c.Param("id")
	params := map[string]any{
		"nodeIdStr": id,
	}
	result, err := session.ExecuteWrite(c, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		records, err := tx.Run(c, db.QueryComponentsId, params)
		if err != nil {
			return nil, err
		}
		var components interface{} = []interface{}{}

		if records.Next(c) {
			rec := records.Record()
			if v, ok := rec.Get("result"); ok {
				components = v
			}
		} else if err := records.Err(); err != nil {
			return nil, err
		}
		return components, nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func getAllComponents(c *gin.Context) {

	session := db.Driver.NewSession(c, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(c)

	id := c.Param("id")
	params := map[string]any{
		"nodeIdStr": id,
	}
	result, err := session.ExecuteWrite(c, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		records, err := tx.Run(c, db.TempQuery, params)
		if err != nil {
			return nil, err
		}
		var components interface{} = []interface{}{}

		if records.Next(c) {
			rec := records.Record()
			if v, ok := rec.Get("components"); ok {
				components = v
			}
		} else if err := records.Err(); err != nil {
			return nil, err
		}
		return components, nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
