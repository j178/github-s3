package main

import (
	"fmt"
	"os"

	githubs3 "github.com/j178/github-s3"
	"github.com/j178/kooky"

	_ "github.com/j178/kooky/browser/chrome"
	_ "github.com/j178/kooky/browser/edge"
	_ "github.com/j178/kooky/browser/firefox"
	_ "github.com/j178/kooky/browser/safari"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: github-s3-auto <file-path>")
		os.Exit(1)
	}

	cookies := kooky.ReadCookies(kooky.Domain("github.com"), kooky.Name("user_session"))
	if len(cookies) == 0 {
		fmt.Println("No github cookies found")
		os.Exit(1)
	}

	gh := githubs3.New(cookies[0].Value)
	loc, err := gh.UploadFromPath(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(loc.GithubLink)
}
