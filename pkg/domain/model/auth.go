package model

import (
	"net/http"
	"strconv"
)

type GitHubWebhookAuth struct {
	// Example
	/*
		X-GitHub-Delivery: 2b844992-3689-11ef-89d9-8126a7fa0a02
		X-GitHub-Event: push
		X-GitHub-Hook-ID: 487296180
		X-GitHub-Hook-Installation-Target-ID: 933446
		X-GitHub-Hook-Installation-Target-Type: integration
	*/

	Delivery    string `json:"delivery"`
	Event       string `json:"event"`
	HookID      int    `json:"hook_id"`
	InstallID   int    `json:"install_id"`
	InstallType string `json:"install_type"`
}

func NewGitHubWebhookAuth(r *http.Request) *GitHubWebhookAuth {
	auth := GitHubWebhookAuth{
		Delivery:    r.Header.Get("X-GitHub-Delivery"),
		Event:       r.Header.Get("X-GitHub-Event"),
		InstallType: r.Header.Get("X-GitHub-Hook-Installation-Target-Type"),
	}

	// Parse integer
	if v, err := strconv.Atoi(r.Header.Get("X-GitHub-Hook-ID")); err == nil {
		auth.HookID = v
	}
	if v, err := strconv.Atoi(r.Header.Get("X-GitHub-Hook-Installation-Target-ID")); err == nil {
		auth.InstallID = v
	}

	return &auth
}
