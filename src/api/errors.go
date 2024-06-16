package api

import (
	"desktop-cleaner/auth"
	"desktop-cleaner/internal"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func handleApiError(r *http.Response, errBody []byte) *internal.ApiError {
	// Check if the response is JSON
	if r.Header.Get("Content-Type") != "application/json" {
		return &internal.ApiError{
			Type:   internal.ApiErrorTypeOther,
			Status: r.StatusCode,
			Msg:    strings.TrimSpace(string(errBody)),
		}
	}

	var apiError internal.ApiError
	if err := json.Unmarshal(errBody, &apiError); err != nil {
		log.Printf("Error unmarshalling JSON: %v\n", err)
		return &internal.ApiError{
			Type:   internal.ApiErrorTypeOther,
			Status: r.StatusCode,
			Msg:    strings.TrimSpace(string(errBody)),
		}
	}

	return &apiError
}

func refreshTokenIfNeeded(apiErr *internal.ApiError) (bool, *internal.ApiError) {
	if apiErr.Type == internal.ApiErrorTypeInvalidToken {
		err := auth.RefreshInvalidToken()
		if err != nil {
			return false, &internal.ApiError{Type: internal.ApiErrorTypeOther, Msg: "error refreshing invalid token"}
		}
		return true, nil
	}
	return false, apiErr
}
