package main

import (
	"fmt"
	"os"

	githubs3 "github.com/j178/github-s3"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: github-s3 <github-user-session> <file-path>")
		os.Exit(1)
	}

	gh := githubs3.NewGitHub(os.Args[1])
	loc, err := gh.UploadFromPath(os.Args[2])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(loc.GithubLink)
}
