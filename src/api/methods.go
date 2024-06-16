package api

import (
	"bytes"
	"desktop-cleaner/internal"
	"desktop-cleaner/types"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func (a *Api) StartTrial() (*internal.StartTrialResponse, *internal.ApiError) {
	serverUrl := cloudApiHost + "/accounts/start_trial"

	resp, err := unauthenticatedClient.Post(serverUrl, "application/json", nil)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		return nil, apiErr
	}

	var startTrialResponse internal.StartTrialResponse
	err = json.NewDecoder(resp.Body).Decode(&startTrialResponse)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %v", err)}
	}

	return &startTrialResponse, nil
}

func (a *Api) CreateProject(req internal.CreateProjectRequest) (*internal.CreateProjectResponse, *internal.ApiError) {
	serverUrl := getApiHost() + "/projects"

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error marshalling request: %v", err)}
	}

	resp, err := authenticatedFastClient.Post(serverUrl, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.CreateProject(req)
		}
		return nil, apiErr
	}

	var respBody internal.CreateProjectResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %v", err)}
	}

	return &respBody, nil
}

func (a *Api) ListProjects() ([]*internal.Project, *internal.ApiError) {
	serverUrl := getApiHost() + "/projects"
	resp, err := authenticatedFastClient.Get(serverUrl)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.ListProjects()
		}
		return nil, apiErr
	}

	var projects []*internal.Project
	err = json.NewDecoder(resp.Body).Decode(&projects)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %v", err)}
	}

	return projects, nil
}

func (a *Api) SetProjectPlan(projectId string, req internal.SetProjectPlanRequest) *internal.ApiError {
	serverUrl := fmt.Sprintf("%s/projects/%s/set_plan", getApiHost(), projectId)
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error marshalling request: %v", err)}
	}

	request, err := http.NewRequest(http.MethodPut, serverUrl, bytes.NewBuffer(reqBytes))
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error creating request: %v", err)}
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := authenticatedFastClient.Do(request)
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		didRefresh, apiErr := refreshTokenIfNeeded(apiErr)
		if didRefresh {
			return a.SetProjectPlan(projectId, req)
		}
		return apiErr
	}

	return nil
}

func (a *Api) RenameProject(projectId string, req internal.RenameProjectRequest) *internal.ApiError {
	serverUrl := fmt.Sprintf("%s/projects/%s/rename", getApiHost(), projectId)
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error marshalling request: %v", err)}
	}

	request, err := http.NewRequest(http.MethodPut, serverUrl, bytes.NewBuffer(reqBytes))
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error creating request: %v", err)}
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := authenticatedFastClient.Do(request)
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)

		didRefresh, apiErr := refreshTokenIfNeeded(apiErr)
		if didRefresh {
			return a.RenameProject(projectId, req)
		}
		return apiErr
	}

	return nil
}
func (a *Api) ListPlans(projectIds []string) ([]*internal.Plan, *internal.ApiError) {
	serverUrl := fmt.Sprintf("%s/plans?", getApiHost())
	parts := []string{}
	for _, projectId := range projectIds {
		parts = append(parts, fmt.Sprintf("projectId=%s", projectId))
	}
	serverUrl += strings.Join(parts, "&")

	resp, err := authenticatedFastClient.Get(serverUrl)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)

		apiErr := handleApiError(resp, errorBody)

		didRefresh, apiErr := refreshTokenIfNeeded(apiErr)
		if didRefresh {
			return a.ListPlans(projectIds)
		}
		return nil, apiErr
	}

	var plans []*internal.Plan
	err = json.NewDecoder(resp.Body).Decode(&plans)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %v", err)}
	}

	return plans, nil
}

func (a *Api) ListArchivedPlans(projectIds []string) ([]*internal.Plan, *internal.ApiError) {
	serverUrl := fmt.Sprintf("%s/plans/archive?", getApiHost())
	parts := []string{}
	for _, projectId := range projectIds {
		parts = append(parts, fmt.Sprintf("projectId=%s", projectId))
	}
	serverUrl += strings.Join(parts, "&")

	resp, err := authenticatedFastClient.Get(serverUrl)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.ListArchivedPlans(projectIds)
		}
		return nil, apiErr
	}

	var plans []*internal.Plan
	err = json.NewDecoder(resp.Body).Decode(&plans)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %v", err)}
	}

	return plans, nil
}

func (a *Api) ListPlansRunning(projectIds []string, includeRecent bool) (*internal.ListPlansRunningResponse, *internal.ApiError) {
	serverUrl := fmt.Sprintf("%s/plans/ps?", getApiHost())
	parts := []string{}
	for _, projectId := range projectIds {
		parts = append(parts, fmt.Sprintf("projectId=%s", projectId))
	}
	serverUrl += strings.Join(parts, "&")
	if includeRecent {
		serverUrl += "&recent=true"
	}

	resp, err := authenticatedFastClient.Get(serverUrl)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.ListPlansRunning(projectIds, includeRecent)
		}
		return nil, apiErr
	}

	var respBody *internal.ListPlansRunningResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %v", err)}
	}

	return respBody, nil
}

