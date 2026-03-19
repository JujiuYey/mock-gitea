package models

import "time"

type GiteaRepo struct {
	ID              int64     `json:"id"`
	Owner           GiteaUser `json:"owner"`
	Name            string    `json:"name"`
	FullName        string    `json:"full_name"`
	Description     string    `json:"description"`
	DefaultBranch   string    `json:"default_branch"`
	StarsCount      int       `json:"stars_count"`
	ForksCount      int       `json:"forks_count"`
	OpenIssuesCount int       `json:"open_issues_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type GiteaUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type GiteaCommit struct {
	SHA    string          `json:"sha"`
	Commit GiteaCommitInfo `json:"commit"`
	Stats  *GiteaStats     `json:"stats"`
	Author *GiteaUser      `json:"author"`
}
type GiteaCommitInfo struct {
	Message   string          `json:"message"`
	Author    GiteaCommitUser `json:"author"`
	Committer GiteaCommitUser `json:"committer"`
}

type GiteaCommitUser struct {
	Name  string    `json:"name"`
	Email string    `json:"email"`
	Date  time.Time `json:"date"`
}

type GiteaStats struct {
	Additions int `json:"additions"`
	Deletions int `json:"deletions"`
	Total     int `json:"total"`
}

type GiteaBranch struct {
	Name   string            `json:"name"`
	Commit GiteaBranchCommit `json:"commit"`
}

type GiteaBranchCommit struct {
	ID string `json:"id"`
}
