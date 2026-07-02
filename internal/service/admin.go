package service

import "sakeofher/internal/repository"

type AdminService struct{ repo *repository.Repositories }

func NewAdminService(repo *repository.Repositories) *AdminService { return &AdminService{repo: repo} }
