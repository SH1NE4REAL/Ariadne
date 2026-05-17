package main

import (
	"fmt"
	"net/http"

	"ariadne/internal/handler"
)

func main() {
	http.HandleFunc("/api/trip/plan", handler.PlanTripHandler)

	fmt.Println("Ariadne HTTP 服务启动成功")
	fmt.Println("监听地址：http://localhost:8080")
	fmt.Println("接口地址：POST http://localhost:8080/api/trip/plan")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("服务启动失败：", err)
	}
}