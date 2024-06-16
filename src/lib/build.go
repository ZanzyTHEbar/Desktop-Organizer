package lib

import "desktop-cleaner/internal"

var buildPlanInlineFn func(maybeContexts []*internal.Context) (bool, error)

func SetBuildPlanInlineFn(fn func(maybeContexts []*internal.Context) (bool, error)) {
	buildPlanInlineFn = fn
}
