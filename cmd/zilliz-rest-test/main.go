package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"ariadne/internal/provider/vector"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	dim := 1024
	if text := os.Getenv("EMBEDDING_DIM"); text != "" {
		value, err := strconv.Atoi(text)
		if err == nil && value > 0 {
			dim = value
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client := vector.NewZillizRESTClient()

	if !client.IsConfigured() {
		log.Fatal("Zilliz config is missing. Please check ZILLIZ_ENDPOINT and ZILLIZ_TOKEN in .env")
	}

	collectionsBefore, err := client.ListCollections(ctx)
	if err != nil {
		log.Fatal("list collections before failed: ", err)
	}

	fmt.Println("collections before:", collectionsBefore)

	collectionsToCreate := []string{
		"ariadne_user_memory",
		"ariadne_plan_memory",
		"ariadne_agent_rules",
		"ariadne_travel_knowledge",
	}

	for _, collectionName := range collectionsToCreate {
		err := client.CreateQuickCollection(ctx, collectionName, dim)
		if err != nil {
			fmt.Println("create collection maybe already exists:", collectionName, err)
			continue
		}

		fmt.Println("created collection:", collectionName)
	}

	collectionsAfter, err := client.ListCollections(ctx)
	if err != nil {
		log.Fatal("list collections after failed: ", err)
	}

	fmt.Println("collections after:", collectionsAfter)
}