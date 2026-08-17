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
	case "config":
		runConfig(args[2:])
	case "where":
		runWhere(args[2:])
	}

	args = args[1:]

	cmd, err := prepare(args)
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

// prepare parses argv, loads the config and works out the destination.
func prepare(args []string) (*tools.Gc2Cmd, error) {
	cmd, err := tools.ParseCommand(args)
	if err != nil {
		return nil, err
	}

	file, err := config.Load()
	if err != nil {
		return nil, err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine user home directory: %w", err)
	}

	tools.ApplyConfig(cmd, file, home)

	return cmd, nil
}

// runWhere prints where a repository would land, without cloning it. Mapping
// rules are invisible until they surprise you; this is how you check one.
func runWhere(args []string) {
	cmd, err := prepare(args)
	exit(err)

	fmt.Println(cmd.DestFullPath)

	if cmd.MatchedRule != nil {
		fmt.Printf("  matched rule: %q -> %q\n", cmd.MatchedRule.Pattern, cmd.MatchedRule.Target)
	} else {
		fmt.Printf("  no rule matched; mirroring %s\n", tools.GroupPath(cmd.Repo_groups))
	}

	os.Exit(0)
}

func runConfig(args []string) {
	path, err := config.Path()
	exit(err)

	sub := ""
	if len(args) > 0 {
		sub = args[0]
	}

	switch sub {
	case "path":
		fmt.Println(path)

	case "init":
		force := len(args) > 1 && args[1] == "--force"

		if config.Exists(path) && !force {
			fmt.Printf("%s already exists; pass --force to overwrite it\n", path)
			os.Exit(1)
		}

		exit(config.WriteTemplate(path))
		fmt.Printf("wrote %s\n", path)

	default:
		fmt.Println("Usage: gc2 config path")
		fmt.Println("       gc2 config init [--force]")
		os.Exit(1)
	}

	os.Exit(0)
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
	fmt.Printf("\nv%s\n\n", config.Version)
	os.Exit(0)
}
