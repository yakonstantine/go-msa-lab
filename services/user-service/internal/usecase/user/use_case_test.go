package user

import (
	"context"
	"errors"
	"testing"

	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/entity"
	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/usecase/mock"
)

func TestCreate_InvalidUserProfile(t *testing.T) {
	// Arrange
	up := &entity.UserProfile{
		CorpKey:        "123abc",
		CountryCode:    "KZ",
		FirstName:      "John",
		LastName:       "Doe",
		FullName:       "Doe, J.P. (John)",
		DepartmentCode: "FOO-987",
	}

	ctx := context.Background()

	urMock := &mock.UserRepository{}
	srMock := &mock.SMTPRepository{}

	sut := NewUseCase(&mock.TransactionFactory{}, urMock, srMock)

	// Act
	gotUser, gotErr := sut.Create(ctx, up)

	// Assert
	var valErr *entity.ValidationError
	if !errors.As(gotErr, &valErr) {
		t.Fatalf("Create = %v, want ValidationError", gotErr)
	}
	if gotUser != nil {
		t.Errorf("Create = %v, want nil", gotUser)
	}
}

func TestCreate_UserAlreadyExists(t *testing.T) {
	// Arrange
	up := &entity.UserProfile{
		CorpKey:        "ABC123",
		CountryCode:    "KZ",
		FirstName:      "John",
		LastName:       "Doe",
		FullName:       "Doe, J.P. (John)",
		DepartmentCode: "FOO-987",
	}

	ctx := context.Background()

	urMock := &mock.UserRepository{
		GetByCorpKeyFn: func(ctx context.Context, ck entity.CorpKey) (*entity.User, error) {
			return &entity.User{
				CorpKey: ck,
			}, nil
		},
	}
	srMock := &mock.SMTPRepository{}

	sut := NewUseCase(&mock.TransactionFactory{}, urMock, srMock)

	// Act
	gotUser, gotErr := sut.Create(ctx, up)

	// Assert
	if !errors.Is(gotErr, entity.ErrAlreadyExists) {
		t.Fatalf("Create = %v, want ErrAlreadyExists", gotErr)
	}
	if gotUser != nil {
		t.Errorf("Create = %v, want nil", gotUser)
	}
}
