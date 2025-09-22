package db
const(
  QueryComponents = `
    MATCH (n)-[r]-()
    WITH n, count(r) AS degree
    WHERE degree > 1
    CALL apoc.path.subgraphAll(n, {}) YIELD nodes, relationships
    RETURN n AS startNode, nodes AS componentNodes, relationships AS componentRels;

  `
)
