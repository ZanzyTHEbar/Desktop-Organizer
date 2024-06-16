package changes_tui

import (
	"desktop-cleaner/term"
	"fmt"
	"log"

	"desktop-cleaner/internal"
)

type selectionInfo struct {
	currentPath                   string
	currentRes                    *internal.PlanFileResult
	currentReplacements           []*internal.Replacement
	currentRep                    *internal.Replacement
	currentFilesBeforeReplacement *internal.CurrentPlanFiles
}

func (m *changesUIModel) setSelectionInfo() {
	paths := m.currentPlan.PlanResult.SortedPaths
	currentPath := paths[m.selectedFileIndex]

	results := m.currentPlan.PlanResult.FileResultsByPath[currentPath]

	var currentRes *internal.PlanFileResult
	var currentRep *internal.Replacement

	var pathReplacements []*internal.Replacement

	for _, res := range results {
		pathReplacements = append(pathReplacements, res.Replacements...)
	}

	i := 0
outer:
	for _, res := range results {
		if i == 0 && len(res.Replacements) == 0 && res.Content != "" {
			if m.selectedReplacementIndex == 0 {
				currentRes = res
				break outer
			} else {
				i++
				continue
			}
		}

		for _, rep := range res.Replacements {
			if i == m.selectedReplacementIndex {
				currentRes = res
				currentRep = rep
				break outer
			}
			i++
		}
	}

	var currentFilesBeforeReplacement *internal.CurrentPlanFiles
	var err error

	// log.Println("currentRep: ", currentRep)

	if currentRep == nil {
		currentFilesBeforeReplacement = m.currentPlan.CurrentPlanFiles
	} else {
		currentFilesBeforeReplacement, err = m.currentPlan.GetFilesBeforeReplacement(currentRep.Id)
	}

	// log.Println(spew.Sdump(currentFilesBeforeReplacement))

	if err != nil {
		err = fmt.Errorf("error getting current plan state: %v", err)
		log.Println(err)
		term.OutputErrorAndExit(err.Error())
	}

	m.selectionInfo = &selectionInfo{
		currentPath:                   currentPath,
		currentRes:                    currentRes,
		currentReplacements:           pathReplacements,
		currentRep:                    currentRep,
		currentFilesBeforeReplacement: currentFilesBeforeReplacement,
	}
}
