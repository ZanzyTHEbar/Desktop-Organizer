package errsx

import (
	"encoding/json"
	"errors"
	"fmt"
)

type DesktopCleanerErrDetails struct {
	errors ErrorMap
}

type DesktopCleanerErrBuilder struct {
	code    DesktopCleanerErr
	msg     string
	cause   error
	label   string
	details DesktopCleanerErrDetails
}

// NewDesktopCleanerErrDetails is a constructor for DesktopCleanerErrDetails
func NewDesktopCleanerErrDetails(errors ErrorMap) DesktopCleanerErrDetails {
	return DesktopCleanerErrDetails{errors: errors}
}

// NewDesktopCleanerErrBuilder is a constructor for DesktopCleanerErrBuilder
func NewDesktopCleanerErrBuilder() *DesktopCleanerErrBuilder {
	return &DesktopCleanerErrBuilder{}
}

// MarshalJSON implements the json.Marshaler interface.
func (builder *DesktopCleanerErrBuilder) MarshalJSON() ([]byte, error) {
	// use json.Marshal to convert the error message to a JSON byte slice
	byteBuffer, err := json.Marshal(map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	return byteBuffer, nil
}

// Error is a method to return an error, this is an implementation of the error interface.
func (builder *DesktopCleanerErrBuilder) Error() string {

	// validate the error instance, if it is nil, return nil
	if builder.code == 0 || builder.msg == "" {
		builder.code = Internal
		builder.msg = "Internal Server Error"
	}

	// if the cause is nil, set the cause to the error message
	if builder.cause == nil {
		builder.cause = errors.New(builder.msg)
	}

	// if the label is empty, set the label to the String of the error code
	if builder.label == "" {
		builder.label = builder.code.String()
	}

	// convert the builder instance to a formatted error message and return it
	return fmt.Sprintf("code: %d, label: %s, message: %s, cause: %s, details: %v", builder.code, builder.label, builder.msg, builder.cause.Error(), builder.details)
}

// WithCode is a method to set the error code.
func (builder *DesktopCleanerErrBuilder) WithCode(code DesktopCleanerErr) *DesktopCleanerErrBuilder {
	builder.code = code
	return builder
}

// WithMsg is a method to set the error message.
func (builder *DesktopCleanerErrBuilder) WithMsg(msg string) *DesktopCleanerErrBuilder {
	builder.msg = msg
	return builder
}

// WithCause is a method to set the error cause.
func (builder *DesktopCleanerErrBuilder) WithCause(cause error) *DesktopCleanerErrBuilder {
	builder.cause = cause
	return builder
}

// WithDetails is a method to set the error details.
func (builder *DesktopCleanerErrBuilder) WithDetails(details DesktopCleanerErrDetails) *DesktopCleanerErrBuilder {
	builder.details = details
	return builder
}

// ErrDetails is a method to return the error details as a map.
func (err *DesktopCleanerErrDetails) ErrDetails() (ErrorMap, error) {
	if err.errors == nil {
		return nil, errors.New("no error details found")
	}

	return err.errors, nil
}
