package services

import (
	"ksp/internal/repository"
)

type Service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) *Service{
	return &Service{
		repo : repo,
	}
}

func (s *Service) GetService() {

}