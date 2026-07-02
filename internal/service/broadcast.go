package service

import "sakeofher/internal/repository"

type broadcastService struct {
	repo          *repository.Repositories
	notifications NotificationService
}

func NewBroadcastService(repo *repository.Repositories, notifications NotificationService) BroadcastService {
	return &broadcastService{repo: repo, notifications: notifications}
}
