package controllers

import (
	"01cloud-api/api/auth"
	"01cloud-api/api/mailer"
	"01cloud-api/api/models/doc"
	"01cloud-api/api/notifications"
	"01cloud-api/api/security"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"

	"01cloud-api/api/models"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/formaterror"
	"01cloud-api/api/utils/helper"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var userinviteInterface = models.NewUserInvite()

// SignupInvite godoc
// @Summary Signup Invite
// @Description Signup Invite
// @Tags UserInvite
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param token path string true "Signup Invite"
// @Success 200 {object} doc.User
// @Router /user-invite-register/{token} [post]
func (server *Server) SignupInvite(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invite, err := userinviteInterface.FindByToken(server.DB, vars["token"])
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
	}
	user := models.User{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	user.Prepare()
	err = user.Validate("")
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	user.IsAdmin = false
	user.Email = invite.Email
	user.EmailVerified = true
	config, err := helper.ReadAdminConfigFile()
	if err != nil {
		log.Error(err)
	}
	quotasBytes, err := json.Marshal(config.Quotas)
	if err != nil {
		log.Error("error while setting quotas", err)
	}
	user.Quotas.RawMessage = quotasBytes
	userCreated, err := userInterface.SaveUser(server.DB, &user)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	_, _ = userinviteInterface.Delete(server.DB, uint64(invite.ID))
	token, _ := auth.CreateToken(user.ID, 0, 0)
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, &map[string]interface{}{
			"user":  userCreated,
			"token": token,
		})
		return
	}
	userInviteResponse, err := CreateUserResponse(userCreated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data: &map[string]interface{}{
			"user":  userInviteResponse,
			"token": token,
		},
		Success: 1,
		Message: "Success",
	})
}

// UserRequestDemo godoc
// @Summary User Request Demo
// @Description User Request Demo
// @Tags UserInvite
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.UserInvite true "Update Request Demo"
// @Success 200
// @Router /user-invite-request [post]
func (server *Server) UserRequestDemo(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data := models.UserInvite{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data.Prepare()
	err = data.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	tokenMap := struct {
		Token string `json:"captcha_code"`
	}{}
	err = json.Unmarshal(body, &tokenMap)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid captcha. "))
		return
	}
	err = VerifyCaptcha(tokenMap.Token)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	user, err := userInterface.FindUserByEmail(server.DB, data.Email)
	if err == nil && user.ID != 0 {
		responses.ERROR(w, http.StatusInternalServerError, errors.New("email already exists "))
		return
	}
	dataRequest, err := userinviteInterface.Save(server.DB, data)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	approve := os.Getenv("USER_AUTO_APPROVE")
	if approve == "true" {
		user, err := userinviteInterface.Find(server.DB, uint64(dataRequest.ID))
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
		token := security.TokenHash(user.Email)
		user.Token = token
		email := mailer.SendInviteEmail(user.Email, user.Token)
		err = notifications.NotifyEmail(server.NotifyClient, &email)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		user.EmailSent = true
		_, err = userinviteInterface.Update(server.DB, *user)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
	} else {
		email := mailer.SendRequestDemoEmailToUser(dataRequest)
		err = notifications.NotifyEmail(server.NotifyClient, &email)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, errors.New("unable to find email"))
			return
		}
		reqEmail := mailer.SendRequestDemoEmailToAdmin(dataRequest)
		err = notifications.NotifyEmail(server.NotifyClient, &reqEmail)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, errors.New("unable to send request"))
			return
		}
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, map[string]interface{}{
			"message": "Request submitted, Our team will review your request and let you know soon",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// UserValidateInvite godoc
// @Summary User Validate Invite
// @Description User Validate Invite
// @Tags UserInvite
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Update Validate Invite"
// @Success 200
// @Router /user-invite-validate/{id} [post]
func (server *Server) UserValidateInvite(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data, err := userinviteInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	token := security.TokenHash(data.Email)
	data.Token = token
	email := mailer.SendInviteEmail(data.Email, data.Token)
	err = notifications.NotifyEmail(server.NotifyClient, &email)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data.EmailSent = true
	_, err = userinviteInterface.Update(server.DB, *data)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]interface{}{
			"message": "Sent invitation email for " + data.Email,
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// CreateUserInvite godoc
// @Summary Create User Invite
// @Description Create User Invite
// @Tags UserInvite
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.UserInvite true "Create User Invite"
// @Success 201 {object} doc.UserInvite
// @Router /user-invite [post]
func (server *Server) CreateUserInvite(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data := models.UserInvite{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data.Prepare()
	err = data.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusNotAcceptable, err)
		return
	}
	_, err = userinviteInterface.FindByEmail(server.DB, data.Email)
	if err == nil {
		responses.ERROR(w, http.StatusNotAcceptable, errors.New("already invited to this email "))
		return
	}
	token := security.TokenHash(data.Email)
	data.Token = token
	email := mailer.SendInviteEmail(data.Email, data.Token)
	err = notifications.NotifyEmail(server.NotifyClient, &email)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data.EmailSent = true
	dataCreated, err := userinviteInterface.Save(server.DB, data)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, dataCreated)
		return
	}
	userInviteResponse, err := CreateUserInviteResponse(dataCreated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    userInviteResponse,
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) ResendUserInvite(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data, err := userinviteInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	email := mailer.SendInviteEmail(data.Email, data.Token)
	err = notifications.NotifyEmail(server.NotifyClient, &email)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// GetUserInvites godoc
// @Summary Get User Invites
// @Description Get User Invites
// @Tags UserInvite
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {array} doc.UserInvite
// @Router /user-invites [get]
func (server *Server) GetUserInvites(w http.ResponseWriter, r *http.Request) {
	datas, err := userinviteInterface.FindAll(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, datas)
		return
	}
	userInviteResponse := []doc.UserInvite{}
	for _, item := range *datas {
		response, err := CreateUserInviteResponse(&item)
		if err != nil {
			log.Error(err)
			continue
		}
		userInviteResponse = append(userInviteResponse, *response)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    userInviteResponse,
		Success: 1,
		Message: "Success",
	})
}

