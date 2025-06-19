package user

import (
	"context"

	"github.com/ariesmaulana/payroll/app/user/lib"
	"github.com/ariesmaulana/payroll/common"
	"github.com/ariesmaulana/payroll/internal/jwtutil"
	"github.com/ariesmaulana/payroll/lib/database"
	log "github.com/ariesmaulana/payroll/lib/logger"
)

var _ lib.ServiceInterface = (*Service)(nil)

type Service struct {
	storage lib.StorageInterface
}

func NewService(storage lib.StorageInterface) *Service {
	return &Service{
		storage: storage,
	}
}

func (s *Service) Login(ctx context.Context, in *lib.LoginIn) *lib.LoginOut {
	resp := lib.LoginOut{}

	if in.UserName == "" {
		log.Warn(in.Trace).Msg("username is empty")
		resp.Message = "username tidak boleh kosong"
		return &resp
	}

	if in.Password == "" {
		log.Warn(in.Trace).Msg("password is empty")
		resp.Message = "password tidak boleh kosong"
		return &resp
	}

	tx, err := s.storage.BeginTxReader(ctx)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("Failed begin tx")
		return &resp
	}
	defer tx.Rollback(ctx)

	user, errType, err := s.storage.GetUserByUsername(ctx, in.UserName)
	if err != nil && errType != database.ErrNotFound {
		log.Error(in.Trace).Err(err).Str("type", string(errType)).Msg("failed get user")
		return &resp
	}

	// Check is password is valid
	valid := common.VerifyDjangoPBKDF2Password(in.Password, user.Password)

	if !valid {
		log.Error(in.Trace).Msg("password invalid")
		resp.Message = "Password tidak valid"
		return &resp
	}

	token, err := jwtutil.GenerateJWT(user.Id, user.Username, user.Role)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("failed generate jwt")
		resp.Message = "internal error"
		return &resp
	}

	resp.Success = true
	resp.Token = token
	return &resp
}

func (s *Service) UserSalary(ctx context.Context, in *lib.UserSalaryIn) *lib.UserSalaryOut {
	resp := lib.UserSalaryOut{}

	tx, err := s.storage.BeginTxReader(ctx)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("failed begin tx")
		return &resp
	}
	defer tx.Rollback(ctx)

	salaries, errType, err := s.storage.GetAllUserBaseSalary(ctx)
	if err != nil {
		log.Error(in.Trace).Err(err).Str("type", string(errType)).Msg("failed get user base salary")
		if errType == database.ErrNotFound {
			resp.Message = "Tidak ada data user ditemukan"
		}
		return &resp
	}

	resp.Success = true
	resp.Result = salaries
	return &resp
}
