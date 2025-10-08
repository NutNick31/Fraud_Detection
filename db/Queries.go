package db
const (
	QueryComponents = `
    MATCH (n)-[r]-()
    WITH n, count(r) AS degree
    WHERE degree > 1
    CALL apoc.path.subgraphAll(n, {}) YIELD nodes, relationships
    // map nodes and relationships to useful serializable maps
    WITH
      [x IN nodes | { id: id(x), labels: labels(x), props: properties(x) }] AS compNodes,
      [rel IN relationships | {
         id: id(rel),
         startId: id(startNode(rel)),
         endId: id(endNode(rel)),
         type: type(rel),
         props: properties(rel)
      }] AS compRels
    // compute a canonical leader for deduplication (minimum node id)
    WITH compNodes, compRels, apoc.coll.min([n IN compNodes | n.id]) AS leader
    // group by leader and collect one component per leader
    WITH leader, head(collect({ leader: leader, nodes: compNodes, relationships: compRels })) AS comp
    WITH collect(comp) AS components
    RETURN components;
  `

  TempQuery = `
    MATCH (n)-[r]-()
    WITH n, count(r) AS degree
    WHERE degree > 1
    CALL apoc.path.subgraphAll(n, {relationshipFilter: 'MATCHES'}) YIELD nodes, relationships
    // map nodes and relationships to useful serializable maps
    WITH
      [x IN nodes | { id: x.id }] AS compNodes,
      [rel IN relationships | {
         startId: startNode(rel).id,
         endId: endNode(rel).id,
         type: type(rel)
      }] AS compRels
    // compute a canonical leader for deduplication (minimum node id)
    WITH compNodes, compRels, apoc.coll.min([n IN compNodes | n.id]) AS leader
    // group by leader and collect one component per leader
    WITH leader, head(collect({ nodes: compNodes})) AS comp
    WITH collect(comp) AS components
    RETURN components;
  `

  QueryComponentsId = `
    MATCH (start)
    WHERE (start.id = $nodeIdStr)
    WITH start
    LIMIT 1
    CALL apoc.path.subgraphAll(start, {relationshipFilter: 'MATCHES'}) YIELD nodes, relationships
    WITH
      [x IN nodes | { 
        id: x.id,
        labels: labels(x)
        // props: properties(x)
      }] AS compNodes, 

      [rel IN relationships | {
        // id: id(rel),
        startId: startNode(rel).id,
        endId: endNode(rel).id,
        type: type(rel),
        props: properties(rel)
      }] AS compRels
    RETURN { component: { nodes: compNodes, relationships: compRels } } AS result;
  `
  QueryComponent = 
  `MATCH (start)
WHERE start.id = $nodeIdStr
WITH start
LIMIT 1

// discover component by walking only MATCHES
CALL apoc.path.subgraphAll(start, { relationshipFilter: 'MATCHES' }) YIELD nodes AS subNodes
WITH subNodes

// ids of component nodes for membership tests
WITH subNodes, [n IN subNodes | n.id] AS compIds

// collect all relationships that touch any component node (internal + to neighbours)
MATCH (a)-[rel]-(b)
WHERE a.id IN compIds OR b.id IN compIds
WITH subNodes,
     collect(DISTINCT {
       startId: startNode(rel).id,
       endId: endNode(rel).id,
       type: type(rel),
       props: properties(rel)
     }) AS compRels,
     apoc.coll.toSet(collect(DISTINCT startNode(rel).id) + collect(DISTINCT endNode(rel).id)) AS touchingIds

// fetch all nodes that appear as endpoints in those relationships (component nodes + neighbours)
MATCH (m)
WHERE m.id IN touchingIds
WITH subNodes, compRels, collect(DISTINCT { id: m.id, labels: labels(m) /*, props: properties(m) */ }) AS allNodes

// return same shape as you originally had
RETURN { component: { nodes: allNodes, relationships: compRels } } AS result;
`

ComponentIndegreeQuery = `
    MATCH (n)
    OPTIONAL MATCH (m)-[r]->(n)
    WITH n, count(r) AS indegree
    WHERE indegree >= $min
    RETURN id(n)         AS node_internal_id,
          labels(n)     AS labels,
          properties(n) AS props,
          indegree
    ORDER BY indegree DESC;

`

)