func (a *Api) GetCurrentBranchByPlanId(projectId string, req internal.GetCurrentBranchByPlanIdRequest) (map[string]*internal.Branch, *internal.ApiError) {
	serverUrl := fmt.Sprintf("%s/projects/%s/plans/current_branches", getApiHost(), projectId)

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, &internal.ApiError{Msg: fmt.Sprintf("error marshalling request: %v", err)}
	}

	request, err := http.NewRequest(http.MethodPost, serverUrl, bytes.NewBuffer(reqBytes))

	if err != nil {
		return nil, &internal.ApiError{Msg: fmt.Sprintf("error creating request: %v", err)}
	}

	request.Header.Set("Content-Type", "application/json")

	resp, err := authenticatedFastClient.Do(request)

	if err != nil {
		return nil, &internal.ApiError{Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)

		apiErr := handleApiError(resp, errorBody)

		didRefresh, apiErr := refreshTokenIfNeeded(apiErr)
		if didRefresh {
			return a.GetCurrentBranchByPlanId(projectId, req)
		}
		return nil, apiErr
	}

	var respBody map[string]*internal.Branch
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return nil, &internal.ApiError{Msg: fmt.Sprintf("error decoding response: %v", err)}
	}

	return respBody, nil
}

func (a *Api) CreatePlan(projectId string, req internal.CreatePlanRequest) (*internal.CreatePlanResponse, *internal.ApiError) {
	serverUrl := fmt.Sprintf("%s/projects/%s/plans", getApiHost(), projectId)
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error marshalling request: %v", err)}
	}

	resp, err := authenticatedFastClient.Post(serverUrl, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.CreatePlan(projectId, req)
		}
		return nil, apiErr
	}

	var respBody internal.CreatePlanResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %v", err)}
	}

	return &respBody, nil
}

func (a *Api) GetPlan(planId string) (*internal.Plan, *internal.ApiError) {
	serverUrl := fmt.Sprintf("%s/plans/%s", getApiHost(), planId)

	resp, err := authenticatedFastClient.Get(serverUrl)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.GetPlan(planId)
		}
		return nil, apiErr
	}

	var plan internal.Plan
	err = json.NewDecoder(resp.Body).Decode(&plan)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %v", err)}
	}

	return &plan, nil
}

func (a *Api) DeletePlan(planId string) *internal.ApiError {
	serverUrl := fmt.Sprintf("%s/plans/%s", getApiHost(), planId)

	req, err := http.NewRequest(http.MethodDelete, serverUrl, nil)
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error creating request: %v", err)}
	}

	resp, err := authenticatedFastClient.Do(req)
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)

		didRefresh, apiErr := refreshTokenIfNeeded(apiErr)
		if didRefresh {
			return a.DeletePlan(planId)
		}
		return apiErr
	}

	return nil
}

func (a *Api) DeleteAllPlans(projectId string) *internal.ApiError {
	serverUrl := fmt.Sprintf("%s/projects/%s/plans", getApiHost(), projectId)

	req, err := http.NewRequest(http.MethodDelete, serverUrl, nil)
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error creating request: %v", err)}
	}

	resp, err := authenticatedFastClient.Do(req)
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)

		didRefresh, apiErr := refreshTokenIfNeeded(apiErr)

		if didRefresh {
			return a.DeleteAllPlans(projectId)
		}
		return apiErr
	}

	return nil
}

func (a *Api) TellPlan(planId, branch string, req internal.TellPlanRequest, onStream types.OnStreamPlan) *internal.ApiError {

	serverUrl := fmt.Sprintf("%s/plans/%s/%s/tell", getApiHost(), planId, branch)
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error marshalling request: %v", err)}
	}

	request, err := http.NewRequest(http.MethodPost, serverUrl, bytes.NewBuffer(reqBytes))
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error creating request: %v", err)}
	}
	request.Header.Set("Content-Type", "application/json")

	var client *http.Client
	if req.ConnectStream {
		client = authenticatedStreamingClient
	} else {
		client = authenticatedFastClient
	}

	resp, err := client.Do(request)
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error sending request: %v", err)}
	}

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)

		didRefresh, apiErr := refreshTokenIfNeeded(apiErr)

		if didRefresh {
			return a.TellPlan(planId, branch, req, onStream)
		}
		return apiErr
	}

	if req.ConnectStream {
		log.Println("Connecting stream")
		connectPlanRespStream(resp.Body, onStream)
	} else {
		// log.Println("Background exec - not connecting stream")
		resp.Body.Close()
	}

	return nil
}

