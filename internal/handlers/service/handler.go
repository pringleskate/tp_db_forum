package serviceHandler

import (
	"github.com/labstack/echo"
	"github.com/pringleskate/tp_db_forum/internal/storages/service"
	"net/http"
)

type Handler interface {
	ServiceClear(c echo.Context) error
	ServiceStatus(c echo.Context) error
}

type handler struct {
	forumStorage service.Service
}

func NewHandler(forumStorage service.Service) *handler {
	return &handler{
		forumStorage: forumStorage,
	}
}

func (h *handler) ServiceClear(c echo.Context) error {
	err := h.forumStorage.Clear()
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *handler) ServiceStatus(c echo.Context) error {
	status, err := h.forumStorage.Status()
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, status)
}