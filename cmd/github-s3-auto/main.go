package main

import (
	"github.com/j178/kooky"

	githubs3 "github.com/j178/github-s3"
)

func main() {
	githubs3.Run(
		func() string {
			cookies := kooky.ReadCookies(kooky.Domain("github.com"), kooky.Name("user_session"))
			if len(cookies) == 0 {
				return ""
			}
			return cookies[0].Value
		},
	)
}
