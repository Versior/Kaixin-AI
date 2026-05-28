package repository

import (
	"sort"
	"strings"

	"github.com/basketikun/infinite-canvas/model"
)

type GenerationUserRank struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Tasks    int    `json:"tasks"`
	Images   int    `json:"images"`
}

type GenerationImageStatsResult struct {
	TotalImages   int                  `json:"totalImages"`
	TodayImages   int                  `json:"todayImages"`
	SuccessImages int                  `json:"successImages"`
	FailedImages  int                  `json:"failedImages"`
	UserRanks     []GenerationUserRank `json:"userRanks"`
}

func GenerationImageStats(todayPrefix string, rankLimit int) (GenerationImageStatsResult, error) {
	db, err := DB()
	if err != nil {
		return GenerationImageStatsResult{}, err
	}
	if rankLimit <= 0 {
		rankLimit = 10
	}
	var logs []model.GenerationLog
	if err := db.Model(&model.GenerationLog{}).Select("generation_logs.*, COALESCE(NULLIF(users.display_name, ''), users.username, '-') AS username").Joins("LEFT JOIN users ON users.id = generation_logs.user_id").Where("generation_logs.kind = ?", model.GenerationLogKindImage).Find(&logs).Error; err != nil {
		return GenerationImageStatsResult{}, err
	}
	stats := GenerationImageStatsResult{UserRanks: []GenerationUserRank{}}
	rankByUser := map[string]*GenerationUserRank{}
	for _, log := range logs {
		imageCount := len(log.Images)
		isFailed := log.Status == "failed" || log.Status == "error" || log.Status == "rate_limited"
		if isFailed {
			failedCount := imageCount
			if failedCount == 0 {
				failedCount = 1
			}
			stats.FailedImages += failedCount
		} else {
			stats.SuccessImages += imageCount
			stats.TotalImages += imageCount
			if todayPrefix != "" && strings.HasPrefix(log.CreatedAt, todayPrefix) {
				stats.TodayImages += imageCount
			}
		}
		if imageCount == 0 || isFailed {
			continue
		}
		rank := rankByUser[log.UserID]
		if rank == nil {
			rank = &GenerationUserRank{UserID: log.UserID, Username: log.Username}
			rankByUser[log.UserID] = rank
		}
		if rank.Username == "" || rank.Username == "-" {
			rank.Username = log.Username
		}
		rank.Tasks++
		rank.Images += imageCount
	}
	stats.UserRanks = make([]GenerationUserRank, 0, len(rankByUser))
	for _, rank := range rankByUser {
		stats.UserRanks = append(stats.UserRanks, *rank)
	}
	sort.SliceStable(stats.UserRanks, func(i, j int) bool {
		left := stats.UserRanks[i]
		right := stats.UserRanks[j]
		if left.Images != right.Images {
			return left.Images > right.Images
		}
		if left.Tasks != right.Tasks {
			return left.Tasks > right.Tasks
		}
		if left.Username != right.Username {
			return left.Username < right.Username
		}
		return left.UserID < right.UserID
	})
	if len(stats.UserRanks) > rankLimit {
		stats.UserRanks = stats.UserRanks[:rankLimit]
	}
	return stats, nil
}
