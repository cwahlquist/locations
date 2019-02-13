package handler

import (
	"context"
	"log"

	pb "locations/api/go"
	m "locations/model"
	s "locations/service"
)

type Handler struct {
	service s.Service
}

func NewHandler(service *s.Service) *Handler {
	return &Handler{
		service: *service,
	}
}

func (h *Handler) Health(ctx context.Context, req *pb.ServiceCommand) (*pb.ServiceCommand, error) {
	log.Println("Health: ", *req)

	model := protoToModel(req)

	todo, err := h.service.Health(model)
	if err != nil {
		return nil, err
	}

	return modelToProto(todo), nil
}

// Private conversion methods
func protoToModel(pb *pb.ServiceCommand) *m.ServiceCommand {
	return &m.ServiceCommand{
		Id:        pb.Id,
		Name:      pb.Name,
		Completed: pb.Completed,
	}
}

func modelToProto(model *m.ServiceCommand) *pb.ServiceCommand {
	return &pb.ServiceCommand{
		Id:        model.Id,
		Name:      model.Name,
		Completed: model.Completed,
	}
}
