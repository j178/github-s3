package github_s3

import (
	"flag"
	"fmt"
	"os"
)

var repo = flag.String("repo", "", "github repo name")

func Run(sessionGetter func() string) {
	flag.Parse()
	if len(flag.Args()) == 0 {
		fmt.Println("Usage: github-s3 [file-path]")
		os.Exit(1)
	}

	session := sessionGetter()
	if session == "" {
		fmt.Println("GITHUB_SESSION env var is required")
		os.Exit(1)
	}
	gh := New(session, *repo)

	for _, path := range flag.Args() {
		res, err := gh.UploadFromPath(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: "+err.Error())
			os.Exit(1)
		}
		fmt.Println(res.GithubLink)
	}
}
