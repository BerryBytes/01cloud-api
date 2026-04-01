package controllers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"01cloud-api/api/mailer"
	"01cloud-api/api/models"
	"01cloud-api/api/notifications"
	"01cloud-api/api/responses"
	"01cloud-api/api/security"
	"01cloud-api/api/utils/helper"
)

var resetPasswordInterface = models.NewResetPassword()

// ForgotPassword godoc
// @Summary Forgot Password
// @Description Forgot Password
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.User true "Forgot Password"
// @Success 200
// @Router /user/forgot-password [post]
func (server *Server) ForgotPassword(w http.ResponseWriter, r *http.Request) {
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
	err = user.Validate("forgotpassword")
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	err = server.DB.Model(models.User{}).Where("email = ?", user.Email).Take(&user).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	resetPassword := models.ResetPassword{}
	resetPassword.Prepare()

	token := security.TokenHash(user.Email)
	resetPassword.Email = user.Email
	resetPassword.Token = token

	resetDetails, err := resetPasswordInterface.SaveDatails(server.DB, &resetPassword)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	resetEmail, response := mailer.SendMail.SendResetPassword(resetDetails.Email, resetDetails.Token)
	err = notifications.NotifyEmail(server.NotifyClient, &resetEmail)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, response)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    response,
		Success: 1,
		Message: "Success",
	})
}

// ResetPassword godoc
// @Summary Reset Password
// @Description Reset Password
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {string} string
// @Router /user/reset-password [post]
func (server *Server) ResetPassword(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
	}
	requestBody := map[string]string{}
	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	user := models.User{}
	resetPassword := models.ResetPassword{}

	err = server.DB.Model(models.ResetPassword{}).Where("token = ?", requestBody["token"]).Take(&resetPassword).Error
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if resetPassword.DeletedAt != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("reset password link has been expired"))
		return
	}
	if requestBody["new_password"] == "" || requestBody["retype_password"] == "" {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("please ensure both field are entered"))
		return
	}
	if requestBody["new_password"] != "" && requestBody["retype_password"] != "" {
		if len(requestBody["new_password"]) < 6 || len(requestBody["retype_password"]) < 6 {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("password should be atleast 6 characters"))
			return
		}
		if requestBody["new_password"] != requestBody["retype_password"] {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("passwords provided do not match"))
			return
		}
		user.Password = requestBody["new_password"]
		user.Email = resetPassword.Email
		user.Prepare()
		err := userInterface.UpdatePassword(server.DB, &user)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("cannot save, Please try again later"))
			return
		}
		_, err = resetPasswordInterface.DeleteDetails(server.DB, &resetPassword)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("cannot delete record, please try again later"))
			return
		}
		if !helper.IsV2(r) {
			responses.JSON(w, http.StatusOK, &map[string]string{
				"message": "Password reset successful",
			})
			return
		}
		responses.JSON(w, http.StatusOK, responses.Response{
			Success: 1,
			Message: "Success",
		})
	}
}
