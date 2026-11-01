package core_errors

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrInvalidArgument = errors.New("Invalid argument")
	ErrConflict        = errors.New("Conflict")
	ErrTaskExpired     = errors.New("Task expiresd")
	ErrTooManyRequests = errors.New("Too many rerquests")
	ErrCodeUsed        = errors.New("Code already used")
	ErrCodeExpired     = errors.New("Code expired")
	ErrInvalidCode     = errors.New("Invalid code")
)
