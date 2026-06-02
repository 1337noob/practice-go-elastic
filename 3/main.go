package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
)

type Product struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Price    int    `json:"price"`
}

func main() {
	ctx := context.Background()

	cfg := elasticsearch.Config{Addresses: []string{"http://localhost:9200"}}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatal(err)
	}

	indexName := "products"

	es.Indices.Delete([]string{indexName}, es.Indices.Delete.WithContext(ctx))

	createRes, err := es.Indices.Create(
		indexName,
		es.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		log.Fatal(err)
	}
	createRes.Body.Close()
	log.Println("Index created")

	products := []Product{
		{
			ID:       "1",
			Name:     "Laptop",
			Category: "Electronics",
			Price:    1000,
		},
		{
			ID:       "2",
			Name:     "Phone",
			Category: "Electronics",
			Price:    500,
		},
		{
			ID:       "3",
			Name:     "Watermelon",
			Category: "Food",
			Price:    20,
		},
	}

	for _, p := range products {
		req, err := es.Index(
			indexName,
			esutil.NewJSONReader(p),
			es.Index.WithContext(ctx),
			es.Index.WithDocumentID(p.ID),
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
			var p Product
			json.Unmarshal(sourceBytes, &p)

			log.Printf("ID: %s, Name: %s, Category: %s, Price: %d\n", p.ID, p.Name, p.Category, p.Price)
		}
	}

	// 1
	q1 := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}
	log.Println("q1 result:")
	search(q1)

	// 2
	q2 := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"category": "Electronics",
			},
		},
	}
	log.Println("q2 result:")
	search(q2)

	// 3
	q3 := map[string]interface{}{
		"query": map[string]interface{}{
			"range": map[string]interface{}{
				"price": map[string]interface{}{
					"gte": 500,
				},
			},
		},
	}
	log.Println("q3 result:")
	search(q3)

	log.Println("Done")
}
