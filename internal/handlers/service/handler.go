package serviceHandler

import (
	"github.com/labstack/echo"
	"net/http"
)

type Handler interface {
	ServiceClear(c echo.Context) error
	ServiceStatus(c echo.Context) error
}

type handler struct {
	forumStorage storages.ForumStorage
}

func NewHandler(forumStorage storages.ForumStorage) *handler {
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