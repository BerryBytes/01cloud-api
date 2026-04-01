package controllers

import (
	"01cloud-api/api/auth"
	"01cloud-api/api/models"
	"01cloud-api/api/models/doc"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/helper"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var tokenInterface = models.NewToken()

type tokenRequest struct {
	Name       string `json:"name"`
	ExpiryDate string `json:"expiry_date"`
}

// CreateToken godoc
// @Summary Create Token new subscription
// @Description Create Token
// @Tags Token
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body tokenRequest true "Create Token"
// @Success 201 {object} doc.Token
// @Router /user/token [post]
func (server *Server) CreateToken(w http.ResponseWriter, r *http.Request) {
	uid, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	tokenReq := tokenRequest{}
	err = json.Unmarshal(body, &tokenReq)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	token := models.Token{}
	if tokenReq.Name == "" {
		responses.ERROR(w, http.StatusBadRequest, errors.New("name mustn't be empty"))
		return
	}
	if tokenReq.ExpiryDate != "" {
		expiryDate, err := time.Parse("2006-01-02", tokenReq.ExpiryDate)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		token.ExpiryDate = &expiryDate
	}
	token.Name = tokenReq.Name
	token.UserId = uint64(uid)
	token.Token = helper.GenerateToken()
	isExist := tokenInterface.IsTokenExistsByName(server.DB, &token)
	if isExist {
		responses.ERROR(w, http.StatusInternalServerError, errors.New("name already exists"))
		return
	}
	data, err := tokenInterface.CreateToken(server.DB, &token)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, data)
		return
	}
	tokenResponse, err := CreateTokenResponse(data)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    tokenResponse,
		Success: 1,
		Message: "Success",
	})
}

// RevokeAllToken godoc
// @Summary Revoke All Token
// @Description Revoke All Token
// @Tags Token
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200
// @Router /user/tokens [delete]
func (server *Server) RevokeAllToken(w http.ResponseWriter, r *http.Request) {
	uid, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	err = tokenInterface.RevokeAllToken(server.DB, uint64(uid))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
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

// GetTokens godoc
// @Summary Get Tokens
// @Description Get All Tokens
// @Tags Token
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {array} doc.Token
// @Router /user/tokens [get]
func (server *Server) GetTokens(w http.ResponseWriter, r *http.Request) {
	uid, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	tokenData, err := tokenInterface.FindAllByUserId(server.DB, uint64(uid))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	modifiedToken := []models.Token{}
	for _, val := range *tokenData {
		val.Token = ""
		modifiedToken = append(modifiedToken, val)
	}
	tokenResponse := []doc.Token{}
	for _, item := range modifiedToken {
		response, err := CreateTokenResponse(&item)
		if err != nil {
			log.Error(err)
			continue
		}
		tokenResponse = append(tokenResponse, *response)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    tokenResponse,
		Success: 1,
		Message: "Success",
	})
}

// RevokeSingleToken godoc
// @Summary Revoke Single Token
// @Description Revoke Single Token
// @Tags Token
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Revoke Token"
// @Success 200
// @Router /user/token/{id} [delete]
func (server *Server) RevokeSingleToken(w http.ResponseWriter, r *http.Request) {
	_, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	vars := mux.Vars(r)
	tokenId, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	err = tokenInterface.RevokeSingleToken(server.DB, uint(tokenId))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
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

func CreateTokenResponse(token *models.Token) (*doc.Token, error) {
	tokenResponse := doc.Token{}
	tokenBytes, _ := json.Marshal(token)
	err := json.Unmarshal(tokenBytes, &tokenResponse)
	if err != nil {
		return nil, err
	}
	return &tokenResponse, nil
}
