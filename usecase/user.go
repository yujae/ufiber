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
	Active(*fiber.Ctx) error
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
	input.ACTIVEKEY = strings.Replace(activekey, "-", "", -1)

	// JSON Unmarchal
	err := json.Unmarshal(c.Body(), &input)
	if err != nil {
		if er := u.userHistoryI.Create(&model.UserHistory{USER_ID: input.ID, MSG: err.Error()}); er != nil {
			return c.JSON(NewError(er))
		}
		return c.JSON(NewError(err))
	}

	// Email 검증
	if !isEmailValid(input.ID) {
		return c.JSON(NewError(ErrInvalidEmail))
	}

	// Nick 검증
	if len(input.NICK) < 3 {
		return c.JSON(NewError(ErrNickTooShort))
	}
	result, err := u.userI.RetrieveWithNick(input.NICK)
	if err != nil && !strings.Contains(err.Error(), "no row") {
		if er := u.userHistoryI.Create(&model.UserHistory{USER_ID: input.ID, MSG: err.Error()}); er != nil {
			return c.JSON(NewError(er))
		}
		//return c.JSON(NewError(ErrNickAlreadyExists))
	}
	if (model.User{}) != result {
		return c.JSON(NewError(ErrNickAlreadyExists))
	}

	// 비밀번호 암호화 & 생성
	genPW, err := bcrypt.GenerateFromPassword([]byte(input.PW), 10)
	if err != nil {
		if er := u.userHistoryI.Create(&model.UserHistory{USER_ID: input.ID, MSG: err.Error()}); er != nil {
			return c.JSON(NewError(er))
		}
		return c.JSON(NewError(err))
	}
	input.PW = string(genPW)

	// User 생성
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

	home := c.Locals("Home").(string)
	err = SendMail(c, Mail{
		fromName: c.Locals("Mail.FromName").(string),
		fromMail: c.Locals("Mail.FromMail").(string),
		toName:   "",
		toMail:   input.ID,
		subject:  "[" + home + "] " + "가입 확인 메일입니다.",
		body:     "클릭해주십시오. <a href='https://api." + home + "/activate/" + input.ACTIVEKEY + "'>가입 확인</a>",
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

func (u *userU) Active(c *fiber.Ctx) error {
	activeKey := c.Params("ActiveKey")
	result, err := u.userI.UpdateActive(activeKey)
	if err != nil {
		if er := u.userHistoryI.Create(&model.UserHistory{USER_ID: "", MSG: err.Error()}); er != nil {
			return c.JSON(NewError(er))
		}
		return c.JSON(NewError(err))
	}

	rowCount, err := result.RowsAffected()
	if err != nil {
		return c.JSON(NewError(err))
	}
	if rowCount == 0 {
		return c.JSON(NewError(ErrActiveKeyNotFound))
	}

	home := c.Locals("Home").(string)
	return c.Redirect("https://" + home)
}

func (u *userU) Login(c *fiber.Ctx) error {
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
	if len(e) < 5 && len(e) > 254 {
		return false
	}
	return emailRegex.MatchString(e)
}
