package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
)

type Article struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func main() {
	ctx := context.Background()

	cfg := elasticsearch.Config{Addresses: []string{"http://localhost:9200"}}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatal(err)
	}

	indexName := "articles"

	es.Indices.Delete([]string{indexName}, es.Indices.Delete.WithContext(ctx))

	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"title":   map[string]string{"type": "text"},
				"content": map[string]string{"type": "text"},
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

	articles := []Article{
		{ID: "1", Title: "docker Article", Content: "Content 1"},
		{ID: "2", Title: "New Article", Content: "Docker new text"},
		{ID: "3", Title: "Other Article", Content: "Other text"},
	}

	for _, a := range articles {
		req, err := es.Index(
			indexName,
			esutil.NewJSONReader(a),
			es.Index.WithContext(ctx),
			es.Index.WithDocumentID(a.ID),
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
			score := hitMap["_score"].(float64)
			source := hitMap["_source"]

			sourceBytes, _ := json.Marshal(source)
			var a Article
			json.Unmarshal(sourceBytes, &a)

			log.Printf("ID: %s, Title: %s, Content: %s, Score: %f", a.ID, a.Title, a.Content, score)
		}
	}

	// 1
	text := "Docker"
	q1 := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  text,
				"fields": []string{"title", "content"},
			},
		},
	}
	log.Println("q1 result:")
	search(q1)

	log.Println("Done")
}
