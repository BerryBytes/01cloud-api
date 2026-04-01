package controllers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"01cloud-api/api/auth"
	"01cloud-api/api/mailer"
	"01cloud-api/api/models"
	"01cloud-api/api/notifications"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/formaterror"
	"01cloud-api/api/utils/git"
	"01cloud-api/api/utils/helper"
	"math/rand"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// Login godoc
// @Summary login to 01cloud
// @Description login to 01cloud
// @Tags User
// @Accept  json
// @Produce  json
// @Param body body doc.UserLogin true "Login Details"
// @Success 200 {object} doc.UserLoginResponse
// @Router /user/login [post]
func (server *Server) Login(w http.ResponseWriter, r *http.Request) {
	mode := strings.ToLower(os.Getenv("MODE"))

	if mode == "legacy" {
		server.loginLegacy(w, r)
		return
	}
	server.loginCurrent(w, r)
}

// loginLegacy handles the legacy login logic with password validation
func (server *Server) loginLegacy(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	user := models.User{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	user.Prepare()
	err = user.Validate("login")
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	userAgent := r.Header.Get("User-Agent")
	user_obj, err := userInterface.FindUserByEmail(server.DB, user.Email)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	err = models.VerifyPassword(user_obj.Password, user.Password)
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, formattedError)
		return
	}
	sessionData := auth.GetSession(userAgent, user_obj.ID, "login")
	sessionData.Active = true
	session, err := userInterface.SaveSession(server.DB, &sessionData)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	token, err := auth.CreateToken(user_obj.ID, 0, session.ID)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, formattedError)
		return
	}
	if !user_obj.EmailVerified {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("please verify your email"))
		return
	}

	if !user_obj.Active {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("user account is disabled , please contact 01cloud admin"))
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"user":  user_obj,
			"token": token,
		})
		return
	}
	userResponse, err := CreateUserResponse(user_obj)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data: &map[string]interface{}{
			"user":  userResponse,
			"token": token,
		},
		Success: 1,
		Message: "Success",
	})
}

// loginCurrent handles the current login logic with X-Email header
func (server *Server) loginCurrent(w http.ResponseWriter, r *http.Request) {
	email := r.Header.Get("X-Email")
	if email == "" {
		responses.ERROR(w, http.StatusBadRequest, errors.New("missing X-Email header"))
		return
	}

	userAgent := r.Header.Get("User-Agent")

	user_obj, err := userInterface.FindUserByEmail(server.DB, email)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		// User doesn't exist, create new user
		user := models.User{
			Email:          r.Header.Get("X-Email"),
			FirstName:      r.Header.Get("X-First-Name"),
			LastName:       r.Header.Get("X-Last-Name"),
			Image:          "",
			Company:        r.Header.Get("X-Company"),
			Designation:    r.Header.Get("X-Designation"),
			Reference:      r.Header.Get("X-Reference"),
			AddressUpdated: r.Header.Get("X-Address-Updated") == "true",
			UsedDemo:       r.Header.Get("X-Used-Demo") == "true",
			EmailVerified:  r.Header.Get("X-Email-Verified") == "true",
			Active:         r.Header.Get("X-Active") == "true",
			IsAdmin:        false, // forcefully set
			Password:       "dummy",
		}
		config, err := helper.ReadAdminConfigFile()
		if err != nil {
			log.Error(err)
		}
		quotasBytes, err := json.Marshal(config.Quotas)
		if err != nil {
			log.Error("error while setting quotas", err)
		}
		user.Quotas.RawMessage = quotasBytes
		log.Debug("quota bytes :: ", string(quotasBytes))

		user.Prepare()

		if err := user.Validate(""); err != nil {
			log.Infof("validation error: %s", err)
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}

		userCreated, err := userInterface.SaveUser(server.DB, &user)
		if err != nil {
			formattedError := formaterror.FormatError(err.Error())
			responses.ERROR(w, http.StatusInternalServerError, formattedError)
			return
		}

		token, _ := auth.CreateToken(user.ID, 0, 0)
		emailContent := mailer.SendVerification(userCreated.Email, token)
		if err := notifications.NotifyEmail(server.NotifyClient, &emailContent); err != nil {
			log.Error("error from notify email :: ", err)
		}

		user_obj = userCreated
	}

	if !user_obj.EmailVerified {
		if r.Header.Get("X-Email-Verified") == "true" {
			verified_user, err := userInterface.VerifyEmail(server.DB, user_obj.ID)
			if err != nil {
				responses.ERROR(w, http.StatusInternalServerError, err)
				return
			}
			user_obj = verified_user
		} else {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("please verify your email"))
			return
		}

	}

	if !user_obj.Active {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("user account is disabled, please contact 01cloud admin"))
		return
	}

	sessionData := auth.GetSession(userAgent, user_obj.ID, "login")
	sessionData.Active = true
	session, err := userInterface.SaveSession(server.DB, &sessionData)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	token, err := auth.CreateToken(user_obj.ID, 0, session.ID)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, formattedError)
		return
	}

	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"user":  user_obj,
			"token": token,
		})
		return
	}

	userResponse, err := CreateUserResponse(user_obj)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	responses.JSON(w, http.StatusOK, responses.Response{
		Data: &map[string]interface{}{
			"user":  userResponse,
			"token": token,
		},
		Success: 1,
		Message: "Success",
	})
}