func (a *Api) BuildPlan(planId, branch string, req internal.BuildPlanRequest, onStream types.OnStreamPlan) *internal.ApiError {

	log.Println("Calling BuildPlan")

	serverUrl := fmt.Sprintf("%s/plans/%s/%s/build", getApiHost(), planId, branch)
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error marshalling request: %v", err)}
	}

	request, err := http.NewRequest(http.MethodPatch, serverUrl, bytes.NewBuffer(reqBytes))
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error creating request: %v", err)}
	}
	request.Header.Set("Content-Type", "application/json")

	var client *http.Client
	if req.ConnectStream {
		client = authenticatedStreamingClient
	} else {
		client = authenticatedFastClient
	}

	resp, err := client.Do(request)
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error sending request: %v", err)}
	}

	if resp.StatusCode >= 400 {
		log.Println("Error response from build plan", resp.StatusCode)

		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)

		didRefresh, apiErr := refreshTokenIfNeeded(apiErr)

		if didRefresh {
			return a.BuildPlan(planId, branch, req, onStream)
		}
		return apiErr
	}

	if req.ConnectStream {
		log.Println("Connecting stream")
		connectPlanRespStream(resp.Body, onStream)
	} else {
		// log.Println("Background exec - not connecting stream")
		resp.Body.Close()
	}

	return nil
}

func (a *Api) RespondMissingFile(planId, branch string, req internal.RespondMissingFileRequest) *internal.ApiError {
	serverUrl := fmt.Sprintf("%s/plans/%s/%s/respond_missing_file", getApiHost(), planId, branch)

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error marshalling request: %v", err)}
	}

	request, err := http.NewRequest(http.MethodPost, serverUrl, bytes.NewBuffer(reqBytes))
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error creating request: %v", err)}
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := authenticatedFastClient.Do(request)
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error sending request: %v", err)}
	}

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)

		didRefresh, apiErr := refreshTokenIfNeeded(apiErr)

		if didRefresh {
			return a.RespondMissingFile(planId, branch, req)
		}
		return apiErr
	}

	return nil

}

func (a *Api) ConnectPlan(planId, branch string, onStream types.OnStreamPlan) *internal.ApiError {
	serverUrl := fmt.Sprintf("%s/plans/%s/%s/connect", getApiHost(), planId, branch)

	req, err := http.NewRequest(http.MethodPatch, serverUrl, nil)
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error creating request: %v", err)}
	}

	resp, err := authenticatedStreamingClient.Do(req)
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error sending request: %v", err)}
	}

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)

		didRefresh, apiErr := refreshTokenIfNeeded(apiErr)

		if didRefresh {
			return a.ConnectPlan(planId, branch, onStream)
		}

		return apiErr
	}

	connectPlanRespStream(resp.Body, onStream)

	return nil
}

func (a *Api) StopPlan(planId, branch string) *internal.ApiError {
	serverUrl := fmt.Sprintf("%s/plans/%s/%s/stop", getApiHost(), planId, branch)

	req, err := http.NewRequest(http.MethodDelete, serverUrl, nil)
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error creating request: %v", err)}
	}

	resp, err := authenticatedFastClient.Do(req)
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		didRefresh, apiErr := refreshTokenIfNeeded(apiErr)
		if didRefresh {
			return a.StopPlan(planId, branch)
		}
		return apiErr
	}

	return nil
}

func (a *Api) GetCurrentPlanState(planId, branch string) (*internal.CurrentPlanState, *internal.ApiError) {
	serverUrl := fmt.Sprintf("%s/plans/%s/%s/current_plan", getApiHost(), planId, branch)

	resp, err := authenticatedFastClient.Get(serverUrl)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.GetCurrentPlanState(planId, branch)
		}
		return nil, apiErr
	}

	var state internal.CurrentPlanState
	err = json.NewDecoder(resp.Body).Decode(&state)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %v", err)}
	}

	return &state, nil
}

func (a *Api) ApplyPlan(planId, branch string, req internal.ApplyPlanRequest) (string, *internal.ApiError) {
	serverUrl := fmt.Sprintf("%s/plans/%s/%s/apply", getApiHost(), planId, branch)

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return "", &internal.ApiError{Msg: fmt.Sprintf("error marshalling request: %v", err)}
	}

	request, err := http.NewRequest(http.MethodPatch, serverUrl, bytes.NewBuffer(reqBytes))
	if err != nil {
		return "", &internal.ApiError{Msg: fmt.Sprintf("error creating request: %v", err)}
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := authenticatedFastClient.Do(request)
	if err != nil {
		return "", &internal.ApiError{Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)

		didRefresh, apiErr := refreshTokenIfNeeded(apiErr)
		if didRefresh {
			return a.ApplyPlan(planId, branch, req)
		}
		return "", apiErr
	}

	// Reading the body on success
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &internal.ApiError{Msg: fmt.Sprintf("error reading response body: %v", err)}
	}

	return string(responseData), nil
}

