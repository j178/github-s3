package github_s3

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/go-resty/resty/v2"
)

type GitHub struct {
	c                 *resty.Client
	authenticityToken string
	repositoryId      string
}

func NewGitHub(userSession string) *GitHub {
	c := resty.New()
	c.SetCookies([]*http.Cookie{
		{
			Name:  "user_session",
			Value: userSession,
		},
		{
			Name:  "__Host-user_session_same_site",
			Value: userSession,
		},
	})
	c.SetDebug(os.Getenv("DEBUG") == "1")
	c.SetRedirectPolicy(resty.NoRedirectPolicy())
	c.SetContentLength(true)
	c.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")
	c.SetHeader("Referer", "https://github.com/cli/cli/issues/7797")
	return &GitHub{
		c: c,
		// repositoryId doesn't matter, use cli/cli as default
		repositoryId: "212613049",
	}
}

var tokenPattern = regexp.MustCompile(`<file-attachment class="js-upload-markdown-image.*?<input type="hidden" value="([^{"]+?)" data-csrf="true"`)

func (g *GitHub) fetchAuthenticityToken() (string, error) {
	resp, err := g.c.R().Get("https://github.com/cli/cli/issues/new?assignees=&labels=bug&projects=&template=bug_report.md")
	if err != nil {
		return "", err
	}
	if !resp.IsSuccess() {
		return "", fmt.Errorf("failed to fetch authenticity token: %s", resp.Status())
	}
	body := resp.String()
	matches := tokenPattern.FindStringSubmatch(body)
	if len(matches) != 2 {
		return "", fmt.Errorf("authenticity token not found")
	}
	return matches[1], nil
}

type preUploadResult struct {
	UploadUrl                    string `json:"upload_url"`
	UploadAuthenticityToken      string `json:"upload_authenticity_token"`
	AssetUploadUrl               string `json:"asset_upload_url"`
	AssetUploadAuthenticityToken string `json:"asset_upload_authenticity_token"`
	Asset                        struct {
		ID           int    `json:"id"`
		Name         string `json:"name"`
		Size         int    `json:"size"`
		ContentType  string `json:"content_type"`
		Href         string `json:"href"`
		OriginalName string `json:"original_name"`
	} `json:"asset"`
	Form       map[string]string `json:"form"`
	Header     any               `json:"header"`
	SameOrigin bool              `json:"same_origin"`
}

func (g *GitHub) preUpload(name string, size int, contentType string) (*preUploadResult, error) {
	if g.authenticityToken == "" {
		token, err := g.fetchAuthenticityToken()
		if err != nil {
			return nil, err
		}
		g.authenticityToken = token
	}

	var result preUploadResult
	resp, err := g.c.R().
		SetMultipartFormData(map[string]string{
			"authenticity_token": g.authenticityToken,
			"repository_id":      g.repositoryId,
			"name":               name,
			"size":               strconv.Itoa(size),
			"content_type":       contentType,
		}).
		SetResult(&result).
		Post("https://github.com/upload/policies/assets")
	if err != nil {
		return nil, err
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("failed to pre-upload: %s\n%s", resp.Status(), resp.String())
	}
	return &result, nil
}

func (g *GitHub) postUpload(result *preUploadResult) error {
	resp, err := g.c.R().
		SetHeader("X-Requested-With", "XMLHttpRequest").
		SetMultipartFormData(map[string]string{"authenticity_token": result.AssetUploadAuthenticityToken}).
		Put("https://github.com" + result.AssetUploadUrl)
	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		return fmt.Errorf("failed to post upload: %s", resp.Status())
	}
	return nil
}

type UploadResult struct {
	// The URL of the uploaded files.
	GithubLink string
	// If the file is an image or video, the direct AWS link to the file. (After redirected from the GitHub link)
	// For other type of files, this field is empty.
	AwsLink string
}

func (g *GitHub) Upload(name string, size int, r io.Reader) (UploadResult, error) {
	contentType := mime.TypeByExtension(filepath.Ext(name))
	result, err := g.preUpload(name, size, contentType)
	if err != nil {
		return UploadResult{}, err
	}

	resp, err := g.c.R().
		SetHeader("X-Requested-With", "XMLHttpRequest").
		SetFormData(result.Form).
		SetFileReader("file", name, r).
		Post(result.UploadUrl)
	if err != nil {
		return UploadResult{}, err
	}
	if !resp.IsSuccess() {
		return UploadResult{}, fmt.Errorf("failed to upload image: %s", resp.Status())
	}
	loc := resp.Header().Get("Location")

	err = g.postUpload(result)
	if err != nil {
		return UploadResult{}, err
	}

	return UploadResult{
		GithubLink: result.Asset.Href,
		AwsLink:    loc,
	}, nil
}

func (g *GitHub) UploadFromPath(path string) (UploadResult, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return UploadResult{}, err
	}
	r, err := os.Open(path)
	if err != nil {
		return UploadResult{}, err
	}
	defer r.Close()
	return g.Upload(filepath.Base(path), int(stat.Size()), r)
}