// LoginExternal godoc
// @Summary login to 01cloud external
// @Description login to 01cloud with external service
// @Tags User
// @Accept  json
// @Produce  json
// @Param body body doc.ExternalLogin true "External Login Details"
// @Success 200 {object} doc.UserLoginResponse
// @Router /user/login/external [post]
func (server *Server) LoginExternal(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	externalLogin := models.ExternalLogin{}
	err = json.Unmarshal(body, &externalLogin)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	err = externalLogin.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	userAgent := r.Header.Get("User-Agent")
	if externalLogin.Service == "github" {
		usr, accessToken, err := git.GithubUserLogin(externalLogin)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, err)
			return
		}
		user_obj, err := userInterface.FindUserByEmail(server.DB, usr.GetEmail())
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, err)
			return
		}
		sessionData := auth.GetSession(userAgent, user_obj.ID, "login")
		sessionData.Active = true
		session, err := userInterface.SaveSession(server.DB, &sessionData)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
		token, err := auth.CreateToken(user_obj.ID, 0, session.ID)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, err)
			return
		}
		if user_obj == nil {
			gitUser := models.GitUser{
				ServiceUserName: *usr.Login,
				GitUserID:       uint64(*usr.ID),
				AccessToken:     accessToken,
				ServiceName:     externalLogin.Service,
			}
			_, err := gitUserInterface.Save(server.DB, &gitUser)
			if err != nil {
				responses.ERROR(w, http.StatusNotFound, err)
				return
			}
			if !helper.IsV2(r) {
				responses.JSON(w, http.StatusNonAuthoritativeInfo, &map[string]interface{}{
					"message": "user not found",
					"user":    usr,
				})
				return
			}
			responses.JSON(w, http.StatusNonAuthoritativeInfo, responses.Response{
				Data: &map[string]interface{}{
					"user": usr,
				},
				Success: 0,
				Message: "user not registered",
			})
			return
		}
		if !user_obj.EmailVerified {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("please verify your email"))
			return
		}
		if !user_obj.Active {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("user account is disabled , please contact 01cloud admin"))
			return
		}
		user_obj.Password = ""
		gitUser := models.GitUser{
			ServiceUserName: *usr.Login,
			GitUserID:       uint64(*usr.ID),
			AccessToken:     accessToken,
			ServiceName:     externalLogin.Service,
			UserID:          uint64(user_obj.ID),
		}

		//gtUsr := models.GitUser{}
		oldUser, err := gitUserInterface.FindByUserIdAndService(server.DB, uint64(user_obj.ID), externalLogin.Service)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, err)
			return
		}
		if oldUser != nil {
			gitUser.ID = oldUser.ID
			_, err = gitUserInterface.Update(server.DB, &gitUser)
		} else {
			_, err = gitUserInterface.Save(server.DB, &gitUser)
		}
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
		if !helper.IsV2(r) {
			responses.JSON(w, http.StatusOK, &map[string]interface{}{
				"user":  user_obj,
				"token": token,
			})
			return
		}
		userResponse, err := CreateUserResponse(user_obj)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return

		}
		responses.JSON(w, http.StatusOK, responses.Response{
			Data: &map[string]interface{}{
				"user":  userResponse,
				"token": token,
			},
			Success: 1,
			Message: "Success",
		})
	} else if externalLogin.Service == "gitlab" {
		usr, err := git.GitlabUserLogin(externalLogin)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, err)
			return
		}
		user_obj, err := userInterface.FindUserByEmail(server.DB, usr.Email)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, err)
			return
		}
		sessionData := auth.GetSession(userAgent, user_obj.ID, "login")
		sessionData.Active = true
		session, err := userInterface.SaveSession(server.DB, &sessionData)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
		token, err := auth.CreateToken(user_obj.ID, 0, session.ID)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, err)
			return
		}
		if user_obj == nil {
			responses.JSON(w, http.StatusNonAuthoritativeInfo, &map[string]interface{}{
				"message": "user not found",
				"user":    usr,
			})
			return
		}
		if !user_obj.EmailVerified {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("please verify your email"))
			return
		}
		if !user_obj.Active {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("user account is disabled , please contact 01cloud admin"))
			return
		}
		user_obj.Password = ""
		if !helper.IsV2(r) {
			responses.JSON(w, http.StatusOK, &map[string]interface{}{
				"user":  user_obj,
				"token": token,
			})
			return
		}
		userResponse, err := CreateUserResponse(user_obj)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return

		}
		responses.JSON(w, http.StatusOK, responses.Response{
			Data: &map[string]interface{}{
				"user":  userResponse,
				"token": token,
			},
			Success: 1,
			Message: "Success",
		})

	} else {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("service not available"))
	}

}