func (a *Api) ArchivePlan(planId string) *internal.ApiError {
	serverUrl := fmt.Sprintf("%s/plans/%s/archive", getApiHost(), planId)

	req, err := http.NewRequest(http.MethodPatch, serverUrl, nil)
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error creating request: %v", err)}
	}

	resp, err := authenticatedFastClient.Do(req)
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)

		didRefresh, apiErr := refreshTokenIfNeeded(apiErr)
		if didRefresh {
			return a.ArchivePlan(planId)
		}
		return apiErr
	}

	return nil
}

func (a *Api) UnarchivePlan(planId string) *internal.ApiError {
	serverUrl := fmt.Sprintf("%s/plans/%s/unarchive", getApiHost(), planId)

	req, err := http.NewRequest(http.MethodPatch, serverUrl, nil)
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error creating request: %v", err)}
	}

	resp, err := authenticatedFastClient.Do(req)
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)

		didRefresh, apiErr := refreshTokenIfNeeded(apiErr)
		if didRefresh {
			return a.ArchivePlan(planId)
		}
		return apiErr
	}

	return nil
}

func (a *Api) RenamePlan(planId string, name string) *internal.ApiError {
	serverUrl := fmt.Sprintf("%s/plans/%s/rename", getApiHost(), planId)

	reqBytes, err := json.Marshal(internal.RenamePlanRequest{Name: name})
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error marshalling request: %v", err)}
	}

	request, err := http.NewRequest(http.MethodPatch, serverUrl, bytes.NewBuffer(reqBytes))

	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error creating request: %v", err)}
	}

	request.Header.Set("Content-Type", "application/json")

	resp, err := authenticatedFastClient.Do(request)

	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error sending request: %v", err)}
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)

		apiErr := handleApiError(resp, errorBody)

		didRefresh, apiErr := refreshTokenIfNeeded(apiErr)
		if didRefresh {
			return a.RenamePlan(planId, name)
		}
		return apiErr
	}

	return nil
}

func (a *Api) RejectAllChanges(planId, branch string) *internal.ApiError {
	serverUrl := fmt.Sprintf("%s/plans/%s/%s/reject_all", getApiHost(), planId, branch)

	req, err := http.NewRequest(http.MethodPatch, serverUrl, nil)
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error creating request: %v", err)}
	}

	resp, err := authenticatedFastClient.Do(req)
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)

		didRefresh, apiErr := refreshTokenIfNeeded(apiErr)
		if didRefresh {
			return a.RejectAllChanges(planId, branch)
		}
		return apiErr
	}

	return nil
}

func (a *Api) RejectFile(planId, branch, filePath string) *internal.ApiError {
	serverUrl := fmt.Sprintf("%s/plans/%s/%s/reject_file", getApiHost(), planId, branch)

	reqBytes, err := json.Marshal(internal.RejectFileRequest{FilePath: filePath})

	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error marshalling request: %v", err)}
	}

	req, err := http.NewRequest(http.MethodPatch, serverUrl, bytes.NewBuffer(reqBytes))
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error creating request: %v", err)}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := authenticatedFastClient.Do(req)
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		didRefresh, apiErr := refreshTokenIfNeeded(apiErr)
		if didRefresh {
			a.RejectFile(planId, branch, filePath)
		}
		return apiErr
	}

	return nil
}

func (a *Api) RejectFiles(planId, branch string, paths []string) *internal.ApiError {
	serverUrl := fmt.Sprintf("%s/plans/%s/%s/reject_files", getApiHost(), planId, branch)

	reqBytes, err := json.Marshal(internal.RejectFilesRequest{Paths: paths})

	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error marshalling request: %v", err)}
	}

	req, err := http.NewRequest(http.MethodPatch, serverUrl, bytes.NewBuffer(reqBytes))
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error creating request: %v", err)}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := authenticatedFastClient.Do(req)
	if err != nil {
		return &internal.ApiError{Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		didRefresh, apiErr := refreshTokenIfNeeded(apiErr)
		if didRefresh {
			a.RejectFiles(planId, branch, paths)
		}
		return apiErr
	}

	return nil
}

func (a *Api) LoadContext(planId, branch string, req internal.LoadContextRequest) (*internal.LoadContextResponse, *internal.ApiError) {
	serverUrl := fmt.Sprintf("%s/plans/%s/%s/context", getApiHost(), planId, branch)
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error marshalling request: %v", err)}
	}

	// use the slow client since we may be uploading relatively large files
	resp, err := authenticatedSlowClient.Post(serverUrl, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.LoadContext(planId, branch, req)
		}
		return nil, apiErr
	}

	var loadContextResponse internal.LoadContextResponse
	err = json.NewDecoder(resp.Body).Decode(&loadContextResponse)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %v", err)}
	}

	return &loadContextResponse, nil
}

