package github_s3

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-resty/resty/v2"
)

type GitHub struct {
	c      *resty.Client
	repo   string
	repoId int
}

type Credential struct {
	UserSession string
	DeviceID    string
}

func New(cred Credential, repo string) *GitHub {
	g := &GitHub{}
	if repo == "" {
		g.repo = "cli/cli"
		g.repoId = 212613049
	} else {
		g.repo = repo
	}

	c := resty.New()
	u, _ := url.Parse("https://github.com")
	// Set cookies to jar avoid leaking to other sites
	c.GetClient().Jar.SetCookies(u, []*http.Cookie{
		{
			Name:     "user_session",
			Value:    cred.UserSession,
			SameSite: http.SameSiteLaxMode,
			Domain:   "github.com",
		},
		{
			Name:     "__Host-user_session_same_site",
			Value:    cred.UserSession,
			SameSite: http.SameSiteLaxMode,
			Domain:   "github.com",
		},
		{
			Name:     "_device_id",
			Value:    cred.DeviceID,
			SameSite: http.SameSiteLaxMode,
			Domain:   "github.com",
		},
	})
	c.SetDebug(os.Getenv("DEBUG") == "1")
	c.SetRedirectPolicy(resty.NoRedirectPolicy())
	c.SetContentLength(true)
	c.SetHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	c.SetHeader("X-Requested-With", "XMLHttpRequest")
	g.c = c

	return g
}

func (g *GitHub) getRepoId() (int, error) {
	var result struct {
		ID int `json:"id"`
	}
	resp, err := g.c.R().SetResult(&result).Get("https://api.github.com/repos/" + g.repo)
	if err != nil {
		return 0, err
	}
	if !resp.IsSuccess() {
		return 0, fmt.Errorf("failed to get repo id: %s", resp.Status())
	}
	return result.ID, nil
}

type uploadPolicy struct {
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

func (g *GitHub) getPolicy(name string, size int, contentType string) (*uploadPolicy, error) {
	if g.repoId == 0 {
		repoId, err := g.getRepoId()
		if err != nil {
			return nil, err
		}
		g.repoId = repoId
	}

	var result uploadPolicy
	resp, err := g.c.R().
		SetMultipartFormData(map[string]string{
			"repository_id": strconv.Itoa(g.repoId),
			"name":          name,
			"size":          strconv.Itoa(size),
			"content_type":  contentType,
		}).
		SetHeader("Github-Verified-Fetch", "true").
		SetHeader("Origin", "https://github.com").
		SetResult(&result).
		Post("https://github.com/upload/policies/assets")
	if err != nil {
		return nil, err
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("failed to get policy: %s\n%s", resp.Status(), resp.String())
	}
	return &result, nil
}

func (g *GitHub) markUploadComplete(result *uploadPolicy) error {
	resp, err := g.c.R().
		SetMultipartFormData(map[string]string{"authenticity_token": result.AssetUploadAuthenticityToken}).
		Put("https://github.com" + result.AssetUploadUrl)
	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		return fmt.Errorf("failed to mark upload complete: %s", resp.Status())
	}
	return nil
}

type UploadResult struct {
	// The URL of the uploaded files.
	GithubLink string
	// If the file is an image or video, the direct AWS link to the file (After redirected from the GitHub link).
	// For other type of files, this field is empty.
	AwsLink string
}

func (g *GitHub) Upload(name string, size int, r io.Reader) (UploadResult, error) {
	ext := filepath.Ext(name)
	contentType := ""
	if ext == ".log" {
		contentType = "text/x-log"
	} else {
		contentType = mime.TypeByExtension(ext)
	}
	policy, err := g.getPolicy(name, size, contentType)
	if err != nil {
		return UploadResult{}, err
	}

	resp, err := g.c.R().
		SetFormData(policy.Form).
		SetFileReader("file", name, r).
		Post(policy.UploadUrl)
	if err != nil {
		return UploadResult{}, err
	}
	if !resp.IsSuccess() {
		return UploadResult{}, fmt.Errorf("failed to upload image: %s", resp.Status())
	}
	loc := resp.Header().Get("Location")

	err = g.markUploadComplete(policy)
	if err != nil {
		return UploadResult{}, err
	}

	return UploadResult{
		GithubLink: policy.Asset.Href,
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
