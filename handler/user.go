package handler

import (
	"github.com/form3tech-oss/jwt-go"
	"github.com/gofiber/fiber/v2"
	"ufiber/usecase"
)

type userH struct {
	userI usecase.UserI
}

func NewUserH(userI usecase.UserI) *userH {
	return &userH{userI}
}

func (u *userH) Router(router fiber.Router) {
	router.Post("/register", u.register)
	router.Post("/login", u.login)
	router.Post("/refresh", u.reToken)
	router.Get("/", accessible)
}

func (u *userH) register(c *fiber.Ctx) error {
	return u.userI.Register(c)
}

func (u *userH) login(c *fiber.Ctx) error {
	return u.userI.Login(c)
}

func (u *userH) reToken(c *fiber.Ctx) error {
	return u.userI.ReToken(c)
}

func accessible(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	id := claims["id"].(string)
	return c.SendString("Welcome " + id)
}
