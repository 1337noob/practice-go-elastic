package main

import (
	"context"
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
			Name:     "Monitor",
			Category: "Electronics",
			Price:    3000,
		},
		{
			ID:       "3",
			Name:     "Watermelon",
			Category: "Food",
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

	log.Println("Done")
}
