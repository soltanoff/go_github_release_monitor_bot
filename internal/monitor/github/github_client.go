package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"

	"github.com/soltanoff/go_github_release_monitor_bot/pkg/logs"
)

const (
	releaseURLMask string = "https://api.github.com/repos/%s/releases/latest"
	tagsURLMask    string = "https://api.github.com/repos/%s/git/refs/tags"
	releaseTagMask string = "https://github.com/%s/releases/tag/%s"
)

var tagURIPattern = regexp.MustCompile(`refs/tags/([\w\d\-\.]+)`)

type Client struct {
	httpClient *http.Client
}

func New() *Client {
	return &Client{httpClient: &http.Client{}}
}

func (c *Client) GetLatestTagFromReleaseURI(
	ctx context.Context,
	repoShortName string,
) (releaseInfo ReleaseInfo, err error) {
	resp, err := c.makeGetHTTPRequest(ctx, c.httpClient, fmt.Sprintf(releaseURLMask, repoShortName))
	if err != nil {
		logs.LogError("[GITHUB-CLIENT] Latest tag for release uri request failed: %s", err)
		return releaseInfo, nil
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&releaseInfo); err != nil {
		logs.LogError("[GITHUB-CLIENT] Latest tag for release uri body decoder failed: %s", err)
		return releaseInfo, fmt.Errorf("[GITHUB-CLIENT] latest tag for release uri body decoder failed: %w", err)
	}

	return releaseInfo, nil
}

func (c *Client) GetLatestTagFromTagURI(
	ctx context.Context,
	repoShortName string,
) (releaseInfo ReleaseInfo, err error) {
	resp, err := c.makeGetHTTPRequest(ctx, c.httpClient, fmt.Sprintf(tagsURLMask, repoShortName))
	if err != nil {
		logs.LogError("[GITHUB-CLIENT] Latest tag for tag uri request failed: %s", err)
		return releaseInfo, nil
	}
	defer resp.Body.Close()

	var tagInfoList []TagInfo
	if err := json.NewDecoder(resp.Body).Decode(&tagInfoList); err != nil {
		logs.LogError("[GITHUB-CLIENT] Latest tag for tag uri read body failed: %s", err)
		return releaseInfo, fmt.Errorf("[GITHUB-CLIENT] latest tag for tag uri read body failed: %w", err)
	}

	if len(tagInfoList) == 0 {
		logs.LogWarn("[GITHUB-CLIENT] Latest tag for tag uri request is empty")
		return releaseInfo, nil
	}

	sort.Slice(tagInfoList, func(i, j int) bool {
		return tagInfoList[i].Ref > tagInfoList[j].Ref
	})

	tagInfo := tagInfoList[0]
	releaseInfo = ReleaseInfo{
		TagName:   tagURIPattern.FindStringSubmatch(tagInfo.Ref)[1],
		SourceURL: fmt.Sprintf(releaseTagMask, repoShortName, releaseInfo.TagName),
	}

	return releaseInfo, nil
}

func (c *Client) makeGetHTTPRequest(ctx context.Context, httpClient *http.Client, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("[GITHUB-CLIENT] get-request failed: %w", err)
	}

	return httpClient.Do(req)
}
