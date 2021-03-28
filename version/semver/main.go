package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/coreos/go-semver/semver"
)

func main() {
	out, err := exec.Command("git", "branch", "--show-current").Output()
	if err != nil {
		panic(err)
	}
	branch := strings.TrimSpace(string(out))
	if branch != "main" {
		panic(fmt.Errorf(`error: must be in "master" branch, current branch: %q`, branch))
	}

	out, err = exec.Command("git", "describe").Output()
	if err != nil {
		panic(err)
	}

	version := strings.TrimPrefix(strings.TrimSpace(string(out)), "v")
	v := semver.New(version)
	v.BumpPatch()
	reader := bufio.NewReader(os.Stdin)
	if _, err = fmt.Fprintf(os.Stderr, "Enter Release Version: [v%v] ", v); err != nil {
		panic(err)
	}

	text, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	if strings.HasPrefix(text, "v") {
		text = text[1:]
		v = semver.New(strings.TrimSpace(text))
	}

	if _, err = fmt.Fprintf(os.Stderr, "Using Version: v%v\n", v); err != nil {
		panic(err)
	}
	fmt.Printf("v%v", v)
}
