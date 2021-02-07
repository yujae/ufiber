package main

import (
	"database/sql"
	"fmt"
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v2"
	_ "github.com/lib/pq"
	"log"
	"ufiber/handler"
	"ufiber/repository"
	"ufiber/usecase"
)

// 인증 서버

// 클라이언트 (브라우저)
// cookie 저장 HTTPONLY, secure 조치 필요

func main() {
	app, db := setup()
	defer db.Close()

	userR := repository.NewUserR(db)
	userHistoryR := repository.NewUserHistoryR(db)

	userU := usecase.NewUserU(userR, userHistoryR)
	userH := handler.NewUserH(userU)
	userH.Router(app.Group("/"))

	log.Fatal(app.Listen(":80"))
}

func setup() (*fiber.App, *sql.DB) {
	conf, err := NewConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	URL := fmt.Sprintf("%s", conf.DB[0].URL)
	db, err := sql.Open(conf.DB[0].DriverName, URL) // DB 설정
	if db != nil {
		db.SetMaxIdleConns(10)
		db.SetMaxOpenConns(10)
	}
	if err != nil {
		log.Fatalf("sql.Open() Error : %v", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatalf("db.Ping() Error : %v", err)
	}

	app := fiber.New(fiber.Config{
		StrictRouting: true, // 허용: /foo, 불허:/foo/
		CaseSensitive: true, // 대소문자 구분
		//Immutable: true,				// Default: false
		ErrorHandler: nil, // Default: DefaultErrorHandler
	})

	app.Use(jwtware.New(jwtware.Config{
		Filter:         skip,                       // For Skip, Filtering Func
		SuccessHandler: nil,                        // SuccessHandler Func
		ErrorHandler:   nil,                        // ErrorHandler Func
		SigningKey:     []byte(conf.JwtSigningKey), // Require
		SigningMethod:  "",                         // Default: "HS256" (HS384, HS512, ES256, ES384, ES512, RS256, RS384, RS512)
		ContextKey:     "",                         // Default: "user"
		Claims:         nil,                        // Default: jwt.MapClaims{}
		TokenLookup:    "",                         // Default: "header:Authorization" (query:<name>, param:<name>, cookie:<name>)
		AuthScheme:     "",                         // Default: "Bearer"
	}))

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("JwtSigningKey", conf.JwtSigningKey)
		c.Locals("ActivateKey", conf.ActivateKey)
		c.Locals("AccessKeyExpiredSec", conf.AccessKeyExpiredSec)
		c.Locals("RefreshKeyExpiredSec", conf.RefreshKeyExpiredSec)

		c.Locals("Mail.Host", conf.Mail[0].Host)
		c.Locals("Mail.Port", conf.Mail[0].Port)
		c.Locals("Mail.ID", conf.Mail[0].ID)
		c.Locals("Mail.PW", conf.Mail[0].PW)
		c.Locals("Mail.FromName", conf.Mail[0].FromName)
		c.Locals("Mail.FromMail", conf.Mail[0].FromMail)
		return c.Next()
	})

	return app, db
}

func skip(c *fiber.Ctx) bool {
	if c.Method() == "POST" {
		if c.Path() == "/login" || c.Path() == "/register" {
			return true
		}
	}
	return false
}
