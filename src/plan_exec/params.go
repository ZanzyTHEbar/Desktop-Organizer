package plan_exec

import "desktop-cleaner/shared"

type ExecParams struct {
	CurrentPlanId        string
	CurrentBranch        string
	ApiKeys              map[string]string
	CheckOutdatedContext func(maybeContexts []*shared.Context) (bool, bool)
}