func (a *Api) UpdateContext(planId, branch string, req internal.UpdateContextRequest) (*internal.UpdateContextResponse, *internal.ApiError) {
	serverUrl := fmt.Sprintf("%s/plans/%s/%s/context", getApiHost(), planId, branch)

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error marshalling request: %v", err)}
	}

	request, err := http.NewRequest(http.MethodPut, serverUrl, bytes.NewBuffer(reqBytes))

	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error creating request: %v", err)}
	}

	request.Header.Set("Content-Type", "application/json")

	// use the slow client since we may be uploading relatively large files
	resp, err := authenticatedSlowClient.Do(request)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.UpdateContext(planId, branch, req)
		}
		return nil, apiErr
	}

	var updateContextResponse internal.UpdateContextResponse
	err = json.NewDecoder(resp.Body).Decode(&updateContextResponse)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %v", err)}
	}

	return &updateContextResponse, nil
}

func (a *Api) DeleteContext(planId, branch string, req internal.DeleteContextRequest) (*internal.DeleteContextResponse, *internal.ApiError) {
	serverUrl := fmt.Sprintf("%s/plans/%s/%s/context", getApiHost(), planId, branch)
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error marshalling request: %v", err)}
	}

	request, err := http.NewRequest(http.MethodDelete, serverUrl, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error creating request: %v", err)}
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := authenticatedFastClient.Do(request)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.DeleteContext(planId, branch, req)
		}
		return nil, apiErr
	}

	var deleteContextResponse internal.DeleteContextResponse
	err = json.NewDecoder(resp.Body).Decode(&deleteContextResponse)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %v", err)}
	}

	return &deleteContextResponse, nil
}

func (a *Api) ListContext(planId, branch string) ([]*internal.Context, *internal.ApiError) {
	serverUrl := fmt.Sprintf("%s/plans/%s/%s/context", getApiHost(), planId, branch)

	resp, err := authenticatedFastClient.Get(serverUrl)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.ListContext(planId, branch)
		}
		return nil, apiErr
	}

	var contexts []*internal.Context
	err = json.NewDecoder(resp.Body).Decode(&contexts)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %v", err)}
	}

	return contexts, nil
}

func (a *Api) ListConvo(planId, branch string) ([]*internal.ConvoMessage, *internal.ApiError) {
	serverUrl := fmt.Sprintf("%s/plans/%s/%s/convo", getApiHost(), planId, branch)

	resp, err := authenticatedFastClient.Get(serverUrl)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.ListConvo(planId, branch)
		}
		return nil, apiErr
	}

	var convos []*internal.ConvoMessage
	err = json.NewDecoder(resp.Body).Decode(&convos)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %v", err)}
	}

	return convos, nil
}

func (a *Api) GetPlanStatus(planId, branch string) (string, *internal.ApiError) {
	serverUrl := fmt.Sprintf("%s/plans/%s/%s/status", getApiHost(), planId, branch)

	resp, err := authenticatedFastClient.Get(serverUrl)
	if err != nil {
		return "", &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.GetPlanStatus(planId, branch)
		}
		return "", apiErr
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error reading response body: %v", err)}
	}

	return string(body), nil
}

func (a *Api) GetPlanDiffs(planId, branch string) (string, *internal.ApiError) {
	serverUrl := fmt.Sprintf("%s/plans/%s/%s/diffs", getApiHost(), planId, branch)

	resp, err := authenticatedFastClient.Get(serverUrl)
	if err != nil {
		return "", &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.GetPlanDiffs(planId, branch)
		}
		return "", apiErr
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error reading response body: %v", err)}
	}

	return string(body), nil
}

func (a *Api) ListLogs(planId, branch string) (*internal.LogResponse, *internal.ApiError) {
	serverUrl := fmt.Sprintf("%s/plans/%s/%s/logs", getApiHost(), planId, branch)

	resp, err := authenticatedFastClient.Get(serverUrl)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.ListLogs(planId, branch)
		}
		return nil, apiErr
	}

	var logs internal.LogResponse
	err = json.NewDecoder(resp.Body).Decode(&logs)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %v", err)}
	}

	return &logs, nil
}

func (a *Api) RewindPlan(planId, branch string, req internal.RewindPlanRequest) (*internal.RewindPlanResponse, *internal.ApiError) {
	serverUrl := fmt.Sprintf("%s/plans/%s/%s/rewind", getApiHost(), planId, branch)
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error marshalling request: %v", err)}
	}

	request, err := http.NewRequest(http.MethodPatch, serverUrl, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error creating request: %v", err)}
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := authenticatedFastClient.Do(request)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.RewindPlan(planId, branch, req)
		}
		return nil, apiErr
	}

	var rewindPlanResponse internal.RewindPlanResponse
	err = json.NewDecoder(resp.Body).Decode(&rewindPlanResponse)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %v", err)}
	}

	return &rewindPlanResponse, nil
}

