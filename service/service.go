package service

import (
	m "locations/model"
)

type Service struct {
}

type LocationsService interface {
	Health(*m.ServiceCommand) (*m.ServiceCommand, error)
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Health(sc *m.ServiceCommand) (*m.ServiceCommand, error) {
	sc.Completed = true

	return sc, nil
}
