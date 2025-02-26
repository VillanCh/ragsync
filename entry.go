package ragsync

import (
	"fmt"

	"github.com/VillanCh/ragsync/cmd/commands"
)

func Help() {
	fmt.Println("ragsync is a tool to sync files to Aliyun Bailian RAG")
	fmt.Println("Available commands:")

	// 遍历所有可用命令
	for _, cmd := range commands.GetCommands() {
		fmt.Printf("  %s - %s\n", cmd.Name, cmd.Usage)
	}
}
