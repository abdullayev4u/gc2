package tools

import (
	"fmt"
	"os"
)

func PrintHelp() {

	fmt.Println("Usage: gc2 <repo-url-from-your-git>")
	fmt.Println("Example: gc2 https://github.com/abdullayev4u/gc2.git")

	os.Exit(0)
}
