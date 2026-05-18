package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"ariadne/internal/provider/embedding"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	provider := embedding.NewDashScopeEmbeddingProvider()

	if !provider.IsConfigured() {
		log.Fatal("DASHSCOPE_API_KEY is missing")
	}

	text := "用户不喜欢青年旅舍，偏好地铁附近的经济型连锁酒店。"

	vector, err := provider.EmbedDocument(ctx, text)
	if err != nil {
		log.Fatal("embed document failed: ", err)
	}

	fmt.Println("embedding generated successfully")
	fmt.Println("dimension:", len(vector))
	fmt.Println("first 5 values:", vector[:5])

	queryVector, err := provider.EmbedQuery(ctx, "帮我规划北京三日游，酒店尽量方便")
	if err != nil {
		log.Fatal("embed query failed: ", err)
	}

	fmt.Println("query embedding generated successfully")
	fmt.Println("query dimension:", len(queryVector))
	fmt.Println("query first 5 values:", queryVector[:5])
}