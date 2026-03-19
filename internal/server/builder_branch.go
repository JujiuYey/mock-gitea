package server

import (
	"math/rand"
	"mockgitea/internal/models"
	"mockgitea/internal/utils"
	"sort"
	"time"
)

// buildBranches 为一个仓库创建多个分支
func buildBranches(users []models.GiteaUser, anchorTime time.Time, repo models.GiteaRepo, domain string) []models.MockBranch {

	// ===== 第 1 步：获取该仓库的分支名列表 =====
	// 根据领域和仓库信息生成对应的分支名（如 main, develop, feature/* 等）
	branchNames := buildBranchNames(repo, domain)

	// ===== 第 2 步：创建空切片 =====
	branches := make([]models.MockBranch, 0, len(branchNames))

	// ===== 第 3 步：遍历每个分支名，创建分支和提交 =====
	for _, branchName := range branchNames {

		// 3.1 为这个分支生成提交记录（200 个提交）
		commits := buildCommits(users, anchorTime, repo, branchName)

		// 3.2 创建 mockBranch 结构体
		branches = append(branches, models.MockBranch{
			Branch: models.GiteaBranch{
				Name: branchName,
				Commit: models.GiteaBranchCommit{
					// 分支的最新提交 SHA（取第一个，因为提交是按时间倒序的）
					ID: commits[0].SHA,
				},
			},
			// 这个分支的所有提交
			Commits: commits,
		})
	}

	// ===== 第 4 步：按分支名排序 =====
	// 排序规则：
	//   1. main 分支排第一
	//   2. develop 分支排第二
	//   3. 其他按字母顺序
	sort.Slice(branches, func(i, j int) bool {
		left, right := branches[i].Branch.Name, branches[j].Branch.Name

		// main 永远排最前
		if left == repo.DefaultBranch { // "main"
			return true
		}
		if right == repo.DefaultBranch {
			return false
		}

		// develop 排第二
		if left == "develop" {
			return true
		}
		if right == "develop" {
			return false
		}

		// 其他按字母顺序
		return left < right
	})

	// ===== 第 5 步：返回分支切片 =====
	return branches
}

// buildBranchNames 为一个仓库生成分支名列表
// 参数：
//   - repo: 仓库信息（获取 FullName 用于生成稳定种子）
//   - domain: 领域类型，决定 feature 分支的类型
//
// 返回：分支名切片
func buildBranchNames(repo models.GiteaRepo, domain string) []string {

	// ===== 第 1 步：创建随机数生成器 =====
	// 用仓库全名和领域作为种子，保证每次运行分支名一致
	rng := rand.New(rand.NewSource(utils.StableSeed("branches", repo.FullName, domain)))

	// ===== 第 2 步：定义各类分支名 =====

	// 基础分支：每个仓库都有
	base := []string{"main", "develop"}

	// 发布分支：随机选 1 个
	releases := []string{"release/v1.0", "release/v1.1", "release/v1.2", "release/v2.0"}

	// 功能分支：根据领域不同，选用不同的 feature 分支
	featuresByDomain := map[string][]string{
		"frontend": {"feature/ui-refresh", "feature/dashboard-filter", "feature/export-center", "feature/theme-token"},
		"backend":  {"feature/query-cache", "feature/batch-sync", "feature/audit-log", "feature/webhook-retry"},
		"mobile":   {"feature/offline-cache", "feature/push-center", "feature/profile-redesign", "feature/perf-monitor"},
		"infra":    {"feature/canary-release", "feature/pod-budget", "feature/secret-rotation", "feature/cluster-cost"},
		"tooling":  {"feature/schema-lint", "feature/generator-v2", "feature/template-market", "feature/dev-cli"},
		"data":     {"feature/metric-drilldown", "feature/etl-backfill", "feature/cohort-export", "feature/data-guard"},
		"ai":       {"feature/agent-memory", "feature/prompt-playground", "feature/rag-cache", "feature/model-router"},
	}

	// 热修复分支：全部可选
	hotfixes := []string{"hotfix/login-timeout", "hotfix/cron-retry", "hotfix/date-range", "hotfix/cache-ttl"}

	// 杂项分支：全部可选
	choreBranches := []string{"chore/upgrade-deps", "perf/query-plan", "test/e2e-stabilize"}

	// ===== 第 3 步：构建分支名列表 =====

	// 3.1 首先添加基础分支（main, develop）
	names := append([]string{}, base...)

	// 3.2 随机添加 1 个发布分支
	names = append(names, releases[rng.Intn(len(releases))])

	// 3.3 获取当前领域的功能分支池
	featurePool := append([]string{}, featuresByDomain[domain]...)

	// 3.4 随机打乱功能分支的顺序
	rng.Shuffle(len(featurePool), func(i, j int) {
		featurePool[i], featurePool[j] = featurePool[j], featurePool[i]
	})

	// 3.5 随机决定分支总数（5~8 个）
	targetCount := 5 + rng.Intn(4) // 5, 6, 7, 或 8

	// 3.6 依次添加功能分支，直到达到目标数量
	for _, feature := range featurePool {
		if len(names) >= targetCount {
			break
		}
		names = append(names, feature)
	}

	// 3.7 添加热修复分支（如果还有空位）
	for _, branch := range hotfixes {
		if len(names) >= targetCount {
			break
		}
		names = append(names, branch)
	}

	// 3.8 添加杂项分支（如果还有空位）
	for _, branch := range choreBranches {
		if len(names) >= targetCount {
			break
		}
		names = append(names, branch)
	}

	// ===== 第 4 步：返回分支名列表 =====
	return names
}
