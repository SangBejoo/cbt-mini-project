package handler

import (
	base "cbt-test-mini-project/gen/proto"
)

// baseHandler implements the Base service
type baseHandler struct {
	base.UnimplementedBaseServer
}

// NewBaseHandler creates a new baseHandler instance
func NewBaseHandler() *baseHandler {
	return &baseHandler{}
}
