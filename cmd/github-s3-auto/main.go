package main

import (
	"fmt"

	"github.com/j178/kooky"

	_ "github.com/j178/kooky/browser/chrome"

	githubs3 "github.com/j178/github-s3"
)

func main() {
	githubs3.Run(
		func() (githubs3.Credential, error) {
			cookies := kooky.ReadCookies(kooky.Domain("github.com"), kooky.Name("user_session"), kooky.Name("_device_id"))
			if len(cookies) < 2 {
				return githubs3.Credential{}, fmt.Errorf("user_session and _device_id cookies not found")
			}
			cred := githubs3.Credential{}
			for _, c := range cookies {
				switch c.Name {
				case "user_session":
					cred.UserSession = c.Value
				case "_device_id":
					cred.DeviceID = c.Value
				}
			}
			return cred, nil
		},
	)
}