func (a *Api) SignIn(req internal.SignInRequest, customHost string) (*internal.SessionResponse, *internal.ApiError) {
	host := customHost
	if host == "" {
		host = cloudApiHost
	}
	serverUrl := host + "/accounts/sign_in"
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error marshalling request: %v", err)}
	}

	resp, err := unauthenticatedClient.Post(serverUrl, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		return nil, apiErr
	}

	var sessionResponse internal.SessionResponse
	err = json.NewDecoder(resp.Body).Decode(&sessionResponse)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %v", err)}
	}

	return &sessionResponse, nil
}

func (a *Api) CreateAccount(req internal.CreateAccountRequest, customHost string) (*internal.SessionResponse, *internal.ApiError) {
	host := customHost
	if host == "" {
		host = cloudApiHost
	}
	serverUrl := host + "/accounts"
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error marshalling request: %v", err)}
	}

	resp, err := unauthenticatedClient.Post(serverUrl, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		return nil, apiErr
	}

	var sessionResponse internal.SessionResponse
	err = json.NewDecoder(resp.Body).Decode(&sessionResponse)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %v", err)}
	}

	return &sessionResponse, nil
}

func (a *Api) ConvertTrial(req internal.ConvertTrialRequest) (*internal.SessionResponse, *internal.ApiError) {
	serverUrl := getApiHost() + "/accounts/convert_trial"
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error marshalling request: %v", err)}
	}

	resp, err := authenticatedFastClient.Post(serverUrl, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		return nil, apiErr
	}

	var sessionResponse internal.SessionResponse
	err = json.NewDecoder(resp.Body).Decode(&sessionResponse)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %v", err)}
	}

	return &sessionResponse, nil
}

func (a *Api) CreateOrg(req internal.CreateOrgRequest) (*internal.CreateOrgResponse, *internal.ApiError) {
	serverUrl := getApiHost() + "/orgs"
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error marshalling request: %v", err)}
	}

	resp, err := authenticatedFastClient.Post(serverUrl, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.CreateOrg(req)
		}
		return nil, apiErr
	}

	var createOrgResponse internal.CreateOrgResponse
	err = json.NewDecoder(resp.Body).Decode(&createOrgResponse)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %v", err)}
	}

	return &createOrgResponse, nil
}

func (a *Api) GetOrgSession() *internal.ApiError {
	serverUrl := getApiHost() + "/orgs/session"
	resp, err := authenticatedFastClient.Get(serverUrl)
	if err != nil {
		return &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		return apiErr
	}

	return nil
}

func (a *Api) ListOrgs() ([]*internal.Org, *internal.ApiError) {
	serverUrl := getApiHost() + "/orgs"
	resp, err := authenticatedFastClient.Get(serverUrl)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.ListOrgs()
		}
		return nil, apiErr
	}

	var orgs []*internal.Org
	err = json.NewDecoder(resp.Body).Decode(&orgs)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %v", err)}
	}

	return orgs, nil
}

func (a *Api) DeleteUser(userId string) *internal.ApiError {
	serverUrl := fmt.Sprintf("%s/orgs/users/%s", getApiHost(), userId)
	req, err := http.NewRequest(http.MethodDelete, serverUrl, nil)
	if err != nil {
		return &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error creating request: %v", err)}
	}

	resp, err := authenticatedFastClient.Do(req)
	if err != nil {
		return &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.DeleteUser(userId)
		}
		return apiErr
	}

	return nil
}

func (a *Api) ListOrgRoles() ([]*internal.OrgRole, *internal.ApiError) {
	serverUrl := getApiHost() + "/orgs/roles"
	resp, err := authenticatedFastClient.Get(serverUrl)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %s", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)

		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.ListOrgRoles()
		}
		return nil, apiErr
	}

	var roles []*internal.OrgRole
	err = json.NewDecoder(resp.Body).Decode(&roles)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %s", err)}
	}

	return roles, nil
}

func (a *Api) InviteUser(req internal.InviteRequest) *internal.ApiError {
	serverUrl := getApiHost() + "/invites"
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error marshalling request: %v", err)}
	}

	resp, err := authenticatedFastClient.Post(serverUrl, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)

		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.InviteUser(req)
		}
		return apiErr
	}

	return nil
}

func (a *Api) ListPendingInvites() ([]*internal.Invite, *internal.ApiError) {
	serverUrl := getApiHost() + "/invites/pending"
	resp, err := authenticatedFastClient.Get(serverUrl)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.ListPendingInvites()
		}
		return nil, apiErr
	}

	var invites []*internal.Invite
	err = json.NewDecoder(resp.Body).Decode(&invites)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %v", err)}
	}

	return invites, nil
}

func (a *Api) ListAcceptedInvites() ([]*internal.Invite, *internal.ApiError) {
	serverUrl := getApiHost() + "/invites/accepted"
	resp, err := authenticatedFastClient.Get(serverUrl)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.ListAcceptedInvites()
		}
		return nil, apiErr
	}

	var invites []*internal.Invite
	err = json.NewDecoder(resp.Body).Decode(&invites)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %v", err)}
	}

	return invites, nil
}

