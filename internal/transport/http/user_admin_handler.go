//go:build ignore

// This file is intentionally excluded from normal builds.
//
// The project already has UserHandler, NewUserHandler and all user admin
// handler methods in internal/transport/http/user_handler.go.
//
// A previous patch added this file again and caused duplicate declaration errors.
// Keep this file ignored, or delete it completely from the project.

package httptransport
