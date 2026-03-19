package server

import (
	"fmt"
	"mockgitea/internal/config"
	"mockgitea/internal/models"
	"mockgitea/internal/utils"

	"github.com/gofiber/fiber/v2"
)

func (s *MockServer) HandleUser(c *fiber.Ctx) error {
	return c.JSON(s.CurrentUser)
}

func (s *MockServer) HandleRepoSearch(c *fiber.Ctx) error {

	page := c.QueryInt("page", config.DefaultPage)
	limit := c.QueryInt("limit", config.DefaultLimit)
	ReposPage := utils.Paginate(s.Repos, page, limit)

	data := make([]models.GiteaRepo, 0, len(ReposPage))
	for _, repo := range ReposPage {
		data = append(data, repo.Repo)
	}

	return c.JSON(fiber.Map{
		"data": data,
		"ok":   true,
	})
}

func (s *MockServer) HandleBranches(c *fiber.Ctx) error {
	owner := c.Params("owner")
	repoName := c.Params("repo")

	repo, ok := s.findRepo(owner, repoName)
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Repository not found"})
	}

	page := c.QueryInt("page", config.DefaultPage)
	limit := c.QueryInt("limit", config.DefaultLimit)
	branchesPage := utils.Paginate(repo.Branches, page, limit)

	data := make([]models.GiteaBranch, 0, len(branchesPage))
	for _, branch := range branchesPage {
		data = append(data, branch.Branch)
	}
	return c.JSON(data)
}

func (s *MockServer) HandleCommits(c *fiber.Ctx) error {
	owner := c.Params("owner")
	repoName := c.Params("repo")

	repo, ok := s.findRepo(owner, repoName)
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Repository not found"})
	}

	sha := c.Query("sha")
	if sha == "" {
		sha = repo.Repo.DefaultBranch
	}

	branch, ok := repo.BranchByID[sha]
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": fmt.Sprintf("branch %q not found", sha)})
	}

	page := c.QueryInt("page", config.DefaultPage)
	limit := c.QueryInt("limit", config.DefaultLimit)
	return c.JSON(utils.Paginate(branch.Commits, page, limit))
}

func (s *MockServer) HandleSingleCommit(c *fiber.Ctx) error {
	owner := c.Params("owner")
	repoName := c.Params("repo")
	sha := c.Params("sha")

	if _, ok := s.findRepo(owner, repoName); !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Repository not found"})
	}

	if commit, ok := s.CommitIndex[sha]; ok {
		return c.JSON(commit)
	}

	return c.JSON(s.syntheticCommit(sha))
}