func (a *Api) ListAllInvites() ([]*internal.Invite, *internal.ApiError) {
	serverUrl := getApiHost() + "/invites/all"
	resp, err := authenticatedFastClient.Get(serverUrl)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.ListAllInvites()
		}
		return nil, apiErr
	}

	var invites []*internal.Invite
	err = json.NewDecoder(resp.Body).Decode(&invites)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %v", err)}
	}

	return invites, nil
}

func (a *Api) DeleteInvite(inviteId string) *internal.ApiError {
	serverUrl := fmt.Sprintf("%s/invites/%s", getApiHost(), inviteId)
	req, err := http.NewRequest(http.MethodDelete, serverUrl, nil)
	if err != nil {
		return &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error creating request: %v", err)}
	}

	resp, err := authenticatedFastClient.Do(req)
	if err != nil {
		return &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)

		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.DeleteInvite(inviteId)
		}
		return apiErr
	}

	return nil
}

func (a *Api) CreateEmailVerification(email, customHost, userId string) (*internal.CreateEmailVerificationResponse, *internal.ApiError) {
	host := customHost
	if host == "" {
		host = cloudApiHost
	}
	serverUrl := host + "/accounts/email_verifications"
	req := internal.CreateEmailVerificationRequest{Email: email, UserId: userId}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error marshalling request: %v", err)}
	}

	resp, err := unauthenticatedClient.Post(serverUrl, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		return nil, handleApiError(resp, errorBody)
	}

	var verificationResponse internal.CreateEmailVerificationResponse
	err = json.NewDecoder(resp.Body).Decode(&verificationResponse)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %v", err)}
	}

	return &verificationResponse, nil
}

func (a *Api) SignOut() *internal.ApiError {
	serverUrl := getApiHost() + "/accounts/sign_out"

	req, err := http.NewRequest(http.MethodPost, serverUrl, nil)
	if err != nil {
		return &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error creating request: %v", err)}
	}

	resp, err := authenticatedFastClient.Do(req)
	if err != nil {
		return &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		return handleApiError(resp, errorBody)
	}

	return nil
}

func (a *Api) ListUsers() (*internal.ListUsersResponse, *internal.ApiError) {
	serverUrl := getApiHost() + "/users"
	resp, err := authenticatedFastClient.Get(serverUrl)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.ListUsers()
		}
		return nil, apiErr
	}

	var r *internal.ListUsersResponse
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %v", err)}
	}

	return r, nil
}

func (a *Api) ListBranches(planId string) ([]*internal.Branch, *internal.ApiError) {
	serverUrl := fmt.Sprintf("%s/plans/%s/branches", getApiHost(), planId)

	resp, err := authenticatedFastClient.Get(serverUrl)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %s", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)

		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.ListBranches(planId)
		}
		return nil, apiErr
	}

	var branches []*internal.Branch
	err = json.NewDecoder(resp.Body).Decode(&branches)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %s", err)}
	}

	return branches, nil
}

func (a *Api) CreateBranch(planId, branch string, req internal.CreateBranchRequest) *internal.ApiError {
	serverUrl := fmt.Sprintf("%s/plans/%s/%s/branches", getApiHost(), planId, branch)

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error marshalling request: %s", err)}
	}

	resp, err := authenticatedFastClient.Post(serverUrl, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %s", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)

		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.CreateBranch(planId, branch, req)
		}
		return apiErr
	}

	return nil
}

func (a *Api) DeleteBranch(planId, branch string) *internal.ApiError {
	serverUrl := fmt.Sprintf("%s/plans/%s/branches/%s", getApiHost(), planId, branch)

	req, err := http.NewRequest(http.MethodDelete, serverUrl, nil)
	if err != nil {
		return &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error creating request: %s", err)}
	}

	resp, err := authenticatedFastClient.Do(req)
	if err != nil {
		return &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %s", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)

		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.DeleteBranch(planId, branch)
		}
		return apiErr
	}

	return nil
}

func (a *Api) GetSettings(planId, branch string) (*internal.PlanSettings, *internal.ApiError) {
	serverUrl := fmt.Sprintf("%s/plans/%s/%s/settings", getApiHost(), planId, branch)

	resp, err := authenticatedFastClient.Get(serverUrl)

	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %s", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.GetSettings(planId, branch)
		}
		return nil, apiErr
	}

	var settings internal.PlanSettings
	err = json.NewDecoder(resp.Body).Decode(&settings)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %s", err)}
	}

	return &settings, nil
}

