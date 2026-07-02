package service

import "sakeofher/internal/repository"

type adminService struct{ repo *repository.Repositories }

func NewAdminService(repo *repository.Repositories) AdminService { return &adminService{repo: repo} }
