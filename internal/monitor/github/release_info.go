package github

type ReleaseInfo struct {
	TagName   string `json:"tag_name"`
	SourceURL string `json:"html_url"`
}

func (r *ReleaseInfo) IsZero() bool {
	return r.SourceURL == "" && r.TagName == ""
}
