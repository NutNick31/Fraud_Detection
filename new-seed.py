import pandas as pd
from typing import List, Dict
from neo4j import GraphDatabase

# Neo4j Config
URI = "neo4j://host.docker.internal:7687"
USER = "neo4j"
PWD = "sample-db-password"
DB = "neo4j"

driver = GraphDatabase.driver(URI, auth=(USER, PWD))


# Neo4j Write Helper
def run_batch_write(query: str, rows: List[Dict], batch_size: int = 500):
    """
    Executes a write query in batches. Each batch is passed in as parameter 'rows'.
    Query should UNWIND $rows AS row and operate on row.query_refid / row.probe_refid.
    """
    with driver.session(database=DB) as session:
        for i in range(0, len(rows), batch_size):
            batch = rows[i : i + batch_size]
            session.execute_write(lambda tx: tx.run(query, {"rows": batch}).consume())


# Cypher Query for 2-column CSV (query_refid -> probe_refid)
CYPHER_CSV_TXNS = """
UNWIND $rows AS row
MERGE (q:Person {id: row.query_refid})
MERGE (p:Person {id: row.probe_refid})
MERGE (q)-[:MATCHES{weight: 0.5}]->(p)
"""


# Load CSV and convert to dict rows
def read_csv_to_rows(path: str) -> List[Dict]:
    """
    Expects a CSV with two columns: query_refid and probe_id (case-insensitive).
    Returns list of dicts: {"query_id": ..., "probe_id": ...}
    """
    df = pd.read_csv(path)
    # sanitize headers: strip and lower
    df.columns = [c.strip().lower() for c in df.columns]

    # determine column names (allow some common variants)
    # prefer exact 'query_id' and 'probe_id' but fall back to first two columns
    if "query_refid" in df.columns and "probe_refid" in df.columns:
        q_col = "query_refid"
        p_col = "probe_refid"
    else:
        # fallback: use the first two columns as query / probe
        if len(df.columns) < 2:
            raise ValueError("CSV must contain at least two columns (query_refid, probe_refid).")
        q_col, p_col = df.columns[0], df.columns[1]

    rows = []
    for _, row in df.iterrows():
        qv = row[q_col]
        pv = row[p_col]
        # convert to string and strip whitespace; skip rows with missing IDs
        if pd.isna(qv) or pd.isna(pv):
            continue
        rows.append(
            {
                "query_refid": str(qv).strip(),
                "probe_refid": str(pv).strip(),
            }
        )
    return rows


# Optional: create uniqueness constraints (run once; safe if DB already has them on recent Neo4j versions)
CONSTRAINTS = """
// Note: For Neo4j 4.x+ this will create constraints if they don't exist.
// If your Neo4j version is older/different, you can remove or adapt these lines.
CREATE CONSTRAINT IF NOT EXISTS FOR (p:Person) REQUIRE p.id IS UNIQUE;
"""


def ensure_constraints():
    with driver.session(database=DB) as session:
        # run as write transaction
        session.execute_write(lambda tx: tx.run(CONSTRAINTS).consume())


# Main function
def main():
    csv_path = "dddd.csv"  # update filename if needed

    print(f"Reading data from {csv_path} ...")
    rows = read_csv_to_rows(csv_path)
    print(f"Loaded {len(rows)} rows. Inserting into Neo4j ...")

    if not rows:
        print("No rows to insert. Exiting.")
        return

    try:
        # optional: create uniqueness constraints (uncomment if you want)
        # ensure_constraints()

        run_batch_write(CYPHER_CSV_TXNS, rows)
        print("CSV data seeding complete.")
    except Exception as e:
        print("Error while writing to Neo4j:", e)
        raise


if __name__ == "__main__":
    try:
        main()
    finally:
        driver.close()
