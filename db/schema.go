package db

// Node Labels
const (
	PersonLabel           = "Person"
	OperatorLabel         = "Operator"
	TransactionLabel      = "Transaction"
	DeviceLabel           = "Device"
	MachineLabel          = "Machine"
	StationLabel          = "Station"
	EnrollmentAgencyLabel = "EnrollmentAgency"
)

// Relationship Types
const (
	INTRODUCED_BY_REL = "INTRODUCED_BY" // (Person)-[:INTRODUCED_BY]->(Person)
	HAS_CONTACT_REL   = "HAS_CONTACT"   // (Person)-[:HAS_CONTACT]->(Contact)
	LIVES_AT_REL      = "LIVES_AT"      // (pERSON) -[:LIVES_AT]->(Location)
	PERFORMED_BY_REL  = "PERFORMED_BY"  // (Transaction)-[:PERFORMED_BY]->(Operator)
	FOR_PERSON_REL    = "FOR_PERSON"    // (Transaction)-[:FOR_PERSON]->(Person)
	USED_DEVICE_REL   = "USED_DEVICE"   // (Transaction)-[:USED_DEVICE]->(Device)
	USED_MACHINE_REL  = "USED_MACHINE"  // (Transaction)-[:USED_STATION]->(Machine)
	HAS_MACHINE_REL   = "HAS_MACHINE"   // (Station)-[:HAS_MACHINE]->(Machine)
	WORKS_ON_REL      = "WORKS_ON"      // (Operator)-[:WORKS_ON]->(Machine)
	CONNECTED_TO_REL  = "CONNECTED_TO"  // (Operator)-[:CONNECTED_TO {freq:n}]->(Person)
)


