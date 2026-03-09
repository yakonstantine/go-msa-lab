package user

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/entity"
	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/usecase"
)

type UseCase struct {
	txf      usecase.TransactionFactory
	userRepo usecase.UserRepository
	smtpRepo usecase.SMTPRepository
}

func NewUseCase(
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
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
		Deleted:        false,
	}

	err = uc.createTx(ctx, u)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (uc *UseCase) Update(ctx context.Context, up *entity.UserProfile) (*entity.User, error) {
	up.Sanitize()
	err := up.Validate()
	if err != nil {
		return nil, err
	}

	u, err := uc.getUserWithSMTPs(ctx, up.CorpKey)
	if err != nil {
		return nil, err
	}

	u.FullName = up.FullName
	u.CountryCode = up.CountryCode
	u.DepartmentCode = up.DepartmentCode
	u.UpdatedAt = time.Now().UTC()
	u.Deleted = false

	overrideSMTPs, err := uc.regeneratePrimarySMTP(ctx, up, u)
	if err != nil {
		return nil, err
	}

	err = uc.updateTx(ctx, u, overrideSMTPs)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (uc *UseCase) GetByCorpKey(ctx context.Context, ck entity.CorpKey) (*entity.User, error) {
	return uc.getUserWithSMTPs(ctx, ck)
}

func (uc *UseCase) GetPage(ctx context.Context, limit, offset int) (entity.Page[entity.User], error) {
	var zero entity.Page[entity.User]
	page, err := uc.userRepo.GetPage(ctx, limit, offset)
	if err != nil {
		return entity.Page[entity.User]{}, err
	}
	if len(page.Items) == 0 {
		return zero, nil
	}

	identities := make([]string, 0, len(page.Items))
	for _, item := range page.Items {
		identities = append(identities, string(item.CorpKey))
	}

	addresses, err := uc.smtpRepo.GetByIdentities(ctx, identities)
	if err != nil {
		return zero, fmt.Errorf("error getting users SMTPs: %w", err)
	}

	for _, item := range page.Items {
		smtps, ok := addresses[string(item.CorpKey)]
		if !ok {
			return zero, fmt.Errorf("no SMTPs for user '%v'", item.CorpKey)
		}
		populateUserSMTPs(&item, smtps)
	}

	return page, nil
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

func (uc *UseCase) updateTx(ctx context.Context, u *entity.User, overrideSMTPs bool) error {
	tx, err := uc.txf.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = uc.userRepo.Update(ctx, tx, u)
	if err != nil {
		return fmt.Errorf("can't update user '%v': %w", u.CorpKey, err)
	}

	if !overrideSMTPs {
		return nil
	}

	smtps := make([]entity.SMTPAddress, 0, len(u.SecondarySMTPs)+1)
	smtps = append(smtps, entity.SMTPAddress{
		Address:  u.PrimarySMTP,
		Identity: string(u.CorpKey),
		Type:     entity.Primary,
	})
	for _, smtp := range u.SecondarySMTPs {
		smtps = append(smtps, entity.SMTPAddress{
			Address:  smtp,
			Identity: string(u.CorpKey),
			Type:     entity.Secondary,
		})
	}

	err = uc.smtpRepo.DeleteAll(ctx, tx, string(u.CorpKey))
	if err != nil {
		return fmt.Errorf("can't remove currents smtps '%v': %w", u.CorpKey, err)
	}
	err = uc.smtpRepo.CreateMany(ctx, tx, smtps)
	if err != nil {
		return fmt.Errorf("can't replace smtps '%v': %w", u.CorpKey, err)
	}

	return nil
}

func (uc *UseCase) getUserWithSMTPs(ctx context.Context, ck entity.CorpKey) (*entity.User, error) {
	u, err := uc.userRepo.GetByCorpKey(ctx, ck)
	if err != nil {
		return nil, fmt.Errorf("error get user '%v': %w", ck, err)
	}
	if u == nil {
		return nil, fmt.Errorf("user with corp key '%v' %w", ck, entity.ErrNotFound)
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

func populateUserSMTPs(u *entity.User, smtps []entity.SMTPAddress) {
	for _, smtp := range smtps {
		if smtp.Type != entity.Secondary {
			continue
		}
		u.SecondarySMTPs = append(u.SecondarySMTPs, smtp.Address)
	}
}

func (uc *UseCase) regeneratePrimarySMTP(ctx context.Context, up *entity.UserProfile, u *entity.User) (bool, error) {
	regenerated := false

	newPrimarySMTP, err := generatePrimarySMTP(ctx, uc.smtpRepo, up)
	if err != nil {
		return regenerated, err
	}
	if u.PrimarySMTP == newPrimarySMTP {
		return regenerated, nil
	}

	regenerated = true

	u.SecondarySMTPs = append(u.SecondarySMTPs, u.PrimarySMTP)
	u.PrimarySMTP = newPrimarySMTP

	for i, sec := range u.SecondarySMTPs {
		if sec == newPrimarySMTP {
			u.SecondarySMTPs = slices.Delete(u.SecondarySMTPs, i, i+1)
			break
		}
	}

	return regenerated, nil
}
