package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/abdullayev4u/gc2/config"
	"github.com/abdullayev4u/gc2/tools"
)

func main() {
	args := os.Args

	if len(args) < 2 {
		tools.PrintHelp()
	}

	switch args[1] {
	case "help", "--help", "-h":
		tools.PrintHelp()
	case "version", "--version", "-v":
		printVersion()
	}

	args = args[1:]

	cmd, err := tools.ParseCommand(args)
	exit(err)

	err = tools.EnsureParent(cmd)
	exit(err)

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go tools.LoadIcons(cmd, wg)

	err = tools.GitClone(cmd)
	exit(err)

	err = tools.OpenEditor(cmd)
	exit(err)

	wg.Wait()

}

func exit(err error, code ...int) {
	if err == nil {
		return
	}
	fmt.Println(err.Error())

	c := 1
	if len(code) > 0 {
		c = code[0]
	}
	os.Exit(c)
}

func printVersion() {
	fmt.Printf("v%s", config.Version)
	os.Exit(0)
}
