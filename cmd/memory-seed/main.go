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

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	writer := tools.NewMemoryWriterTool()

	userID := "demo_user"
	sessionID := "seed_session"

	memories := []model.MemoryWriteRequest{
		{
			UserID:     userID,
			SessionID:  sessionID,
			MemoryType: "hotel_preference",
			Text:       "用户不喜欢青年旅舍、青年酒店、多人间、床位房；未来住宿推荐应优先避开青旅，优先选择地铁附近的经济型连锁酒店。",
			Source:     "seed",
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
		},
		{
			UserID:     userID,
			SessionID:  sessionID,
			MemoryType: "transport_preference",
			Text:       "用户默认更喜欢高铁或动车，不喜欢普快硬座；如果用户本轮明确说坐飞机，则以本轮明确需求为准。",
			Source:     "seed",
			Importance: 8,
			MetadataJSON: `{
				"constraints": [
					{
						"domain": "transport",
						"exclude_keywords": ["普快", "硬座"],
						"prefer_keywords": ["高铁", "动车", "二等座"],
						"strength": "hard",
						"reason": "用户默认更喜欢高铁或动车，不喜欢普快硬座"
					}
				]
			}`,
		},
		{
			UserID:     userID,
			SessionID:  sessionID,
			MemoryType: "attraction_preference",
			Text:       "用户偏好博物馆、展览馆、历史人文类景点，不喜欢体力消耗太大的爬山、徒步路线。",
			Source:     "seed",
			Importance: 7,
			MetadataJSON: `{
				"constraints": [
					{
						"domain": "attraction",
						"exclude_keywords": ["爬山", "登山", "徒步", "山湖田园"],
						"prefer_keywords": ["博物馆", "展览馆", "美术馆", "纪念馆", "历史", "人文", "古迹"],
						"avoid_tags": ["mountain", "hiking", "high_exertion"],
						"prefer_tags": ["museum", "exhibition", "art_gallery", "memorial", "culture", "indoor", "low_exertion"],
						"strength": "soft",
						"reason": "用户偏好人文类景点，不喜欢体力消耗太大的路线"
					}
				]
			}`,
		},
		{
			UserID:     userID,
			SessionID:  sessionID,
			MemoryType: "route_preference",
			Text:       "用户偏好轻松旅行，不希望每天安排太多景点，也不希望长距离奔波。",
			Source:     "seed",
			Importance: 7,
			MetadataJSON: `{
				"constraints": [
					{
						"domain": "route",
						"exclude_keywords": ["tight_schedule", "long_transfer", "too_many_attractions"],
						"prefer_keywords": ["relaxed", "slow_pace", "short_transfer"],
						"avoid_tags": ["tight_schedule", "long_transfer", "too_many_attractions"],
						"prefer_tags": ["relaxed", "slow_pace", "short_transfer"],
						"strength": "soft",
						"reason": "用户偏好轻松旅行，不希望每天安排太多景点或长距离通勤"
					}
				]
			}`,
		},
	}

	for _, req := range memories {
		record, err := writer.WriteUserMemory(ctx, req)
		if err != nil {
			log.Fatalf("write memory failed, type=%s, err=%v", req.MemoryType, err)
		}

		fmt.Printf("seeded memory: type=%s id=%s\n", record.MemoryType, record.ID)
	}

	fmt.Println("demo_user structured memories seeded successfully")
}
