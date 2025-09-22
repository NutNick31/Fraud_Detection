import pandas as pd
from typing import List, Dict
from neo4j import GraphDatabase

# Neo4j Config
URI = "neo4j://host.docker.internal:7687"
USER = "neo4j"
PWD  = "sample-db-password"
DB   = "neo4j"

driver = GraphDatabase.driver(URI, auth=(USER, PWD))

# Neo4j Write Helper
def run_batch_write(query: str, rows: List[Dict], batch_size: int = 500):
    with driver.session(database=DB) as session:
        for i in range(0, len(rows), batch_size):
            batch = rows[i:i+batch_size]
            session.execute_write(lambda tx: tx.run(query, {"rows": batch}).consume())

# Cypher Query
CYPHER_CSV_TXNS = """
UNWIND $rows AS row

MERGE (p:Person {id: row.uid})
MERGE (t:Transaction {id: row.refid})
MERGE (o:Operator {id: row.operator_id})
MERGE (m:Machine {id: row.machine_id})

MERGE (t)-[:FOR_PERSON]->(p)
MERGE (t)-[:PERFORMED_BY]->(o)
MERGE (t)-[:USED_MACHINE]->(m)
MERGE (o)-[:WORKS_ON]->(m)

MERGE (o)-[ct:CONNECTED_TO]->(p)
  ON CREATE SET ct.freq = 1
  ON MATCH  SET ct.freq = ct.freq + 1
"""

# Load CSV and convert to dict rows
def read_csv_to_rows(path: str) -> List[Dict]:
    df = pd.read_csv(path)
    df.columns = [c.strip() for c in df.columns]  # sanitize headers
    return [
        {
            "uid": str(row["UID"]),
            "refid": str(row["Refid"]),
            "operator_id": str(row["Operator ID"]),
            "machine_id": str(row["Source Client Machine ID"])
        }
        for _, row in df.iterrows()
    ]

# Main function
def main():
    csv_path = "dddd.csv"
    
    print(f"Reading data from {csv_path} ...")
    rows = read_csv_to_rows(csv_path)
    print(f"Loaded {len(rows)} rows. Inserting into Neo4j ...")
    run_batch_write(CYPHER_CSV_TXNS, rows)
    print("CSV data seeding complete.")

if __name__ == "__main__":
    try:
        main()
    finally:
        driver.close()
