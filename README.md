# GitHub as a file server 

Abuse GitHub web interface [attachment feature](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/attaching-files) to upload a file, and use GitHub as a file server.

Especially useful for hosting image files, which can be used in markdown files.

## Usage

First, login to your GitHub account, and get cookie named `user_session` from GitHub web browser session.

```shell
go install github.com/j178/github-s3/cmd@latest
github-s3 <github-user-session> <path-to-file>
```
