package main

import (
	"bufio"
	"fmt"
	"os"
	"ariadne/internal/parser"
	"ariadne/internal/tools"
)

func main() {
	fmt.Println("Ariadne 启动成功")
	fmt.Println("请输入你的旅行需求：")

	scanner := bufio.NewScanner(os.Stdin)

	if scanner.Scan() {
		input := scanner.Text()
		tripRequest := parser.ParseTripRequest(input)
		fmt.Printf("当前旅行请求对象：%+v\n", tripRequest)
		transportPlans := tools.GenerateTransportPlans(tripRequest)

		fmt.Println("交通方案：")
		for _, plan := range transportPlans {
			fmt.Printf("- %s：%s，预估价格：%d 元\n", plan.Method, plan.Description, plan.Price)
			fmt.Println("  预订/查询链接：", plan.BookingLink)
		}
	}
}
	