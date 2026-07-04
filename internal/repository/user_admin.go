//go:build ignore

// This file is intentionally excluded from normal builds.
//
// The project already has UserRepository.List, UserRepository.Update and
// UserRepository.SetStatus in internal/repository/user.go.
// A previous patch added this file again and caused duplicate method errors:
//
//   method UserRepository.List already declared
//   method UserRepository.Update already declared
//   method UserRepository.SetStatus already declared
//
// Keep this file ignored, or delete it completely from the project.

package repository
