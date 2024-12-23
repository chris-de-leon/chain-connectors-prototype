package gh

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const (
	GITHUB_API_URL = "https://api.github.com"
	GITHUB_URL     = "https://github.com"
)

type (
	ReleaseAsset struct {
		Name string `json:"name"`
	}

	Release struct {
		Assets []ReleaseAsset `json:"assets"`
	}

	Repository struct {
		Owner string
		Name  string
	}

	GithubApi struct {
		HttpClient *http.Client
		Repo       *Repository
	}
)

func NewClient(repo Repository) *GithubApi {
	return &GithubApi{HttpClient: &http.Client{}, Repo: &repo}
}

func NewRepository(owner string, name string) Repository {
	return Repository{Owner: owner, Name: name}
}

func (api *GithubApi) handleHttpError(res *http.Response) error {
	data := map[string]any{}
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return err
	}

	msg, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}

	return errors.New(string(msg))
}

func (api *GithubApi) get(ctx context.Context, baseUrl string, endpoint string) (*http.Response, error) {
	ept, _ := strings.CutPrefix(endpoint, "/")
	url := fmt.Sprintf("%s/%s", baseUrl, ept)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	} else {
		req.Header.Set("Accept", "application/vnd.github+json")
	}

	res, err := api.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, api.handleHttpError(res)
	} else {
		return res, nil
	}
}

func (api *GithubApi) GetReleaseByTag(ctx context.Context, tag string) (*Release, error) {
	res, err := api.get(
		ctx,
		GITHUB_API_URL,
		fmt.Sprintf(
			"/repos/%s/%s/releases/tags/%s",
			api.Repo.Owner,
			api.Repo.Name,
			tag,
		),
	)

	if err != nil {
		return nil, err
	} else {
		defer res.Body.Close()
	}

	var data *Release
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	} else {
		return data, nil
	}
}

func (api *GithubApi) DownloadReleaseAsset(ctx context.Context, tag string, name string) (*http.Response, error) {
	return api.get(
		ctx,
		GITHUB_URL,
		fmt.Sprintf(
			"/%s/%s/releases/download/%s/%s",
			api.Repo.Owner,
			api.Repo.Name,
			tag,
			name,
		),
	)
}
