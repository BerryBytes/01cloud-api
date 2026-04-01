package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"01cloud-api/api/auth"
	"01cloud-api/api/mailer"
	"01cloud-api/api/models"
	"01cloud-api/api/models/doc"
	"01cloud-api/api/notifications"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/formaterror"

	"01cloud-api/api/utils/helper"

	"github.com/gorilla/mux"
)

var userInterface = models.NewUser()

// Register godoc
// @Summary register to 01cloud
// @Description register to 01cloud
// @Tags User
// @Accept  json
// @Produce  json
// @Param body body doc.UserRegister true "User register"
// @Param gitid query string true "Git Id"
// @Success 200 {object} doc.User
// @Router /user/register [post]

func (server *Server) CreateAuth0User(w http.ResponseWriter, r *http.Request) {

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

	if existingUser, _ := userInterface.FindUserByEmail(server.DB, user.Email); existingUser.ID != 0 {
		responses.ERROR(w, http.StatusConflict, errors.New("user already exists"))
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

	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, userCreated)
		return
	}

	userResponse, err := CreateUserResponse(userCreated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    userResponse,
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	gitid := r.URL.Query().Get("gitid")
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
	userCreated, err := userInterface.SaveUser(server.DB, &user)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	token, _ := auth.CreateToken(user.ID, 0, 0)
	email := mailer.SendVerification(userCreated.Email, token)
	err = notifications.NotifyEmail(server.NotifyClient, &email)
	if err != nil {
		log.Error("error from notify email :: ", err)
	}
	if err != nil {
		log.Error(err)
	}
	if gitid != "" {
		id, _ := strconv.ParseUint(gitid, 10, 64)
		git, err := gitUserInterface.FindByGitUserID(server.DB, id)
		if err != nil {
			log.Error(err)
		}
		git.UserID = uint64(userCreated.ID)
		_, err = gitUserInterface.Update(server.DB, git)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
		}

	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, userCreated)
		return
	}
	userResponse, err := CreateUserResponse(userCreated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    userResponse,
		Success: 1,
		Message: "Success",
	})
}

// ChangePassword godoc
// @Summary ChangePassword to 01cloud
// @Description ChangePassword to 01cloud
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Param body body doc.ChangePassword true "Change Password"
// @Success 200 {object} string
// @Router /user/change-password [post]
func (server *Server) ChangePassword(w http.ResponseWriter, r *http.Request) {
	uid, _, _ := auth.ExtractTokenID(r)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user Not Found"))
		return
	}
	requestBody := map[string]string{}
	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	if requestBody["new_password"] == "" || requestBody["retype_password"] == "" || requestBody["password"] == "" {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("all fields are required"))
		return
	}

	if len(requestBody["new_password"]) < 8 || len(requestBody["retype_password"]) < 8 {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("password should be atleast 8 characters"))
		return
	}
	if requestBody["password"] == requestBody["new_password"] {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("new Password cannot be same as Old Password"))
		return
	}
	if requestBody["new_password"] != requestBody["retype_password"] {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("passwords provided do not match"))
		return
	}
	err = models.VerifyPassword(user.Password, requestBody["password"])
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid password. "))
		return
	}
	user.Password = requestBody["new_password"]
	user.Prepare()
	err = userInterface.UpdatePassword(server.DB, user)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("cannot Save, Pls try again later"))
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]string{
			"message": "Password changed successfully",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Password changed successfully",
	})

}

// GetUsers godoc
// @Summary GetUsers of 01cloud for admin
// @Description GetUsers of 01cloud for admin
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param page query int true "Page"
// @Param size query int true "Size"
// @Param search query string true "Search"
// @Param sort-column query string true "Sort Column"
// @Param sort-direction query string true "Sort Direction"
// @Param body body doc.GetUsersRequest true "Users Filter"
// @Success 200 {array} doc.User
// @Router /users [get]
func (server *Server) GetUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.ParseUint(r.FormValue("page"), 10, 32)
	size, _ := strconv.ParseUint(r.FormValue("size"), 10, 32)
	search := r.FormValue("search")
	sortColumn := r.FormValue("sort-column")
	sortDirection := r.FormValue("sort-direction")
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 100
	}
	if sortColumn == "" {
		sortColumn = "id"
	}
	if sortDirection == "" {
		sortDirection = "desc"
	}
	users, count, err := userInterface.FindAllUsersWithFilters(server.DB, page, size, search, sortColumn, sortDirection)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"count": count,
			"data":  users,
		})
		return
	}
	userResponseList := []interface{}{}
	for _, item := range *users {
		userResponse, err := CreateUserResponse(&item)
		if err != nil {
			log.Error(err)
			continue
		}
		uJson, _ := helper.ToJson(userResponse)
		uJson["balance"] = paymentInterface.GetUserBalance(server.DB, item.ID)
		userResponseList = append(userResponseList, uJson)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    userResponseList,
		Success: 1,
		Message: "Success",
		Count:   count,
		Page:    int(page),
		Size:    int(size),
	})
}

