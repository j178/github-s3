package main

import (
	"flag"
	"fmt"
	"os"

	githubs3 "github.com/j178/github-s3"
)

var repo = flag.String("repo", "", "github repo name")

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: github-s3 <file-path>")
		os.Exit(1)
	}
	flag.Parse()
	if *repo == "" {
		*repo = "cli/cli"
	}

	session := os.Getenv("GITHUB_SESSION")
	if session == "" {
		fmt.Println("GITHUB_SESSION env var is required")
		os.Exit(1)
	}
	gh := githubs3.New(session, githubs3.WithRepo(*repo))

	for _, path := range os.Args[1:] {
		res, err := gh.UploadFromPath(path)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(res.GithubLink)
	}
}
