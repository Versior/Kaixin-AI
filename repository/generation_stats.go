package repository

import (
	"strings"

	"github.com/basketikun/infinite-canvas/model"
)

type GenerationUserRank struct {
	UserID    string `json:"userId"`
	Username  string `json:"username,omitempty"`
	AvatarURL string `json:"avatarUrl,omitempty"`
	Tasks     int    `json:"tasks"`
	Images    int    `json:"images"`
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
	type logWithUser struct {
		model.GenerationLog
		AvatarURL string `gorm:"column:avatar_url"`
	}
	var logs []logWithUser
	if err := db.Model(&model.GenerationLog{}).Select("generation_logs.*, COALESCE(NULLIF(users.display_name, ''), users.username, '-') AS username, users.avatar_url AS avatar_url").Joins("LEFT JOIN users ON users.id = generation_logs.user_id").Where("generation_logs.kind = ?", model.GenerationLogKindImage).Find(&logs).Error; err != nil {
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
			rank = &GenerationUserRank{UserID: log.UserID, Username: log.Username, AvatarURL: strings.TrimSpace(log.AvatarURL)}
			rankByUser[log.UserID] = rank
		}
		if rank.Username == "" || rank.Username == "-" {
			rank.Username = log.Username
		}
		if rank.AvatarURL == "" {
			rank.AvatarURL = strings.TrimSpace(log.AvatarURL)
		}
		rank.Tasks++
		rank.Images += imageCount
	}
	stats.UserRanks = make([]GenerationUserRank, 0, len(rankByUser))
	for _, rank := range rankByUser {
		stats.UserRanks = append(stats.UserRanks, *rank)
	}
	for i := 0; i < len(stats.UserRanks); i++ {
		for j := i + 1; j < len(stats.UserRanks); j++ {
			if stats.UserRanks[j].Images > stats.UserRanks[i].Images || (stats.UserRanks[j].Images == stats.UserRanks[i].Images && stats.UserRanks[j].Tasks > stats.UserRanks[i].Tasks) {
				stats.UserRanks[i], stats.UserRanks[j] = stats.UserRanks[j], stats.UserRanks[i]
			}
		}
	}
	if len(stats.UserRanks) > rankLimit {
		stats.UserRanks = stats.UserRanks[:rankLimit]
	}
	return stats, nil
}