// GetUser godoc
// @Summary GetUser of 01cloud for specific id
// @Description GetUser of 01cloud for specific id
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path string true "User Id"
// @Success 200 {object} doc.User
// @Router /user/{id} [get]
func (server *Server) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	key := fmt.Sprintf("user-%d", uid)
	var value interface{}
	if ok := server.Cache.Get(key, &value); ok {
		responses.JSON(w, http.StatusOK, value)
		return
	}
	userGotten, err := userInterface.FindUserByID(server.DB, uint(uid))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}

	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, userGotten)
		return
	}
	userResponse, err := CreateUserResponse(userGotten)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	response := responses.Response{
		Data:    userResponse,
		Success: 1,
		Message: "Success",
	}
	server.Cache.Set(key, response)
	responses.JSON(w, http.StatusOK, response)
}

// GetProfile godoc
// @Summary GetProfile of 01cloud for user
// @Description GetProfile of 01cloud for user
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} doc.User
// @Router /profile [get]
func (server *Server) GetProfile(w http.ResponseWriter, r *http.Request) {
	uid, _, _ := auth.ExtractTokenID(r)
	userGotten, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		userJson, _ := userGotten.ToJson()
		userJson["balance"] = paymentInterface.GetUserBalance(server.DB, uid)
		responses.JSON(w, http.StatusOK, userJson)
		return
	}
	userResponse, err := CreateUserResponse(userGotten)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	uJson, _ := helper.ToJson(userResponse)
	uJson["balance"] = paymentInterface.GetUserBalance(server.DB, uid)
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    uJson,
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
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
	tokenID, _, _ := auth.ExtractTokenID(r)
	if tokenID != uint(uid) {
		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}
	user.Prepare()
	//err = user.Validate("update")
	//if err != nil {
	//	responses.ERROR(w, http.StatusUnprocessableEntity, err)
	//	return
	//}
	updatedUser, err := userInterface.UpdateAUser(server.DB, uint(uid), &user)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, updatedUser)
		return
	}
	userResponse, err := CreateUserResponse(updatedUser)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    userResponse,
		Success: 1,
		Message: "Success",
	})
}

// UpdateProfile godoc
// @Summary UpdateProfile of 01cloud for user
// @Description UpdateProfile of 01cloud for user
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.User true "User details"
// @Success 200 {object} doc.User
// @Router /profile [put]
func (server *Server) UpdateProfile(w http.ResponseWriter, r *http.Request) {
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
	uid, _, _ := auth.ExtractTokenID(r)
	user.UpdatedAt = time.Now()
	updatedUser, err := userInterface.UpdateAUser(server.DB, uid, &user)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, updatedUser)
		return
	}
	userResponse, err := CreateUserResponse(updatedUser)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    userResponse,
		Success: 1,
		Message: "Success",
	})

}

// BlockUnBlockAccount godoc
// @Summary BlockUnBlockAccount of 01cloud for user
// @Description BlockUnBlockAccount of 01cloud for user
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "User Id"
// @Param type query string true "Type"
// @Success 200 {object} string
// @Router /user/{id}/block [get]
func (server *Server) BlockUnBlockAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, errors.New("invalid user id"))
		return
	}
	activeType := r.FormValue("type")
	if activeType == "" {
		responses.ERROR(w, http.StatusBadRequest, errors.New("type field is required"))
		return
	}
	userGotten, err := userInterface.FindUserByID(server.DB, uint(userId))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	err = userGotten.ChangeActive(server.DB, activeType == "unblock")

	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	server.ActiveDeactiveProjects(uint(userId), 0, activeType == "unblock")
	server.ActivateDeactiveOrganization(uint(userId), activeType == "unblock")
	email := mailer.SendBlockUnblockEmail(userGotten.Email, activeType)
	err = notifications.NotifyEmail(server.NotifyClient, &email)
	if err != nil {
		log.Error("error from notify email :: ", err)
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]interface{}{
			"message": "User " + activeType + "ed",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "User " + activeType + "ed",
	})
}

