package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/j178/kooky"

	githubs3 "github.com/j178/github-s3"
)

var repo = flag.String("repo", "", "github repo name")

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: github-s3 <file-path>")
		os.Exit(1)
	}
	flag.Parse()

	cookies := kooky.ReadCookies(kooky.Domain("github.com"), kooky.Name("user_session"))
	if len(cookies) == 0 {
		fmt.Println("No github cookies found")
		os.Exit(1)
	}

	gh := githubs3.New(cookies[0].Value, *repo)

	for _, path := range flag.Args() {
		res, err := gh.UploadFromPath(path)
		if err != nil {
			fmt.Println("Error: " + err.Error())
		}
		fmt.Println(res.GithubLink)
	}
}
