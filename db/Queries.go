package db
const(
  QueryComponents = `
    MATCH (n)
  WHERE degree(n) > 1
  CALL apoc.path.subgraphAll(n, {relationshipFilter: ">", minLevel:0}) YIELD nodes, relationships
  RETURN n AS startNode, nodes AS componentNodes, relationships AS componentRels;

  `
)
