package listing

import "fresh/internal/domain"

type ActivityFinalizeResult struct {
	Completed bool
	Info      InfoMessageResult
}

func (m *Model) finalizeRepoActivity(index int, next domain.Repository, complete func(activity domain.Activity) ActivityFinalizeResult) {
	m.applyRepoUpdate(index, next, func(repo *domain.Repository, activity domain.Activity) {
		result := complete(activity)
		if !result.Completed {
			return
		}
		if result.Info.OK {
			m.storeRecentActivityInfo(repo.Path, result.Info.Message)
		}
		repo.Activity = &domain.IdleActivity{}
	})
}
