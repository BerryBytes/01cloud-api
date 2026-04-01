package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"01cloud-api/api/auth"
	"01cloud-api/api/models/doc"
	"01cloud-api/api/notifications"
	pb "01cloud-api/api/notifications/proto"
	"01cloud-api/api/utils/constants"
	"01cloud-api/api/utils/helper"

	"google.golang.org/grpc"

	"01cloud-api/api/responses"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// CreateNotification godoc
// @Summary Create Notification
// @Description Create Notification
// @Tags Notification
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.Notification true "Create Notification"
// @Success  201 {object} doc.Notification
// @Router /notification [post]
func (server *Server) CreateNotification(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data := &notifications.Notification{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	mes := &pb.Message{}
	data.ToGrpc(mes)
	c := pb.NewNotificationClient(server.NotifyClient)

	gr, err := c.Post(context.Background(), &pb.PostRequest{Message: mes})
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	res := &notifications.Notification{}
	res.ToGrpc(gr.GetMessage())
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, res)
		return
	}
	notificationResponse, err := CreateNotificationResponse(res)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    notificationResponse,
		Success: 1,
		Message: "Success",
	})
}

// PublishNotification godoc
// @Summary Publish Notification
// @Description Publish Notification
// @Tags Notification
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.Notification true "Publish Notification"
// @Success  201 {object} doc.Notification
// @Router /notification/publish [post]
func (server *Server) PublishNotification(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data := &notifications.Notification{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	res, err := PublishNotificationBase(server.NotifyClient, data)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, res)
		return
	}
	notificationResponse, err := CreateNotificationResponse(res)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    notificationResponse,
		Success: 1,
		Message: "Success",
	})
}

func PublishNotificationBase(conn *grpc.ClientConn, data *notifications.Notification) (*notifications.Notification, error) {
	mes := &pb.Message{}
	data.ToGrpc(mes)
	c := pb.NewNotificationClient(conn)

	gr, err := c.Publish(context.Background(), &pb.PostRequest{Message: mes})
	if err != nil {
		return data, err
	}
	res := &notifications.Notification{}
	res.ToGrpc(gr.GetMessage())
	return res, nil
}

// GetNotifications godoc
// @Summary Get notifications
// @Description Get list of notifications
// @Tags Notification
// @Accept  json
// @Produce  json
// @Param page path int true "Page"
// @Param limit path int true "Limit"
// @Param filter path string true "Filter"
// @Security ApiKeyAuth
// @Success 200 {array} doc.Notification
// @Router /notifications [get]
func (server *Server) GetNotifications(w http.ResponseWriter, r *http.Request) {
	c := pb.NewNotificationClient(server.NotifyClient)
	offset, err := strconv.ParseUint(r.FormValue("page"), 10, 64)
	if err != nil {
		offset = 0
	} else {
		offset = offset - 1
	}
	limit, err := strconv.ParseUint(r.FormValue("limit"), 10, 64)
	if err != nil {
		limit = 100
	}
	offset = offset * limit
	filter := r.FormValue("filter")
	if filter == "" {
		filter = constants.All
	}
	uid, _, _ := auth.ExtractTokenID(r)
	key := fmt.Sprintf("notifications-list-%d-%d-%d-%s", uid, offset, limit, filter)
	var value interface{}
	if ok := server.Cache.Get(key, &value); ok {
		responses.JSON(w, http.StatusOK, value)
		return
	}
	q := &pb.QueryRequest{User: int64(uid), Offset: int64(offset), Limit: int64(limit), Filter: filter}
	gr, err := c.Query(context.Background(), q)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	datas := []*notifications.Notification{}
	for _, s := range gr.GetMessages() {
		data := &notifications.Notification{}
		data.FromGrpc(s)
		datas = append(datas, data)
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, datas)
		return
	}
	notificationResponseList := []doc.Notification{}
	for _, notification := range datas {
		notificationResponse, err := CreateNotificationResponse(notification)
		if err != nil {
			log.Error(err)
		}
		notificationResponseList = append(notificationResponseList, *notificationResponse)
	}
	response := responses.Response{
		Data:    notificationResponseList,
		Success: 1,
		Message: "Success",
	}
	server.Cache.Set(key, response)
	responses.JSON(w, http.StatusOK, response)
}

