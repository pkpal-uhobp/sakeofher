package service

import "sakeofher/internal/repository"

type BroadcastService struct {
	repo          *repository.Repositories
	notifications *NotificationService
}

func NewBroadcastService(repo *repository.Repositories, notifications *NotificationService) *BroadcastService {
	return &BroadcastService{repo: repo, notifications: notifications}
}
