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
    CALL apoc.path.subgraphAll(n, {}) YIELD nodes, relationships
    // map nodes and relationships to useful serializable maps
    WITH
      [x IN nodes | { id: x.id, labels: labels(x), props: properties(x) }] AS compNodes,
      [rel IN relationships | {
         startId: startNode(rel).id,
         endId: endNode(rel).id,
         type: type(rel)
      }] AS compRels
    // compute a canonical leader for deduplication (minimum node id)
    WITH compNodes, compRels, apoc.coll.min([n IN compNodes | n.id]) AS leader
    // group by leader and collect one component per leader
    WITH leader, head(collect({ nodes: compNodes, relationships: compRels })) AS comp
    WITH collect(comp) AS components
    RETURN components;
  `
)