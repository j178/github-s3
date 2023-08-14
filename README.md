# GitHub as a file server 

Abuse GitHub unpublicized [attachment](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/attaching-files) API to serve a file.

Especially useful for hosting image files that can be referenced in markdown files.

## Usage

1. Install
    ```shell
    brew install j178/tap/github-s3
    # or
    go install github.com/j178/github-s3/cmd/github-s3@latest
    ```
2. Login to your GitHub account, and obtain cookie named `user_session` from GitHub web browser session.
3. Run
    ```shell
    export GITHUB_SESSION=<github-user-session>   
    github-s3 <path-to-file>
    ```

If you don't want to obtain the cookie manually, you can use [github-s3-auto](./cmd/github-s3-auto) to automatically find the cookie from your web browser session.

```shell
go install github.com/j178/github-s3/cmd/github-s3-auto@latest

github-s3-auto <path-to-file>
```

### Use in GitHub Actions

```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Upload files to GitHub
        uses: j178/github-s3@v1
        with:
          files: <list-of-paths-to-file>
          github-session: ${{ secrets.GH_SESSION }}
      - name: Use uploaded files
        run: |
            # Use the uploaded files
        with:
         files: ${{ steps.upload.outputs.links }}
```

## Disclaimer

Please note that this project relies on an unpublicized API of GitHub, and its usage may be subject to changes in GitHub's policies or API. Use it responsibly and ensure compliance with GitHub's terms of service.
