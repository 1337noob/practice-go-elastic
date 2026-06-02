package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
)

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	ctx := context.Background()

	cfg := elasticsearch.Config{Addresses: []string{"http://localhost:9200"}}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatal(err)
	}

	indexName := "users"

	es.Indices.Delete([]string{indexName}, es.Indices.Delete.WithContext(ctx))

	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"name":  map[string]string{"type": "text"},
				"email": map[string]string{"type": "keyword"},
			},
		},
	}

	createRes, err := es.Indices.Create(
		indexName,
		es.Indices.Create.WithContext(ctx),
		es.Indices.Create.WithBody(esutil.NewJSONReader(mapping)),
	)
	if err != nil {
		log.Fatal(err)
	}
	createRes.Body.Close()
	log.Println("Index created")

	users := []User{
		{ID: "1", Name: "John", Email: "jonh@example.com"},
		{ID: "2", Name: "Bob", Email: "bob@example.com"},
		{ID: "3", Name: "Alice", Email: "alice@example.com"},
	}

	for _, u := range users {
		req, err := es.Index(
			indexName,
			esutil.NewJSONReader(u),
			es.Index.WithContext(ctx),
			es.Index.WithDocumentID(u.ID),
			es.Index.WithRefresh("true"),
		)
		if err != nil {
			log.Fatal(err)
		}
		req.Body.Close()
	}

	search := func(q map[string]interface{}) {
		searchRes, err := es.Search(
			es.Search.WithContext(ctx),
			es.Search.WithIndex(indexName),
			es.Search.WithBody(esutil.NewJSONReader(q)),
			es.Search.WithPretty(),
		)
		if err != nil {
			log.Fatal(err)
		}
		defer searchRes.Body.Close()

		// 6. Разбор результатов
		var result map[string]interface{}
		if err := json.NewDecoder(searchRes.Body).Decode(&result); err != nil {
			log.Fatal(err)
		}

		hits := result["hits"].(map[string]interface{})["hits"].([]interface{})

		if len(hits) == 0 {
			fmt.Println("No results found")
			return
		}

		for _, hit := range hits {
			hitMap := hit.(map[string]interface{})
			source := hitMap["_source"]

			sourceBytes, _ := json.Marshal(source)
			var u User
			json.Unmarshal(sourceBytes, &u)

			log.Printf("ID: %s, Name: %s, Email: %s", u.ID, u.Name, u.Email)
		}
	}

	// 1
	email := "bob@example.com"
	q1 := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"email": email,
			},
		},
	}
	log.Println("q1 result:")
	search(q1)

	log.Println("Done")
}
