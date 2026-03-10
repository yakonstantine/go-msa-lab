package dto

import (
	"time"

	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/entity"
)

type UserDTO struct {
	CorpKey        string    `json:"corpKey"`
	FullName       string    `json:"fullName"`
	CountryCode    string    `json:"countryCode"`
	DepartmentCode string    `json:"departmentCode"`
	PrimarySMTP    string    `json:"primarySMTP"`
	SecondarySMTPs []string  `json:"secondarySMTPs"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func ToUserDTO(u entity.User) UserDTO {
	secondaries := make([]string, 0, len(u.SecondarySMTPs))
	for _, sec := range u.SecondarySMTPs {
		secondaries = append(secondaries, string(sec))
	}

	return UserDTO{
		CorpKey:        string(u.CorpKey),
		FullName:       u.FullName,
		CountryCode:    string(u.CountryCode),
		DepartmentCode: u.DepartmentCode,
		PrimarySMTP:    string(u.PrimarySMTP),
		SecondarySMTPs: secondaries,
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
	}
}