// ChangeAdminStatus godoc
// @Summary ChangeAdminStatus of 01cloud for user
// @Description ChangeAdminStatus of 01cloud for user
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param is_admin query bool true "Is Admin"
// @Success 200 {object} string
// @Router /user/{id}/change-admin-status [get]
func (server *Server) ChangeAdminStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, errors.New("invalid user id"))
		return
	}
	isAdmin, err := strconv.ParseBool(r.FormValue("is_admin"))
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("parse error"))
		return
	}
	userGotten, err := userInterface.FindUserByID(server.DB, uint(userId))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	err = userGotten.ChangeAdmin(server.DB, isAdmin)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]interface{}{
			"message": "Status changed successfully",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Status changed successfully",
	})
}

// DeactivateAccount godoc
// @Summary DeactivateAccount of 01cloud for user
// @Description DeactivateAccount of 01cloud for user
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} string
// @Router /user/account/deactivate [get]
func (server *Server) DeactivateAccount(w http.ResponseWriter, r *http.Request) {
	uid, _, _ := auth.ExtractTokenID(r)
	userGotten, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	err = userGotten.ChangeActive(server.DB, false)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	server.ActiveDeactiveProjects(uid, 0, false)
	server.ActivateDeactiveOrganization(uid, false)
	deactivateEmail := mailer.SendDeactivateAccountEmail(userGotten.Email)
	err = notifications.NotifyEmail(server.NotifyClient, &deactivateEmail)
	if err != nil {
		log.Error("error from notify email :: ", err)
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]interface{}{
			"message": "User deactivated",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "User deactivated",
	})
}

func (server *Server) ActivateDeactiveOrganization(uid uint, active bool) {
	organizations, err := orgInterface.FindAll(server.DB, uid, !active, "")
	if err != nil {
		log.Error(err)
		return
	}
	for _, org := range *organizations {
		_, _ = orgInterface.ActiveDeactiveOrganization(server.DB, org.ID, active)
		server.ActiveDeactiveProjects(uid, org.ID, active)
	}
}

func (server *Server) ActiveDeactiveProjects(uid, oid uint, active bool) {
	datas, err := iproject.FindUserProjectOnly(server.DB, uid, oid)
	if err != nil {
		log.Error(err)
	}
	for _, proj := range datas {
		err = server.projectActivation(proj.ID, uid, active)
		if err != nil {
			log.Error(err)
		}
	}

}

// DeleteUser godoc
// @Summary DeactivateAccount of 01cloud for user
// @Description DeactivateAccount of 01cloud for user
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path string true "User Id"
// @Success 200 {object} string
// @Router /user/{id} [delete]
func (server *Server) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	tokenID, _, _ := auth.ExtractTokenID(r)
	if tokenID != 0 && tokenID != uint(uid) {
		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}
	_, err = userInterface.DeleteAUser(server.DB, uint(uid))
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Entity", fmt.Sprintf("%d", uid))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusNoContent, "")
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) UpdateUserQuotas(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid, err := strconv.ParseUint(vars["uid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data := models.Quotas{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	user, err := userInterface.FindUserByID(server.DB, uint(uid))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user Not Found"))
		return
	}
	user.Quotas.RawMessage = body

	updatedUser, err := userInterface.UpdateUserQuota(server.DB, user)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	response, err := CreateUserResponse(updatedUser)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
		Data:    response,
	})

}

func CreateUserResponse(user *models.User) (*doc.User, error) {
	userResponse := doc.User{}
	userBytes, _ := json.Marshal(user)
	err := json.Unmarshal(userBytes, &userResponse)
	if err != nil {
		return nil, err
	}
	return &userResponse, nil

}

func VerifyCaptcha(token string) error {
	secretKey := os.Getenv("RECAPTCHA_SECRET")
	resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify",
		url.Values{"secret": {secretKey}, "response": {token}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyObj := struct {
			Success bool `json:"success"`
		}{}
		_ = json.Unmarshal(bodyBytes, &bodyObj)

		if !bodyObj.Success {
			return errors.New("invalid captcha. ")
		}
	} else {
		return errors.New("invalid captcha. ")

	}
	return nil
}

func (server *Server) GetAllSessionByUserId(w http.ResponseWriter, r *http.Request) {
	uid, _, _ := auth.ExtractTokenID(r)
	size, err := strconv.Atoi(r.FormValue("size"))
	if err != nil || size == 0 {
		size = 100
	}
	page, err := strconv.Atoi(r.FormValue("page"))
	if err != nil || page == 0 {
		page = 1
	}
	offset := (page - 1) * size
	datas, count, err := userInterface.GetAllSessionByUserId(server.DB, uint(uid), size, offset)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    datas,
		Success: 1,
		Message: "Success",
		Count:   count,
		Page:    page,
		Size:    size,
	})
}

func (server *Server) DeactivateSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	sessionData, err := userInterface.GetSessionById(server.DB, uint(id))
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
	session := auth.GetSession(r.Header.Get("User-Agent"), sessionData.UserId, "deactivate")
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