// ResendVerificationEmail godoc
// @Summary send verification email
// @Description send verification email
// @Tags User
// @Accept  json
// @Produce  json
// @Param body body doc.EmailRequest true "External Login Details"
// @Success 200 {object} string
// @Router /user/resend-verification [post]
func (server *Server) ResendVerificationEmail(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	em := map[string]interface{}{}
	err = json.Unmarshal(body, &em)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("email is required"))
		return
	}
	email := em["email"]
	user := models.User{}

	err = server.DB.Model(models.User{}).Where("email = ?", email).Take(&user).Error
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	if !user.EmailVerified {
		token, _ := auth.CreateToken(user.ID, 0, 0)
		verfiyEmail := mailer.SendVerification(user.Email, token)
		err = notifications.NotifyEmail(server.NotifyClient, &verfiyEmail)
		if err == nil {
			if !helper.IsV2(r) {
				responses.JSON(w, http.StatusOK, map[string]string{"message": "Verification sent, Please check your email."})
				return
			}
			responses.JSON(w, http.StatusOK, responses.Response{
				Success: 1,
				Message: "Success",
			})
			return
		}
	}
	responses.ERROR(w, http.StatusBadRequest, err)
}

// VerifyEmail godoc
// @Summary verify email
// @Description verify email
// @Tags User
// @Accept  json
// @Produce  json
// @Param token path string true "Email Token"
// @Success 200 {object} string
// @Router /user/verify/{token} [post]
func (server *Server) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]
	uid, _, err := auth.ExtractIdFromToken(token)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	userGotten, err := userInterface.VerifyEmail(server.DB, uint(uid))
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"message": "Email verified",
			"user":    userGotten,
		})
		return
	}
	userResponse, err := CreateUserResponse(userGotten)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data: &map[string]interface{}{
			"message": "Email verified",
			"user":    userResponse,
		},
		Success: 1,
		Message: "Success",
	})
}

//
//func (server *Server) GithubCallback() http.Handler {
//	fn := func(w http.ResponseWriter, req *http.Request) {
//		ctx := req.Context()
//		githubUser, err := githublogin.UserFromContext(ctx)
//		if err != nil {
//			fmt.Println(err)
//			http.Error(w, err.Error(), http.StatusInternalServerError)
//			return
//		}
//
//		fmt.Printf("User id %s", githubUser.GetEmail())
//		http.Redirect(w, req, "/", http.StatusFound)
//	}
//	return http.HandlerFunc(fn)
//}

