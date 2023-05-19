package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/google/uuid"
)

type Entity struct {
	Value     string
	CreatedAt time.Time
}

func main() {
	fmt.Println("Hello, Kubernetes Novice!")

	ctx := context.Background()
	projectID := os.Getenv("PROJECT_ID")

	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	uuid := uuid.New().String()
	key := datastore.NameKey("entity_k8snovice", uuid, nil)
	e := new(Entity)
	emojis := []string{"😀", "😃", "😄", "😁", "😆", "😅", "😂", "🤣", "😊", "😇"}
	rand.Seed(time.Now().UnixNano())
	randomEmoji := emojis[rand.Intn(len(emojis))]

	e.Value = "Hi! Kubernetes Novice " + randomEmoji
	loc, _ := time.LoadLocation("Asia/Tokyo")
	e.CreatedAt = time.Now().In(loc)

	if _, err := client.Put(ctx, key, e); err != nil {
		log.Fatalf("Failed to save entity: %v", err)
	}

	e = new(Entity)
	if err = client.Get(ctx, key, e); err != nil {
		log.Fatalf("Failed to get entity: %v", err)
	}

	fmt.Printf("Fetched entity: %v", e)
}
