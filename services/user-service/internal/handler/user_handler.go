package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/dto"
	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/entity"
)

type ChangeUserRequest struct {
	CorpKey        string `json:"corpKey"`
	FirstName      string `json:"firstName" binding:"required"`
	LastName       string `json:"lastName" binding:"required"`
	FullName       string `json:"fullName" binding:"required"`
	CountryCode    string `json:"countryCode" binding:"required"`
	DepartmentCode string `json:"departmentCode" binding:"required"`
}

func toUserProfile(r ChangeUserRequest) *entity.UserProfile {
	return &entity.UserProfile{
		CorpKey:        entity.CorpKey(r.CorpKey),
		FirstName:      entity.Name(r.FirstName),
		LastName:       entity.Name(r.LastName),
		FullName:       r.FullName,
		CountryCode:    entity.CountryCode(r.CountryCode),
		DepartmentCode: r.DepartmentCode,
	}
}

type UserUseCase interface {
	Create(context.Context, *entity.UserProfile) (*entity.User, error)
	Update(context.Context, *entity.UserProfile) (*entity.User, error)
	GetByCorpKey(context.Context, entity.CorpKey) (*entity.User, error)
	GetPage(ctx context.Context, limit, offset int) (entity.Page[entity.User], error)
}

type UserHandler struct {
	useCase UserUseCase
}

func NewUserHandler(uc UserUseCase) *UserHandler {
	return &UserHandler{
		useCase: uc,
	}
}

func (h *UserHandler) Create(c *gin.Context) {
	var req ChangeUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	up := toUserProfile(req)
	u, err := h.useCase.Create(c.Request.Context(), up)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToUserDTO(*u))
}

func (h *UserHandler) Update(c *gin.Context) {
	ck := c.Param("corpKey")
	var req ChangeUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	up := toUserProfile(req)
	up.CorpKey = entity.NewCorpKey(ck)
	u, err := h.useCase.Update(c.Request.Context(), up)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dto.ToUserDTO(*u))
}

func (h *UserHandler) GetByCorpKey(c *gin.Context) {
	ck := c.Param("corpKey")
	u, err := h.useCase.GetByCorpKey(c.Request.Context(), entity.NewCorpKey(ck))
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dto.ToUserDTO(*u))
}
