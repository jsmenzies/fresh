package listing

import (
	"fresh/internal/domain"
	"fresh/internal/git"

	tea "charm.land/bubbletea/v2"
)

type streamedWorkState[T any] struct {
	Index    int
	lineChan chan string
	doneChan chan T
}

type pullWorkState = streamedWorkState[pullCompleteMsg]
type pruneWorkState = streamedWorkState[pruneCompleteMsg]

func startStreamedRepoCommand[R any, M any](
	index int,
	repoPath string,
	execute func(lineCallback func(string)) R,
	buildDone func(index int, repo domain.Repository, result R) M,
) streamedWorkState[M] {
	lineChan := make(chan string, 10)
	doneChan := make(chan M, 1)

	go func() {
		result := execute(func(line string) {
			lineChan <- line
		})

		close(lineChan)

		repo := git.BuildRepository(repoPath, cfg)
		doneChan <- buildDone(index, repo, result)
		close(doneChan)
	}()

	return streamedWorkState[M]{
		Index:    index,
		lineChan: lineChan,
		doneChan: doneChan,
	}
}

func listenForStreamedProgress[M any](state streamedWorkState[M], toLineMsg func(index int, line string, next *streamedWorkState[M]) tea.Msg) tea.Cmd {
	return func() tea.Msg {
		select {
		case line, ok := <-state.lineChan:
			if ok {
				return toLineMsg(state.Index, line, &state)
			}
			return <-state.doneChan
		case complete := <-state.doneChan:
			return complete
		}
	}
}
