package user

import (
	"context"
	"fmt"
	"time"

	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/entity"
	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/usecase"
)

type UseCase struct {
	txf      usecase.TransactionFactory
	userRepo usecase.UserRepository
	smtpRepo usecase.SMTPRepository
}

func New(
	txf usecase.TransactionFactory,
	userRepo usecase.UserRepository,
	smtpRepo usecase.SMTPRepository,
) *UseCase {
	return &UseCase{
		txf:      txf,
		userRepo: userRepo,
		smtpRepo: smtpRepo,
	}
}

func (uc *UseCase) Create(ctx context.Context, up *entity.UserProfile) (*entity.User, error) {
	up.Sanitize()
	err := up.Validate()
	if err != nil {
		return nil, err
	}
	if exist, err := uc.userRepo.GetByCorpKey(ctx, up.CorpKey); exist != nil || err != nil {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("user with corp key '%v': %w", up.CorpKey, entity.ErrAlreadyExists)
	}

	primarySMTP, err := generatePrimarySMTP(ctx, uc.smtpRepo, up)
	if err != nil {
		return nil, err
	}

	u := &entity.User{
		CorpKey:        up.CorpKey,
		FullName:       up.FullName,
		CountryCode:    up.CountryCode,
		DepartmentCode: up.DepartmentCode,
		PrimarySMTP:    primarySMTP,
		Deleted:        false,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	err = uc.createTx(ctx, u)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (uc *UseCase) GetByCorpKey(ctx context.Context, ck entity.CorpKey) (*entity.User, error) {
	err := ck.Validate()
	if err != nil {
		return nil, err
	}

	u, err := uc.userRepo.GetByCorpKey(ctx, ck)
	if err != nil {
		return nil, fmt.Errorf("error get user '%v': %w", ck, err)
	}

	smtps, err := uc.smtpRepo.GetByIdentity(ctx, string(ck))
	if err != nil {
		return nil, fmt.Errorf("error get user's '%v' SMTPs: %w", ck, err)
	}
	for _, smtp := range smtps {
		if smtp.Type != entity.Secondary {
			continue
		}
		u.SecondarySMTPs = append(u.SecondarySMTPs, smtp.Address)
	}

	return u, nil
}

func (uc *UseCase) createTx(ctx context.Context, u *entity.User) error {
	tx, err := uc.txf.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = uc.userRepo.Create(ctx, tx, u)
	if err != nil {
		return fmt.Errorf("can't create user '%v': %w", u.CorpKey, err)
	}

	err = uc.smtpRepo.Create(ctx, tx, &entity.SMTPAddress{
		Address:  u.PrimarySMTP,
		Identity: string(u.CorpKey),
		Type:     entity.Primary,
	})
	if err != nil {
		return fmt.Errorf("can't create user's '%v' primary smtp '%v': %w", u.CorpKey, u.PrimarySMTP, err)
	}

	return tx.Commit()
}
