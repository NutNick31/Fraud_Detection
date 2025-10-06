package models

type Node struct {
	Id string `json: "id"`
	Label string `json: "label"`
}

type Relationship struct {
	StartId string `json: "start_id"`
	EndId string `json: "end_id"`
	Type string `json: "type"`
	Properties map[string][int] `json: "props"`
}