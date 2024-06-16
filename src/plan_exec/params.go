package plan_exec

import "desktop-cleaner/internal"

type ExecParams struct {
	CurrentPlanId        string
	CurrentBranch        string
	ApiKeys              map[string]string
	CheckOutdatedContext func(maybeContexts []*internal.Context) (bool, bool)
}
