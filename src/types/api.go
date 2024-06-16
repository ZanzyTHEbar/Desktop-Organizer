package types

import (
	"desktop-cleaner/internal"
)

type OnStreamPlanParams struct {
	Msg *internal.StreamMessage
	Err error
}

type OnStreamPlan func(params OnStreamPlanParams)

type ApiClient interface {
	StartTrial() (*internal.StartTrialResponse, *internal.ApiError)
	ConvertTrial(req internal.ConvertTrialRequest) (*internal.SessionResponse, *internal.ApiError)

	CreateEmailVerification(email, customHost, userId string) (*internal.CreateEmailVerificationResponse, *internal.ApiError)

	CreateAccount(req internal.CreateAccountRequest, customHost string) (*internal.SessionResponse, *internal.ApiError)
	SignIn(req internal.SignInRequest, customHost string) (*internal.SessionResponse, *internal.ApiError)
	SignOut() *internal.ApiError

	GetOrgSession() *internal.ApiError
	ListOrgs() ([]*internal.Org, *internal.ApiError)
	CreateOrg(req internal.CreateOrgRequest) (*internal.CreateOrgResponse, *internal.ApiError)

	ListUsers() (*internal.ListUsersResponse, *internal.ApiError)
	DeleteUser(userId string) *internal.ApiError

	ListOrgRoles() ([]*internal.OrgRole, *internal.ApiError)

	InviteUser(req internal.InviteRequest) *internal.ApiError
	ListPendingInvites() ([]*internal.Invite, *internal.ApiError)
	ListAcceptedInvites() ([]*internal.Invite, *internal.ApiError)
	ListAllInvites() ([]*internal.Invite, *internal.ApiError)
	DeleteInvite(inviteId string) *internal.ApiError

	CreateProject(req internal.CreateProjectRequest) (*internal.CreateProjectResponse, *internal.ApiError)
	ListProjects() ([]*internal.Project, *internal.ApiError)
	SetProjectPlan(projectId string, req internal.SetProjectPlanRequest) *internal.ApiError
	RenameProject(projectId string, req internal.RenameProjectRequest) *internal.ApiError

	ListPlans(projectIds []string) ([]*internal.Plan, *internal.ApiError)
	ListArchivedPlans(projectIds []string) ([]*internal.Plan, *internal.ApiError)
	ListPlansRunning(projectIds []string, includeRecent bool) (*internal.ListPlansRunningResponse, *internal.ApiError)

	GetCurrentBranchByPlanId(projectId string, req internal.GetCurrentBranchByPlanIdRequest) (map[string]*internal.Branch, *internal.ApiError)

	GetPlan(planId string) (*internal.Plan, *internal.ApiError)
	CreatePlan(projectId string, req internal.CreatePlanRequest) (*internal.CreatePlanResponse, *internal.ApiError)

	TellPlan(planId, branch string, req internal.TellPlanRequest, onStreamPlan OnStreamPlan) *internal.ApiError
	BuildPlan(planId, branch string, req internal.BuildPlanRequest, onStreamPlan OnStreamPlan) *internal.ApiError
	RespondMissingFile(planId, branch string, req internal.RespondMissingFileRequest) *internal.ApiError

	DeletePlan(planId string) *internal.ApiError
	DeleteAllPlans(projectId string) *internal.ApiError
	ConnectPlan(planId, branch string, onStreamPlan OnStreamPlan) *internal.ApiError
	StopPlan(planId, branch string) *internal.ApiError

	ArchivePlan(planId string) *internal.ApiError
	UnarchivePlan(planId string) *internal.ApiError
	RenamePlan(planId string, name string) *internal.ApiError

	GetCurrentPlanState(planId, branch string) (*internal.CurrentPlanState, *internal.ApiError)
	ApplyPlan(planId, branch string, req internal.ApplyPlanRequest) (string, *internal.ApiError)
	RejectAllChanges(planId, branch string) *internal.ApiError
	RejectFile(planId, branch, filePath string) *internal.ApiError
	RejectFiles(planId, branch string, paths []string) *internal.ApiError
	GetPlanDiffs(planId, branch string) (string, *internal.ApiError)

	LoadContext(planId, branch string, req internal.LoadContextRequest) (*internal.LoadContextResponse, *internal.ApiError)
	UpdateContext(planId, branch string, req internal.UpdateContextRequest) (*internal.UpdateContextResponse, *internal.ApiError)
	DeleteContext(planId, branch string, req internal.DeleteContextRequest) (*internal.DeleteContextResponse, *internal.ApiError)
	ListContext(planId, branch string) ([]*internal.Context, *internal.ApiError)

	ListConvo(planId, branch string) ([]*internal.ConvoMessage, *internal.ApiError)
	GetPlanStatus(planId, branch string) (string, *internal.ApiError)
	ListLogs(planId, branch string) (*internal.LogResponse, *internal.ApiError)
	RewindPlan(planId, branch string, req internal.RewindPlanRequest) (*internal.RewindPlanResponse, *internal.ApiError)

	ListBranches(planId string) ([]*internal.Branch, *internal.ApiError)
	DeleteBranch(planId, branch string) *internal.ApiError
	CreateBranch(planId, branch string, req internal.CreateBranchRequest) *internal.ApiError

	GetSettings(planId, branch string) (*internal.PlanSettings, *internal.ApiError)
	UpdateSettings(planId, branch string, req internal.UpdateSettingsRequest) (*internal.UpdateSettingsResponse, *internal.ApiError)

	GetOrgDefaultSettings() (*internal.PlanSettings, *internal.ApiError)
	UpdateOrgDefaultSettings(req internal.UpdateSettingsRequest) (*internal.UpdateSettingsResponse, *internal.ApiError)

	CreateCustomModel(model *internal.AvailableModel) *internal.ApiError
	ListCustomModels() ([]*internal.AvailableModel, *internal.ApiError)
	DeleteAvailableModel(modelId string) *internal.ApiError

	CreateModelPack(set *internal.ModelPack) *internal.ApiError
	ListModelPacks() ([]*internal.ModelPack, *internal.ApiError)
	DeleteModelPack(setId string) *internal.ApiError
}
