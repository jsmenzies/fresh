package git

import "strings"

type ghCheckConclusion string

const (
	ghCheckConclusionSuccess        ghCheckConclusion = "SUCCESS"
	ghCheckConclusionSkipped        ghCheckConclusion = "SKIPPED"
	ghCheckConclusionNeutral        ghCheckConclusion = "NEUTRAL"
	ghCheckConclusionFailure        ghCheckConclusion = "FAILURE"
	ghCheckConclusionCancelled      ghCheckConclusion = "CANCELLED"
	ghCheckConclusionTimedOut       ghCheckConclusion = "TIMED_OUT"
	ghCheckConclusionActionRequired ghCheckConclusion = "ACTION_REQUIRED"
	ghCheckConclusionStartupFailure ghCheckConclusion = "STARTUP_FAILURE"
	ghCheckConclusionStale          ghCheckConclusion = "STALE"
)

type ghCheckStatus string

const (
	ghCheckStatusCompleted  ghCheckStatus = "COMPLETED"
	ghCheckStatusInProgress ghCheckStatus = "IN_PROGRESS"
)

type ghCheckState string

const (
	ghCheckStateSuccess  ghCheckState = "SUCCESS"
	ghCheckStateFailure  ghCheckState = "FAILURE"
	ghCheckStateError    ghCheckState = "ERROR"
	ghCheckStatePending  ghCheckState = "PENDING"
	ghCheckStateExpected ghCheckState = "EXPECTED"
)

type checkSummaryClass string

const (
	checkSummaryPassed  checkSummaryClass = "passed"
	checkSummaryRunning checkSummaryClass = "running"
	checkSummaryFailed  checkSummaryClass = "failed"
	checkSummarySkipped checkSummaryClass = "skipped"
	checkSummaryWaiting checkSummaryClass = "waiting"
)

type pullRequestCheckGate int

const (
	pullRequestCheckGateClear pullRequestCheckGate = iota
	pullRequestCheckGatePending
	pullRequestCheckGateFailing
)

func classifyCheckSummary(conclusionRaw, statusRaw, stateRaw string) checkSummaryClass {
	conclusion := normalizeCheckConclusion(conclusionRaw)
	switch conclusion {
	case ghCheckConclusionSuccess:
		return checkSummaryPassed
	case ghCheckConclusionSkipped, ghCheckConclusionNeutral:
		return checkSummarySkipped
	case ghCheckConclusionFailure, ghCheckConclusionCancelled, ghCheckConclusionTimedOut, ghCheckConclusionActionRequired, ghCheckConclusionStartupFailure, ghCheckConclusionStale:
		return checkSummaryFailed
	}

	status := normalizeCheckStatus(statusRaw)
	if status != "" && status != ghCheckStatusCompleted {
		if status == ghCheckStatusInProgress {
			return checkSummaryRunning
		}
		return checkSummaryWaiting
	}

	state := normalizeCheckState(stateRaw)
	switch state {
	case ghCheckStateSuccess:
		return checkSummaryPassed
	case ghCheckStateFailure, ghCheckStateError:
		return checkSummaryFailed
	case ghCheckStatePending, ghCheckStateExpected:
		return checkSummaryWaiting
	}

	if status == ghCheckStatusCompleted {
		return checkSummaryPassed
	}

	return checkSummaryWaiting
}

func classifyCheckGate(conclusionRaw, statusRaw, stateRaw string) pullRequestCheckGate {
	conclusion := normalizeCheckConclusion(conclusionRaw)
	switch conclusion {
	case ghCheckConclusionFailure, ghCheckConclusionCancelled, ghCheckConclusionTimedOut, ghCheckConclusionActionRequired, ghCheckConclusionStartupFailure:
		return pullRequestCheckGateFailing
	case "":
		status := normalizeCheckStatus(statusRaw)
		if status != ghCheckStatusCompleted {
			return pullRequestCheckGatePending
		}
	}

	state := normalizeCheckState(stateRaw)
	switch state {
	case ghCheckStateFailure, ghCheckStateError:
		return pullRequestCheckGateFailing
	case ghCheckStatePending, ghCheckStateExpected:
		return pullRequestCheckGatePending
	default:
		return pullRequestCheckGateClear
	}
}

func classifyRollupGate(stateRaw string) pullRequestCheckGate {
	switch normalizeCheckState(stateRaw) {
	case ghCheckStateFailure, ghCheckStateError:
		return pullRequestCheckGateFailing
	case ghCheckStatePending, ghCheckStateExpected:
		return pullRequestCheckGatePending
	default:
		return pullRequestCheckGateClear
	}
}

func normalizeCheckConclusion(value string) ghCheckConclusion {
	return ghCheckConclusion(strings.ToUpper(strings.TrimSpace(value)))
}

func normalizeCheckStatus(value string) ghCheckStatus {
	return ghCheckStatus(strings.ToUpper(strings.TrimSpace(value)))
}

func normalizeCheckState(value string) ghCheckState {
	return ghCheckState(strings.ToUpper(strings.TrimSpace(value)))
}
