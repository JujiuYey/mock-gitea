package server

import (
	"math/rand"
	"mockgitea/internal/config"
	"mockgitea/internal/data"
	"mockgitea/internal/models"
	"mockgitea/internal/utils"
	"sort"
	"strings"
	"time"
)

// buildCommits 为一个分支生成提交记录
func buildCommits(users []models.GiteaUser, anchorTime time.Time, repo models.GiteaRepo, branchName string) []models.GiteaCommit {

	// ===== 第 1 步：创建随机数生成器 =====
	// 用仓库名和分支名作为种子，保证每次运行提交记录一致
	rng := rand.New(rand.NewSource(utils.StableSeed("commits", repo.FullName, branchName)))

	// ===== 第 2 步：获取预定义的提交消息模板 =====
	messages := data.BuildCommitMessages()

	// ===== 第 3 步：构建加权日期偏移数组 =====
	// 用于模拟真实的工作日分布（工作日权重高，周末权重低）
	weightedDays := buildWeightedDayOffsets(anchorTime)

	// ===== 第 4 步：创建空提交切片 =====
	// 每个分支固定生成 commitsPerBranch (200) 个提交
	commits := make([]models.GiteaCommit, 0, config.CommitsPerBranch)

	// ===== 第 5 步：循环生成 200 个提交 =====
	for idx := 0; idx < config.CommitsPerBranch; idx++ {

		// 5.1 随机选择作者（从 10 个用户中选）
		authorIndex := rng.Intn(len(users))

		// 5.2 随机选择提交者
		// 82% 的情况：提交者 = 作者
		// 18% 的情况：提交者 = 随机用户（模拟代码审查场景）
		committerIndex := authorIndex
		if rng.Intn(100) < 18 {
			committerIndex = rng.Intn(len(users))
		}

		// 5.3 随机选择提交日期（基于加权分布）
		// 权重：周一/周五 = 2, 周二三四 = 3, 周末 = 1
		dayOffset := weightedDays[rng.Intn(len(weightedDays))]
		timestamp := randomTimestampForOffset(anchorTime, rng, dayOffset)

		// 5.4 随机选择提交消息
		// 从预定义的消息池中循环选取
		message := messages[(idx+rng.Intn(len(messages)))%len(messages)]

		// 5.5 根据分支类型调整提交消息
		// 如果是 hotfix 分支但消息不是 fix 开头，强制改成 fix
		if strings.HasPrefix(branchName, "hotfix/") && !strings.HasPrefix(message, "fix") {
			message = "fix(hotfix): stabilize production incident handling"
		}

		// 如果是 release 分支但消息是 feat 开头，改成 chore
		if strings.HasPrefix(branchName, "release/") && strings.HasPrefix(message, "feat") {
			message = "chore(release): prepare release candidate and bump version metadata"
		}

		// 5.6 根据消息类型生成代码变更统计
		// 不同类型的提交，代码行数不同（如 docs 改动少，refactor 改动大）
		additions, deletions := randomStats(rng, message)

		// 5.7 生成 SHA（提交 ID）
		sha := utils.ShortSHA(repo.FullName, branchName, idx, timestamp.UnixNano(), message)

		// 5.8 获取作者和提交者用户信息
		author := users[authorIndex]
		committer := users[committerIndex]

		// 5.9 组装提交结构体
		commit := models.GiteaCommit{
			SHA: sha,
			Commit: models.GiteaCommitInfo{
				Message: message,
				Author: models.GiteaCommitUser{
					Name:  author.FullName,
					Email: author.Email,
					Date:  timestamp,
				},
				Committer: models.GiteaCommitUser{
					Name:  committer.FullName,
					Email: committer.Email,
					// 提交时间比作者时间晚 0~45 分钟（模拟审查延迟）
					Date: timestamp.Add(time.Duration(rng.Intn(45)) * time.Minute),
				},
			},
			Stats: &models.GiteaStats{
				Additions: additions,             // 新增行数
				Deletions: deletions,             // 删除行数
				Total:     additions + deletions, // 总变更行数
			},
			Author: &users[authorIndex],
		}

		// 5.10 添加到提交切片
		commits = append(commits, commit)
	}

	// ===== 第 6 步：按提交时间倒序排序 =====
	// 最新的提交排在前面
	sort.Slice(commits, func(i, j int) bool {
		left := commits[i].Commit.Author.Date
		right := commits[j].Commit.Author.Date

		// 如果时间相同，按 SHA 字典序（保证顺序一致）
		if left.Equal(right) {
			return commits[i].SHA > commits[j].SHA
		}

		// 否则按时间倒序（新的在前）
		return left.After(right)
	})

	// ===== 第 7 步：返回提交切片 =====
	return commits
}

func buildWeightedDayOffsets(anchor time.Time) []int {
	offsets := make([]int, 0, config.MaxRecentDays*3)
	dayStart := time.Date(anchor.Year(), anchor.Month(), anchor.Day(), 0, 0, 0, 0, time.UTC)
	for offset := 0; offset < config.MaxRecentDays; offset++ {
		day := dayStart.AddDate(0, 0, -offset)
		weight := 3
		switch day.Weekday() {
		case time.Saturday:
			weight = 1
		case time.Sunday:
			weight = 1
		case time.Monday, time.Friday:
			weight = 2
		}
		for repeat := 0; repeat < weight; repeat++ {
			offsets = append(offsets, offset)
		}
	}
	return offsets
}

func randomTimestampForOffset(anchorTime time.Time, rng *rand.Rand, dayOffset int) time.Time {
	dayStart := time.Date(anchorTime.Year(), anchorTime.Month(), anchorTime.Day(), 0, 0, 0, 0, time.UTC).
		AddDate(0, 0, -dayOffset)

	hour := 9 + rng.Intn(11)
	minute := rng.Intn(60)
	second := rng.Intn(60)
	return dayStart.Add(time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute + time.Duration(second)*time.Second)
}

func randomStats(rng *rand.Rand, message string) (int, int) {
	switch {
	case strings.HasPrefix(message, "docs"):
		return 4 + rng.Intn(40), 0 + rng.Intn(8)
	case strings.HasPrefix(message, "test"):
		return 10 + rng.Intn(80), 2 + rng.Intn(24)
	case strings.HasPrefix(message, "refactor"):
		return 30 + rng.Intn(160), 20 + rng.Intn(120)
	case strings.HasPrefix(message, "perf"):
		return 18 + rng.Intn(90), 10 + rng.Intn(40)
	case strings.HasPrefix(message, "ci"), strings.HasPrefix(message, "build"), strings.HasPrefix(message, "chore"):
		return 6 + rng.Intn(60), 2 + rng.Intn(20)
	case strings.HasPrefix(message, "fix"):
		return 8 + rng.Intn(70), 3 + rng.Intn(30)
	default:
		return 20 + rng.Intn(160), 6 + rng.Intn(70)
	}
}