func (server *Server) ShadowUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid, err := strconv.ParseUint(vars["uid"], 10, 32)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	userGotten, err := userInterface.FindUserByID(server.DB, uint(uid))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}

	token, err := auth.CreateToken(userGotten.ID, 0, 0)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	if !userGotten.Active {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("user account is disabled"))
		return
	}
	userGotten.Password = ""
	userResponse, err := CreateUserResponse(userGotten)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data: &map[string]interface{}{
			"user":  userResponse,
			"token": token,
		},
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) Logout(w http.ResponseWriter, r *http.Request) {
	uid, _, _ := auth.ExtractTokenID(r)
	sessionId, err := auth.ExtractSessionID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("invalid session"))
		return
	}
	sessionData, err := userInterface.GetSessionById(server.DB, sessionId)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("invalid session"))
		return
	}
	sessionData.Active = false
	err = userInterface.UpdateSession(server.DB, *sessionData)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	session := auth.GetSession(r.Header.Get("User-Agent"), uid, "logout")
	_, err = userInterface.SaveSession(server.DB, &session)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) LogoutAll(w http.ResponseWriter, r *http.Request) {
	uid, _, _ := auth.ExtractTokenID(r)
	sessionData, _, err := userInterface.GetActiveSessionByUserId(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("invalid session"))
		return
	}
	for _, data := range *sessionData {
		data.Active = false
		err = userInterface.UpdateSession(server.DB, data)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
	}
	session := auth.GetSession(r.Header.Get("User-Agent"), uid, "logout-all")
	_, err = userInterface.SaveSession(server.DB, &session)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

var allSSOCode = make(map[string]string)
var allAuth0Tokens = make(map[string]string)

func (server *Server) CreateNewSSOCode(w http.ResponseWriter, _ *http.Request) {
	// generate a random number of 8 characters
	code := generateRandomCode()
	for {
		_, ok := allSSOCode[code]
		if !ok {
			break
		}
		code = generateRandomCode()
	}

	allSSOCode[code] = ""
	allAuth0Tokens[code] = ""

	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
		Data: &map[string]interface{}{
			"code": code,
		},
	})
	go func() {
		time.Sleep(5 * time.Minute)
		delete(allSSOCode, code)
		delete(allAuth0Tokens, code)
	}()
}

func generateRandomCode() string {
	newCode := ""
	for i := 0; i < 6; i++ {
		newCode += strconv.Itoa(rand.Intn(10))
	}
	return newCode
}

func (server *Server) CheckSSOStatus(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	codeReq := struct {
		Code string `json:"code"`
	}{}
	err = json.Unmarshal(body, &codeReq)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	val, ok := allSSOCode[codeReq.Code]
	auth0Token, tokenOk := allAuth0Tokens[codeReq.Code]
	if !ok || !tokenOk {
		responses.ERROR(w, http.StatusNotFound, errors.New("no such code"))
		return
	}
	if val == "" {
		responses.JSON(w, http.StatusOK, responses.Response{
			Success: 1,
			Message: "code not verified",
		})
		return
	}

	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
		Data: &map[string]interface{}{
			"token":       val,
			"auth0_token": auth0Token,
		},
	})
	delete(allSSOCode, codeReq.Code)
	delete(allAuth0Tokens, codeReq.Code)
}

func (server *Server) SubmitSSOCode(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	codeReq := struct {
		Code string `json:"code"`
	}{}
	err = json.Unmarshal(body, &codeReq)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	val, ok := allSSOCode[codeReq.Code]
	if !ok {
		responses.ERROR(w, http.StatusNotFound, errors.New("no such code"))
		return
	}

	if val != "" {
		responses.ERROR(w, http.StatusBadRequest, errors.New("already submitted"))
		return
	}

	auth0Token := r.Header.Get("X-Custom-Auth")
	allSSOCode[codeReq.Code] = auth.ExtractToken(r)
	allAuth0Tokens[codeReq.Code] = auth0Token
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}
