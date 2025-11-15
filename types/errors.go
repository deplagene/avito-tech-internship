package types

import "errors"

var (
	ErrNotFound      = errors.New("resource not found")
	ErrAlreadyExists = errors.New("resource already exists")
	ErrPRMerged      = errors.New("pr is already merged")
	ErrNotAssigned   = errors.New("reviewer is not assigned to this pr")
	ErrNoCandidate   = errors.New("no active replacement candidate in team")
)
