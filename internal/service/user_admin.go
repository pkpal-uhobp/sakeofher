//go:build ignore

// This file is intentionally excluded from normal builds.
//
// The project already has userService.List, userService.GetByID,
// userService.Update, userService.Block, userService.Unblock and
// userService.MarkDeleted in internal/service/user.go.
//
// A previous patch added this file again and caused duplicate method errors.
// Keep this file ignored, or delete it completely from the project.

package service
