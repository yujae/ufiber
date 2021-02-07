package usecase

import (
	"encoding/json"
	"github.com/form3tech-oss/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"regexp"
	"strings"
	"time"
	"ufiber/model"
	"ufiber/repository"
)

type UserI interface {
	Register(*fiber.Ctx) error
	Login(*fiber.Ctx) error
	ReToken(*fiber.Ctx) error
}

type userU struct {
	userI        repository.UserI
	userHistoryI repository.UserHistoryI
}

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func NewUserU(userI repository.UserI, userHistoryI repository.UserHistoryI) *userU {
	return &userU{userI, userHistoryI}
}

func (u *userU) Register(c *fiber.Ctx) error {
	input := &model.User{}
	activekey := uuid.NewString()
	input.NICK = strings.Replace(activekey, "-", "", -1)
	input.ACTIVEKEY = strings.Replace(activekey, "-", "", -1)

	err := json.Unmarshal(c.Body(), &input)
	if err != nil {
		if er := u.userHistoryI.Create(&model.UserHistory{USER_ID: input.ID, MSG: err.Error()}); er != nil {
			return c.JSON(NewError(er))
		}
		return c.JSON(NewError(err))
	}

	if !isEmailValid(input.ID) {
		return c.JSON(NewError(ErrInvalidEmail))
	}

	genPW, err := bcrypt.GenerateFromPassword([]byte(input.PW), 10)
	if err != nil {
		if er := u.userHistoryI.Create(&model.UserHistory{USER_ID: input.ID, MSG: err.Error()}); er != nil {
			return c.JSON(NewError(er))
		}
		return c.JSON(NewError(err))
	}
	input.PW = string(genPW)

	err = u.userI.Create(input)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			err = ErrEmailAlreadyExists
		}
		if er := u.userHistoryI.Create(&model.UserHistory{USER_ID: input.ID, MSG: err.Error()}); er != nil {
			return c.JSON(NewError(er))
		}
		return c.JSON(NewError(err))
	}

	err = SendMail(c, Mail{
		fromName: c.Locals("Mail.FromName").(string),
		fromMail: c.Locals("Mail.FromMail").(string),
		toName:   "",
		toMail:   "master@dyonbe.com", //input.ID,
		subject:  "[Dyonbe] 가입 확인 메일입니다.",
		body:     "클릭해주십시오. <a href='https://api.dyonbe.com/activate/" + input.ACTIVEKEY + "'>가입 확인</a>",
	})
	if err != nil {
		if er := u.userHistoryI.Create(&model.UserHistory{USER_ID: input.ID, MSG: err.Error()}); er != nil {
			return c.JSON(NewError(er))
		}
		return c.JSON(NewError(err))
	}

	input.PW = ""
	if er := u.userHistoryI.Create(&model.UserHistory{USER_ID: input.ID, MSG: "Register"}); er != nil {
		return c.JSON(NewError(er))
	}
	return c.JSON(input)
}

func (u userU) Login(c *fiber.Ctx) error {
	//curl -v -X POST -H "User-Agent: linux bla bla" -H "Content-Type: application/json" -d " {\"id\":\"1\",\"pw\":\"1\"} " http://localhost/login
	input := &model.User{}

	err := json.Unmarshal(c.Body(), &input)
	if err != nil {
		return c.JSON(NewError(err))
	}

	result, err := u.userI.Retrieve(input.ID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			err = ErrEmailNotFound
		}
		if er := u.userHistoryI.Create(&model.UserHistory{USER_ID: result.ID, MSG: err.Error()}); er != nil {
			return c.JSON(NewError(er))
		}
		return c.JSON(NewError(err))
	}

	err = bcrypt.CompareHashAndPassword([]byte(result.PW), []byte(input.PW))
	if err != nil {
		if er := u.userHistoryI.Create(&model.UserHistory{USER_ID: result.ID, MSG: ErrPasswordIncorrect.Error()}); er != nil {
			return c.JSON(NewError(er))
		}
		return c.JSON(NewError(ErrPasswordIncorrect))
	}

	token, err := genToken(c, result)
	if err != nil {
		if er := u.userHistoryI.Create(&model.UserHistory{USER_ID: result.ID, MSG: err.Error()}); er != nil {
			return c.JSON(NewError(er))
		}
		return c.JSON(NewError(err))
	}

	if er := u.userHistoryI.Create(&model.UserHistory{USER_ID: result.ID, MSG: "Login"}); er != nil {
		return c.JSON(NewError(er))
	}

	err = u.userI.UpdateLogin(result.ID)
	if err != nil {
		if er := u.userHistoryI.Create(&model.UserHistory{USER_ID: result.ID, MSG: err.Error()}); er != nil {
			return c.JSON(NewError(er))
		}
		return c.JSON(NewError(err))
	}
	return c.JSON(token)
}

func (u userU) ReToken(c *fiber.Ctx) error {
	claims := c.Locals("user").(*jwt.Token).Claims.(jwt.MapClaims)

	result, err := u.userI.Retrieve(claims["id"].(string))
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			err = ErrEmailNotFound
		}
		if er := u.userHistoryI.Create(&model.UserHistory{USER_ID: result.ID, MSG: err.Error()}); er != nil {
			return c.JSON(NewError(er))
		}
		return c.JSON(NewError(err))
	}

	token, err := genToken(c, result)
	if err != nil {
		if er := u.userHistoryI.Create(&model.UserHistory{USER_ID: result.ID, MSG: err.Error()}); er != nil {
			return c.JSON(NewError(er))
		}
		return c.JSON(NewError(err))
	}

	if er := u.userHistoryI.Create(&model.UserHistory{USER_ID: result.ID, MSG: "ReToken"}); er != nil {
		return c.JSON(NewError(er))
	}
	return c.JSON(token)
}

func genToken(c *fiber.Ctx, u model.User) (fiber.Map, error) {
	SingingKey := c.Locals("JwtSigningKey").(string)
	tExp := time.Now().Add(time.Duration(c.Locals("AccessKeyExpiredSec").(int)) * time.Second).Unix()
	rtExp := time.Now().Add(time.Duration(c.Locals("RefreshKeyExpiredSec").(int)) * time.Second).Unix()

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = u.ID
	claims["act"] = u.ACTIVE // Activated
	claims["exp"] = tExp

	t, err := token.SignedString([]byte(SingingKey))
	if err != nil {
		return nil, err
	}

	refreshToken := jwt.New(jwt.SigningMethodHS256)
	rClaims := refreshToken.Claims.(jwt.MapClaims)
	rClaims["id"] = u.ID
	rClaims["act"] = u.ACTIVE // Activated
	rClaims["iat"] = tExp
	rClaims["exp"] = rtExp

	rt, err := refreshToken.SignedString([]byte(SingingKey))
	if err != nil {
		return nil, err
	}

	return fiber.Map{
		"access_token":  t,
		"refresh_token": rt,
	}, nil
}

func isEmailValid(e string) bool {
	if len(e) < 3 && len(e) > 254 {
		return false
	}
	return emailRegex.MatchString(e)
}
