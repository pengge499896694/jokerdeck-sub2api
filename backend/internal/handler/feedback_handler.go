package handler

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

type FeedbackHandler struct { service *service.FeedbackService }
func NewFeedbackHandler(s *service.FeedbackService) *FeedbackHandler { return &FeedbackHandler{service:s} }
type feedbackCreateRequest struct { Subject string `json:"subject"`; Category string `json:"category"`; Body string `json:"body"`; Model string `json:"model"`; ErrorCode string `json:"error_code"`; RequestID string `json:"request_id"`; Metadata map[string]string `json:"metadata"` }
type FeedbackReplyRequestExport struct { Body string `json:"body"`; Model string `json:"model"`; ErrorCode string `json:"error_code"`; RequestID string `json:"request_id"`; Metadata map[string]string `json:"metadata"` }

func (h *FeedbackHandler) List(c *gin.Context) { subject,ok:=middleware.GetAuthSubjectFromContext(c); if !ok {response.Unauthorized(c,"User not found in context");return}; page,size:=response.ParsePagination(c); items,total,err:=h.service.ListTickets(c,subject.UserID,false,c.Query("status"),page,size);if err!=nil{response.Error(c,500,err.Error());return};response.Paginated(c,items,total,page,size) }
func (h *FeedbackHandler) Create(c *gin.Context) { subject,ok:=middleware.GetAuthSubjectFromContext(c);if !ok{response.Unauthorized(c,"User not found in context");return};var req feedbackCreateRequest;if err:=c.ShouldBindJSON(&req);err!=nil{response.BadRequest(c,"Invalid request: "+err.Error());return};t,err:=h.service.CreateTicket(c,subject.UserID,req.Subject,req.Category,req.Body,service.FeedbackContext{Model:req.Model,ErrorCode:req.ErrorCode,RequestID:req.RequestID,Metadata:req.Metadata});if err!=nil{response.BadRequest(c,err.Error());return};response.Success(c,t) }
func (h *FeedbackHandler) Get(c *gin.Context) { subject,ok:=middleware.GetAuthSubjectFromContext(c);if !ok{response.Unauthorized(c,"User not found in context");return};id,err:=strconv.ParseInt(c.Param("id"),10,64);if err!=nil||id<=0{response.BadRequest(c,"Invalid ticket ID");return};t,err:=h.service.GetTicket(c,id,subject.UserID,false);if err!=nil{response.Error(c,404,err.Error());return};response.Success(c,t) }
func (h *FeedbackHandler) Reply(c *gin.Context) { subject,ok:=middleware.GetAuthSubjectFromContext(c);if !ok{response.Unauthorized(c,"User not found in context");return};id,err:=strconv.ParseInt(c.Param("id"),10,64);if err!=nil||id<=0{response.BadRequest(c,"Invalid ticket ID");return};var req FeedbackReplyRequestExport;if err=c.ShouldBindJSON(&req);err!=nil{response.BadRequest(c,"Invalid request: "+err.Error());return};err=h.service.Reply(c,id,subject.UserID,false,req.Body,service.FeedbackContext{Model:req.Model,ErrorCode:req.ErrorCode,RequestID:req.RequestID,Metadata:req.Metadata});if err!=nil{response.Error(c,400,err.Error());return};t,err:=h.service.GetTicket(c,id,subject.UserID,false);if err!=nil{response.Error(c,404,err.Error());return};response.Success(c,t) }
func (h *FeedbackHandler) Close(c *gin.Context) { subject,ok:=middleware.GetAuthSubjectFromContext(c);if !ok{response.Unauthorized(c,"User not found in context");return};id,err:=strconv.ParseInt(c.Param("id"),10,64);if err!=nil||id<=0{response.BadRequest(c,"Invalid ticket ID");return};if _,err=h.service.GetTicket(c,id,subject.UserID,false);err!=nil{response.Error(c,404,err.Error());return};if err=h.service.SetStatus(c,id,"closed");err!=nil{response.Error(c,400,err.Error());return};response.Success(c,gin.H{"message":"ok"}) }