// GetUserInvite godoc
// @Summary Get User Invite
// @Description Get User Invite
// @Tags UserInvite
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param token path string true "Get User Invite"
// @Success 200 {object} doc.UserInvite
// @Router /user-invite-register/{token} [get]
func (server *Server) GetUserInvite(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dataReceived, err := userinviteInterface.FindByToken(server.DB, vars["token"])
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataReceived)
		return
	}
	userInviteResponse, err := CreateUserInviteResponse(dataReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    userInviteResponse,
		Success: 1,
		Message: "Success",
	})
}

// GetUserInvite godoc
// @Summary Get User Invite
// @Description Get User Invite
// @Tags UserInvite
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Get User Invite"
// @Param body body doc.UserInvite true "Get User Invite"
// @Success 200 {object} doc.UserInvite
// @Router /user-invite/{id} [put]
func (server *Server) UpdateUserInvite(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data, err := userinviteInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("UserInvite not found"))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	dataUpdate := models.UserInvite{}
	err = json.Unmarshal(body, &dataUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	dataUpdate.Prepare()
	err = dataUpdate.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	dataUpdate.ID = data.ID
	dataUpdated, err := userinviteInterface.Update(server.DB, dataUpdate)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataUpdated)
		return
	}
	userInviteResponse, err := CreateUserInviteResponse(dataUpdated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    userInviteResponse,
		Success: 1,
		Message: "Success",
	})
}

// DeleteUserInvite godoc
// @Summary Delete User Invite
// @Description Delete User Invite
// @Tags UserInvite
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Delete User Invite"
// @Success 200
// @Router /user-invite/{id} [delete]
func (server *Server) DeleteUserInvite(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	_, err = userinviteInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("Unauthorized"))
		return
	}
	_, err = userinviteInterface.Delete(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusNoContent, "")
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

func CreateUserInviteResponse(userInvite *models.UserInvite) (*doc.UserInvite, error) {
	userInviteResponse := doc.UserInvite{}
	userInviteBytes, _ := json.Marshal(userInvite)
	err := json.Unmarshal(userInviteBytes, &userInviteResponse)
	if err != nil {
		return nil, err
	}
	return &userInviteResponse, nil
}
