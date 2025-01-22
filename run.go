package github_s3

import (
	"flag"
	"fmt"
	"os"
)

var repo = flag.String("repo", "", "github repo name")

func Run(sessionGetter func() (Credential, error)) {
	flag.Parse()
	if len(flag.Args()) == 0 {
		fmt.Println("Usage: github-s3 [file-path]")
		os.Exit(1)
	}

	session, err := sessionGetter()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error: "+err.Error())
		os.Exit(1)
	}
	gh := New(session, *repo)

	for _, path := range flag.Args() {
		res, err := gh.UploadFromPath(path)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "Error: "+err.Error())
			os.Exit(1)
		}
		fmt.Println(res.GithubLink)
	}
}
