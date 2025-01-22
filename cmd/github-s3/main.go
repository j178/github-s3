package main

import (
	"fmt"
	"os"

	githubs3 "github.com/j178/github-s3"
)

func main() {
	githubs3.Run(
		func() (githubs3.Credential, error) {
			session := os.Getenv("GITHUB_SESSION")
			deviceID := os.Getenv("GITHUB_DEVICE_ID")
			if session == "" || deviceID == "" {
				return githubs3.Credential{}, fmt.Errorf("GITHUB_SESSION and GITHUB_DEVICE_ID are required")
			}
			return githubs3.Credential{
				UserSession: session,
				DeviceID:    deviceID,
			}, nil
		},
	)
}
