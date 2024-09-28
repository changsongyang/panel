package frp

import (
	"fmt"
	"net/http"

	"github.com/TheTNB/panel/internal/panel"
	"github.com/TheTNB/panel/internal/service"
	"github.com/TheTNB/panel/pkg/io"
	"github.com/TheTNB/panel/pkg/systemctl"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GetConfig(w http.ResponseWriter, r *http.Request) {
	req, err := service.Bind[Name](r)
	if err != nil {
		service.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	config, err := io.Read(fmt.Sprintf("%s/server/frp/%s.toml", panel.Root, req.Name))
	if err != nil {
		service.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	service.Success(w, config)
}

func (s *Service) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	req, err := service.Bind[UpdateConfig](r)
	if err != nil {
		service.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	if err = io.Write(fmt.Sprintf("%s/server/frp/%s.toml", panel.Root, req.Name), req.Config, 0644); err != nil {
		service.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err = systemctl.Restart(req.Name); err != nil {
		service.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	service.Success(w, nil)
}