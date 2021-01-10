package profileHandler

import (
	"github.com/jackc/pgx"
	"github.com/labstack/echo"
	"github.com/pringleskate/tp_db_forum/internal/models"
	"github.com/pringleskate/tp_db_forum/internal/storages/profile"
	"log"
	"net/http"
	"strings"
)

type Handler interface {
	ProfileCreate(c echo.Context) error
	ProfileGet(c echo.Context) error
	ProfileUpdate(c echo.Context) error
}

type handler struct {
	userStorage profile.Storage
}

func NewHandler(userStorage profile.Storage) *handler {
	return &handler{
		userStorage: userStorage,
	}
}

func (h *handler) ProfileCreate(c echo.Context) error {
	userInput := new(models.User)
	if err := c.Bind(userInput); err != nil {
		return err
	}

	userInput.Nickname = c.Param("nickname")

	user := make([]models.User, 0)

	usr, err := h.userStorage.GetFullUserByNickname(userInput.Nickname)
	if err == nil {
		user = append(user, usr)
		if strings.ToLower(usr.Email) == strings.ToLower(userInput.Email) {
			return c.JSON(http.StatusConflict, user)
		}
	}

	if err != pgx.ErrNoRows && err != nil {
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}

	usr, err = h.userStorage.GetFullUserByEmail(userInput.Email)
	if err == nil {
		user = append(user, usr)
	}
	if err != pgx.ErrNoRows && err != nil{
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}

	if len(user) != 0 {
		return c.JSON(http.StatusConflict, user)
	}

	err = h.userStorage.InsertUser(*userInput)
	if err != nil {
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}

	user = append(user, *userInput)

	return c.JSON(http.StatusCreated, user[0])
}

func (h *handler) ProfileGet(c echo.Context) error {

	nickname := c.Param("nickname")

	user, err := h.userStorage.GetFullUserByNickname(nickname)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.JSON(models.NotFound, models.Error{Message: ""})
		}
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}
	return c.JSON(http.StatusOK, user)
}

func (h *handler) ProfileUpdate(c echo.Context) error {
	userInput := new(models.User)
	if err := c.Bind(userInput); err != nil {
		return err
	}

	userInput.Nickname = c.Param("nickname")

	user, err := h.userStorage.GetFullUserByNickname(userInput.Nickname)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.JSON(models.NotFound, models.Error{Message: ""})
		}
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}

	_, err = h.userStorage.GetFullUserByEmail(userInput.Email)
	if err == nil {
		return c.JSON(http.StatusConflict, user)
	}
	if err != pgx.ErrNoRows {
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}

	if userInput.Email != "" {
		user.Email = userInput.Email
	}
	if userInput.About != "" {
		user.About = userInput.About
	}
	if userInput.FullName != "" {
		user.FullName = userInput.FullName
	}

	err = h.userStorage.UpdateUser(user)
	if err != nil {
		log.Print(err)
		return c.JSON(models.InternalServerError, models.Error{Message: ""})
	}

	return c.JSON(http.StatusOK, user)
}