// GetNotificationsCount godoc
// @Summary Get notifications count
// @Description Get count of unseen notifications
// @Tags Notification
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {array} doc.NotificationCountResponse
// @Router /notification/get-unseen-count [get]
func (server *Server) GetNotificationsCount(w http.ResponseWriter, r *http.Request) {
	c := pb.NewNotificationClient(server.NotifyClient)

	uid, _, _ := auth.ExtractTokenID(r)
	key := fmt.Sprintf("notifications-count-%d", uid)
	var value interface{}
	if ok := server.Cache.Get(key, &value); ok {
		responses.JSON(w, http.StatusOK, value)
		return
	}
	gr, err := c.Query(context.Background(), &pb.QueryRequest{User: int64(uid), Offset: int64(0), Limit: int64(99), Filter: constants.Unseen})
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	count := "99+"
	if len(gr.GetMessages()) < 99 {
		count = fmt.Sprint(len(gr.GetMessages()))
	}
	response := doc.NotificationCountResponse{
		ShowBubble: len(gr.GetMessages()) > 0,
		Count:      count,
	}
	server.Cache.Set(key, response)
	responses.JSON(w, http.StatusOK, response)
}

// UpdateNotification godoc
// @Summary Update a notifications
// @Description Update a notifications with the input paylod
// @Tags Notification
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Plugin id"
// @Param body body doc.Notification true "Update notification"
// @Success 200 {object} doc.Notification
// @Router /notification/{id} [put]
func (server *Server) UpdateNotification(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data := notifications.Notification{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if data.Id != int64(pid) {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("url id and payload id not matched"))
		return
	}
	c := pb.NewNotificationClient(server.NotifyClient)
	mes := &pb.Message{}
	data.ToGrpc(mes)
	gr, err := c.Update(context.Background(), &pb.UpdateRequest{Message: mes})
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	updated := &notifications.Notification{}
	updated.FromGrpc(gr.GetMessage())
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, updated)
		return
	}
	notificationResponse, err := CreateNotificationResponse(updated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    notificationResponse,
		Success: 1,
		Message: "Success",
	})
}

// MarkAllAsRead godoc
// @Summary Mark All as read for notifications
// @Description Mark all notifications as read
// @Tags Notification
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} doc.SuccessResponse
// @Router /notification/mark-all-as-read [get]
func (server *Server) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	uid, _, _ := auth.ExtractTokenID(r)
	c := pb.NewNotificationClient(server.NotifyClient)
	_, err := c.MarkAll(context.Background(), &pb.MarkAllRequest{User: int64(uid), Type: 1})
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, doc.SuccessResponse{Message: "Notification marked as read"})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// UpdateMultipleNotification godoc
// @Summary Update multiple a notifications
// @Description Update multiple notifications with the input paylod
// @Tags Notification
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.MarkAsSeeNotificationRequest true "Update notifications"
// @Success 200 {object} doc.SuccessResponse
// @Router /notifications/seen-unseen [post]
func (server *Server) UpdateMultipleNotification(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	input := doc.MarkAsSeeNotificationRequest{}
	err = json.Unmarshal(body, &input)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	c := pb.NewNotificationClient(server.NotifyClient)
	for _, v := range input.Ids {
		mes := &pb.Message{}
		data := notifications.Notification{Id: int64(v), Seen: input.Seen}
		data.ToGrpc(mes)
		_, _ = c.Update(context.Background(), &pb.UpdateRequest{Message: mes})
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, doc.SuccessResponse{Message: "Notification updated successfully"})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data := notifications.Notification{}
	c := pb.NewNotificationClient(server.NotifyClient)
	mes := &pb.Message{}
	data.ToGrpc(mes)
	_, err = c.Delete(context.Background(), &pb.DeleteRequest{Id: int64(pid)})
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
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

func CreateNotificationResponse(notification *notifications.Notification) (*doc.Notification, error) {
	notificationResponse := doc.Notification{}
	notificationBytes, _ := json.Marshal(notification)
	err := json.Unmarshal(notificationBytes, &notificationResponse)
	if err != nil {
		return nil, err
	}
	return &notificationResponse, nil
}
