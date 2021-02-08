package handler

import (
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
	router.Get("/active/:ActiveKey", u.active)
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

func (u *userH) active(c *fiber.Ctx) error {
	return u.userI.Active(c)
}