func (a *Api) UpdateSettings(planId, branch string, req internal.UpdateSettingsRequest) (*internal.UpdateSettingsResponse, *internal.ApiError) {
	serverUrl := fmt.Sprintf("%s/plans/%s/%s/settings", getApiHost(), planId, branch)

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error marshalling request: %s", err)}
	}

	// log.Println("UpdateSettings", string(reqBytes))

	request, err := http.NewRequest(http.MethodPut, serverUrl, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error creating request: %s", err)}
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := authenticatedFastClient.Do(request)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %s", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)

		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.UpdateSettings(planId, branch, req)
		}
		return nil, apiErr
	}

	var updateRes internal.UpdateSettingsResponse
	err = json.NewDecoder(resp.Body).Decode(&updateRes)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %s", err)}
	}

	return &updateRes, nil

}

func (a *Api) GetOrgDefaultSettings() (*internal.PlanSettings, *internal.ApiError) {
	serverUrl := fmt.Sprintf("%s/default_settings", getApiHost())

	resp, err := authenticatedFastClient.Get(serverUrl)

	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %s", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.GetOrgDefaultSettings()
		}
		return nil, apiErr
	}

	var settings internal.PlanSettings
	err = json.NewDecoder(resp.Body).Decode(&settings)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %s", err)}
	}

	return &settings, nil
}

func (a *Api) UpdateOrgDefaultSettings(req internal.UpdateSettingsRequest) (*internal.UpdateSettingsResponse, *internal.ApiError) {
	serverUrl := fmt.Sprintf("%s/default_settings", getApiHost())

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error marshalling request: %s", err)}
	}

	// log.Println("UpdateSettings", string(reqBytes))

	request, err := http.NewRequest(http.MethodPut, serverUrl, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error creating request: %s", err)}
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := authenticatedFastClient.Do(request)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %s", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)

		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.UpdateOrgDefaultSettings(req)
		}
		return nil, apiErr
	}

	var updateRes internal.UpdateSettingsResponse
	err = json.NewDecoder(resp.Body).Decode(&updateRes)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %s", err)}
	}

	return &updateRes, nil

}

func (a *Api) CreateCustomModel(model *internal.AvailableModel) *internal.ApiError {
	serverUrl := fmt.Sprintf("%s/custom_models", getApiHost())
	body, err := json.Marshal(model)
	if err != nil {
		return &internal.ApiError{Msg: "Failed to marshal model"}
	}

	resp, err := authenticatedFastClient.Post(serverUrl, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)

		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.CreateCustomModel(model)
		}
		return apiErr
	}

	return nil
}

func (a *Api) ListCustomModels() ([]*internal.AvailableModel, *internal.ApiError) {
	serverUrl := fmt.Sprintf("%s/custom_models", getApiHost())
	resp, err := authenticatedFastClient.Get(serverUrl)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)

		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.ListCustomModels()
		}
		return nil, apiErr
	}

	var models []*internal.AvailableModel
	err = json.NewDecoder(resp.Body).Decode(&models)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %v", err)}
	}

	return models, nil
}

func (a *Api) DeleteAvailableModel(modelId string) *internal.ApiError {
	serverUrl := fmt.Sprintf("%s/custom_models/%s", getApiHost(), modelId)
	req, err := http.NewRequest(http.MethodDelete, serverUrl, nil)
	if err != nil {
		return &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error creating request: %v", err)}
	}

	resp, err := authenticatedFastClient.Do(req)
	if err != nil {
		return &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)

		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.DeleteAvailableModel(modelId)
		}
		return apiErr
	}

	return nil
}

func (a *Api) CreateModelPack(set *internal.ModelPack) *internal.ApiError {
	serverUrl := fmt.Sprintf("%s/model_sets", getApiHost())
	body, err := json.Marshal(set)
	if err != nil {
		return &internal.ApiError{Msg: "Failed to marshal model pack"}
	}

	resp, err := authenticatedFastClient.Post(serverUrl, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)

		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.CreateModelPack(set)
		}
		return apiErr
	}

	return nil

}

func (a *Api) ListModelPacks() ([]*internal.ModelPack, *internal.ApiError) {
	serverUrl := fmt.Sprintf("%s/model_sets", getApiHost())

	resp, err := authenticatedFastClient.Get(serverUrl)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)

		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.ListModelPacks()
		}
		return nil, apiErr
	}

	var sets []*internal.ModelPack
	err = json.NewDecoder(resp.Body).Decode(&sets)
	if err != nil {
		return nil, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error decoding response: %v", err)}
	}

	return sets, nil

}

func (a *Api) DeleteModelPack(setId string) *internal.ApiError {
	serverUrl := fmt.Sprintf("%s/model_sets/%s", getApiHost(), setId)

	req, err := http.NewRequest(http.MethodDelete, serverUrl, nil)
	if err != nil {
		return &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error creating request: %v", err)}
	}

	resp, err := authenticatedFastClient.Do(req)
	if err != nil {
		return &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: fmt.Sprintf("error sending request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)

		apiErr := handleApiError(resp, errorBody)
		tokenRefreshed, apiErr := refreshTokenIfNeeded(apiErr)
		if tokenRefreshed {
			return a.DeleteModelPack(setId)
		}
		return apiErr
	}

	return nil
}
