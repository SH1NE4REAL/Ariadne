package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"ariadne/internal/model"
	"ariadne/internal/tools"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	userID := "demo_user"

	writer := tools.NewMemoryWriterTool()
	retriever := tools.NewMemoryRetrieverTool()

	memoryText := "用户不喜欢青年旅舍、多人间、床位房；未来住宿推荐应优先避开青旅，优先选择地铁附近的经济型连锁酒店。"

	record, err := writer.WriteUserMemory(ctx, model.MemoryWriteRequest{
		UserID:     userID,
		SessionID:  "demo_session",
		MemoryType: "hotel_preference",
		Text:       memoryText,
		Source:     "user_message",
		Importance: 9,
		MetadataJSON: `{
		"constraints": [
			{
			"domain": "hotel",
			"exclude_keywords": ["青旅", "青年旅舍", "青年酒店", "青年旅馆", "多人间", "床位房", "hostel"],
			"prefer_keywords": ["地铁", "连锁", "经济型", "酒店", "速8", "布丁", "如家", "汉庭", "7天", "全季", "海友"],
			"strength": "hard",
			"reason": "用户不喜欢青年旅舍、多人间、床位房，偏好地铁附近的经济型连锁酒店"
			}
		]
		}`,
	})
	if err != nil {
		log.Fatal("write memory failed: ", err)
	}

	fmt.Println("memory written successfully")
	fmt.Println("id:", record.ID)
	fmt.Println("text:", record.Text)

	// Zilliz 写入后可能有短暂可见延迟，等一下再搜。
	time.Sleep(2 * time.Second)

	query := "帮我规划北京三日游，酒店尽量方便，不要太贵。"

	memories, err := retriever.RetrieveUserMemories(ctx, userID, query, 5)
	if err != nil {
		log.Fatal("retrieve memory failed: ", err)
	}

	fmt.Println()
	fmt.Println("retrieved memories:")
	for i, memory := range memories {
		fmt.Printf("%d. score=%.4f type=%s importance=%d\n", i+1, memory.Score, memory.MemoryType, memory.Importance)
		fmt.Println("   text:", memory.Text)
	}
}
