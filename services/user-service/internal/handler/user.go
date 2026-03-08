package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/dto"
	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/entity"
	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/usecase/user"
)

type ChangeUserRequest struct {
	CorpKey        string `json:"corpKey"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	FullName       string `json:"fullName"`
	CountryCode    string `json:"countryCode"`
	DepartmentCode string `json:"departmentCode"`
}

func (r *ChangeUserRequest) toUserProfile() *entity.UserProfile {
	return &entity.UserProfile{
		CorpKey:        entity.CorpKey(r.CorpKey),
		FirstName:      entity.Name(r.FirstName),
		LastName:       entity.Name(r.LastName),
		FullName:       r.FullName,
		CountryCode:    entity.CountryCode(r.CountryCode),
		DepartmentCode: r.DepartmentCode,
	}
}

type UserHandler struct {
	useCase *user.UseCase
}

func NewUserHandler(uc *user.UseCase) *UserHandler {
	return &UserHandler{
		useCase: uc,
	}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	req := &ChangeUserRequest{}
	if err := c.BindJSON(req); err != nil {
		return
	}

	up := req.toUserProfile()
	u, err := h.useCase.Create(c.Request.Context(), up)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, dto.UserFromEntity(u))
}

func (h *UserHandler) GetByCorpKey(c *gin.Context) {
	ck := c.Param("corpKey")
	if ck == "" {
		c.Error(fmt.Errorf("no corpKey path parameter provided"))
		return
	}

	u, err := h.useCase.GetByCorpKey(c.Request.Context(), entity.NewCorpKey(ck))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dto.UserFromEntity(u))
}
