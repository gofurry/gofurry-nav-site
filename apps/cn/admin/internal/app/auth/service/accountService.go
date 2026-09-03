package service

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/GoFurry/easyhash"
	"github.com/gofurry/gofurry-admin/internal/app/auth/authorization"
	"github.com/gofurry/gofurry-admin/internal/app/auth/models"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	adminsqlc "github.com/gofurry/gofurry-admin/internal/db/admin/sqlc"
	"github.com/gofurry/gofurry-admin/pkg/common"
	"github.com/jackc/pgx/v5"
)

func (s *AuthService) ChangeOwnUsername(ctx context.Context, accountID int64, value, currentPassword string, meta audit.Meta) (*authorization.Principal, common.Error) {
	username, validationErr := validateUsername(value)
	if validationErr != nil {
		return nil, validationErr
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, daoError(err)
	}
	defer tx.Rollback(ctx)
	queries := s.q.WithTx(tx)
	before, appErr := lockAccount(ctx, queries, accountID)
	if appErr != nil {
		return nil, appErr
	}
	if passwordErr := verifyCurrentPassword(before, currentPassword); passwordErr != nil {
		return nil, passwordErr
	}
	if before.Username == username {
		if commitErr := commitAccountTx(ctx, tx); commitErr != nil {
			return nil, commitErr
		}
		return principalForAccount(before), nil
	}
	row, err := queries.UpdateAdminAccountUsername(ctx, adminsqlc.UpdateAdminAccountUsernameParams{Username: username, AccountID: accountID})
	if err != nil {
		return nil, accountWriteError(err)
	}
	after, convertErr := accountFromValues(row.ID, row.Username, row.DisplayName, row.Role, row.Status,
		row.PasswordHash, row.SessionVersion, row.LastLoginAt, row.CreatedAt, row.UpdatedAt, row.PasswordUpdatedAt)
	if convertErr != nil {
		return nil, convertErr
	}
	if auditErr := s.audit.LogTx(ctx, tx, meta, "account.username_changed", "gfa_admin_account", accountID, accountAuditSnapshot(before), accountAuditSnapshot(after)); auditErr != nil {
		return nil, auditErr
	}
	if commitErr := commitAccountTx(ctx, tx); commitErr != nil {
		return nil, commitErr
	}
	return principalForAccount(after), nil
}

func (s *AuthService) ChangeOwnPassword(ctx context.Context, accountID int64, currentPassword, newPassword string, meta audit.Meta) common.Error {
	if validationErr := validatePassword(newPassword); validationErr != nil {
		return validationErr
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return daoError(err)
	}
	defer tx.Rollback(ctx)
	queries := s.q.WithTx(tx)
	before, appErr := lockAccount(ctx, queries, accountID)
	if appErr != nil {
		return appErr
	}
	if passwordErr := verifyCurrentPassword(before, currentPassword); passwordErr != nil {
		return passwordErr
	}
	hash, hashErr := s.createPasswordHash(newPassword)
	if hashErr != nil {
		return hashErr
	}
	row, err := queries.UpdateAdminAccountPassword(ctx, adminsqlc.UpdateAdminAccountPasswordParams{PasswordHash: hash, PasswordUpdatedAt: timestamp(time.Now()), AccountID: accountID})
	if err != nil {
		return accountWriteError(err)
	}
	after, convertErr := accountFromValues(row.ID, row.Username, row.DisplayName, row.Role, row.Status,
		row.PasswordHash, row.SessionVersion, row.LastLoginAt, row.CreatedAt, row.UpdatedAt, row.PasswordUpdatedAt)
	if convertErr != nil {
		return convertErr
	}
	if auditErr := s.audit.LogTx(ctx, tx, meta, "account.password_changed", "gfa_admin_account", accountID, accountAuditSnapshot(before), accountAuditSnapshot(after)); auditErr != nil {
		return auditErr
	}
	return commitAccountTx(ctx, tx)
}

func verifyCurrentPassword(account *models.AdminAccount, password string) common.Error {
	if account == nil || password == "" {
		return invalidCredentials()
	}
	ok, err := easyhash.VerifyPBKDF2(password, account.PasswordHash)
	if err != nil {
		return common.NewServiceError("password verification failed")
	}
	if !ok {
		return invalidCredentials()
	}
	return nil
}

func (s *AuthService) ListAccounts(ctx context.Context, keyword string, limit, offset int32) (models.AccountPage, common.Error) {
	keyword = strings.TrimSpace(keyword)
	total, err := s.q.CountAdminAccountsFiltered(ctx, keyword)
	if err != nil {
		return models.AccountPage{}, daoError(err)
	}
	rows, err := s.q.ListAdminAccounts(ctx, adminsqlc.ListAdminAccountsParams{Keyword: keyword, RowLimit: limit, RowOffset: offset})
	if err != nil {
		return models.AccountPage{}, daoError(err)
	}
	page := models.AccountPage{Total: total, List: make([]models.AccountResponse, 0, len(rows))}
	for _, row := range rows {
		account, convertErr := accountFromValues(row.ID, row.Username, row.DisplayName, row.Role, row.Status,
			row.PasswordHash, row.SessionVersion, row.LastLoginAt, row.CreatedAt, row.UpdatedAt, row.PasswordUpdatedAt)
		if convertErr != nil {
			return models.AccountPage{}, convertErr
		}
		page.List = append(page.List, models.AccountDTO(*account))
	}
	return page, nil
}

func (s *AuthService) CreateAccount(ctx context.Context, request models.AccountCreateRequest, meta audit.Meta) (models.AccountResponse, common.Error) {
	username, validationErr := validateUsername(request.Username)
	if validationErr != nil {
		return models.AccountResponse{}, validationErr
	}
	displayName, validationErr := validateDisplayName(request.DisplayName)
	if validationErr != nil {
		return models.AccountResponse{}, validationErr
	}
	role, ok := authorization.ParseRole(string(request.Role))
	if !ok {
		return models.AccountResponse{}, common.NewValidationError("role must be owner, developer, or operator")
	}
	if validationErr = validatePassword(request.Password); validationErr != nil {
		return models.AccountResponse{}, validationErr
	}
	hash, hashErr := s.createPasswordHash(request.Password)
	if hashErr != nil {
		return models.AccountResponse{}, hashErr
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.AccountResponse{}, daoError(err)
	}
	defer tx.Rollback(ctx)
	queries := s.q.WithTx(tx)
	now := time.Now()
	row, err := queries.InsertAdminAccount(ctx, adminsqlc.InsertAdminAccountParams{
		Username: username, DisplayName: displayName, Role: string(role), Status: string(authorization.StatusActive),
		PasswordHash: hash, SessionVersion: 1, PasswordUpdatedAt: timestamp(now),
	})
	if err != nil {
		return models.AccountResponse{}, accountWriteError(err)
	}
	account, convertErr := accountFromValues(row.ID, row.Username, row.DisplayName, row.Role, row.Status,
		row.PasswordHash, row.SessionVersion, row.LastLoginAt, row.CreatedAt, row.UpdatedAt, row.PasswordUpdatedAt)
	if convertErr != nil {
		return models.AccountResponse{}, convertErr
	}
	if auditErr := s.audit.LogTx(ctx, tx, meta, "account.created", "gfa_admin_account", account.ID, nil, accountAuditSnapshot(account)); auditErr != nil {
		return models.AccountResponse{}, auditErr
	}
	if err := tx.Commit(ctx); err != nil {
		return models.AccountResponse{}, daoError(err)
	}
	return models.AccountDTO(*account), nil
}

func (s *AuthService) UpdateDisplayName(ctx context.Context, accountID int64, value string, meta audit.Meta) (models.AccountResponse, common.Error) {
	displayName, validationErr := validateDisplayName(value)
	if validationErr != nil {
		return models.AccountResponse{}, validationErr
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.AccountResponse{}, daoError(err)
	}
	defer tx.Rollback(ctx)
	queries := s.q.WithTx(tx)
	before, appErr := lockAccount(ctx, queries, accountID)
	if appErr != nil {
		return models.AccountResponse{}, appErr
	}
	if before.DisplayName == displayName {
		return models.AccountDTO(*before), commitAccountTx(ctx, tx)
	}
	row, err := queries.UpdateAdminAccountDisplayName(ctx, adminsqlc.UpdateAdminAccountDisplayNameParams{DisplayName: displayName, AccountID: accountID})
	if err != nil {
		return models.AccountResponse{}, accountWriteError(err)
	}
	after, convertErr := accountFromValues(row.ID, row.Username, row.DisplayName, row.Role, row.Status,
		row.PasswordHash, row.SessionVersion, row.LastLoginAt, row.CreatedAt, row.UpdatedAt, row.PasswordUpdatedAt)
	if convertErr != nil {
		return models.AccountResponse{}, convertErr
	}
	if auditErr := s.audit.LogTx(ctx, tx, meta, "account.display_name_changed", "gfa_admin_account", accountID, accountAuditSnapshot(before), accountAuditSnapshot(after)); auditErr != nil {
		return models.AccountResponse{}, auditErr
	}
	if commitErr := commitAccountTx(ctx, tx); commitErr != nil {
		return models.AccountResponse{}, commitErr
	}
	return models.AccountDTO(*after), nil
}

func (s *AuthService) ChangeRole(ctx context.Context, accountID int64, value authorization.Role, meta audit.Meta) (models.AccountResponse, common.Error) {
	role, ok := authorization.ParseRole(string(value))
	if !ok {
		return models.AccountResponse{}, common.NewValidationError("role must be owner, developer, or operator")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.AccountResponse{}, daoError(err)
	}
	defer tx.Rollback(ctx)
	queries := s.q.WithTx(tx)
	ownerIDs, err := queries.LockActiveOwnerIDs(ctx)
	if err != nil {
		return models.AccountResponse{}, daoError(err)
	}
	before, appErr := lockAccount(ctx, queries, accountID)
	if appErr != nil {
		return models.AccountResponse{}, appErr
	}
	if before.Role == role {
		return models.AccountDTO(*before), commitAccountTx(ctx, tx)
	}
	if before.Role == authorization.RoleOwner && before.Status == authorization.StatusActive && role != authorization.RoleOwner && len(ownerIDs) <= 1 {
		return models.AccountResponse{}, lastOwnerError()
	}
	row, err := queries.UpdateAdminAccountRole(ctx, adminsqlc.UpdateAdminAccountRoleParams{Role: string(role), AccountID: accountID})
	if err != nil {
		return models.AccountResponse{}, accountWriteError(err)
	}
	after, convertErr := accountFromValues(row.ID, row.Username, row.DisplayName, row.Role, row.Status,
		row.PasswordHash, row.SessionVersion, row.LastLoginAt, row.CreatedAt, row.UpdatedAt, row.PasswordUpdatedAt)
	if convertErr != nil {
		return models.AccountResponse{}, convertErr
	}
	if auditErr := s.audit.LogTx(ctx, tx, meta, "account.role_changed", "gfa_admin_account", accountID, accountAuditSnapshot(before), accountAuditSnapshot(after)); auditErr != nil {
		return models.AccountResponse{}, auditErr
	}
	if commitErr := commitAccountTx(ctx, tx); commitErr != nil {
		return models.AccountResponse{}, commitErr
	}
	return models.AccountDTO(*after), nil
}

func (s *AuthService) ChangeStatus(ctx context.Context, accountID int64, value authorization.AccountStatus, meta audit.Meta) (models.AccountResponse, common.Error) {
	status, ok := authorization.ParseStatus(string(value))
	if !ok {
		return models.AccountResponse{}, common.NewValidationError("status must be active or disabled")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.AccountResponse{}, daoError(err)
	}
	defer tx.Rollback(ctx)
	queries := s.q.WithTx(tx)
	ownerIDs, err := queries.LockActiveOwnerIDs(ctx)
	if err != nil {
		return models.AccountResponse{}, daoError(err)
	}
	before, appErr := lockAccount(ctx, queries, accountID)
	if appErr != nil {
		return models.AccountResponse{}, appErr
	}
	if before.Status == status {
		return models.AccountDTO(*before), commitAccountTx(ctx, tx)
	}
	if before.Role == authorization.RoleOwner && before.Status == authorization.StatusActive && status == authorization.StatusDisabled && len(ownerIDs) <= 1 {
		return models.AccountResponse{}, lastOwnerError()
	}
	row, err := queries.UpdateAdminAccountStatus(ctx, adminsqlc.UpdateAdminAccountStatusParams{Status: string(status), AccountID: accountID})
	if err != nil {
		return models.AccountResponse{}, accountWriteError(err)
	}
	after, convertErr := accountFromValues(row.ID, row.Username, row.DisplayName, row.Role, row.Status,
		row.PasswordHash, row.SessionVersion, row.LastLoginAt, row.CreatedAt, row.UpdatedAt, row.PasswordUpdatedAt)
	if convertErr != nil {
		return models.AccountResponse{}, convertErr
	}
	action := "account.enabled"
	if status == authorization.StatusDisabled {
		action = "account.disabled"
	}
	if auditErr := s.audit.LogTx(ctx, tx, meta, action, "gfa_admin_account", accountID, accountAuditSnapshot(before), accountAuditSnapshot(after)); auditErr != nil {
		return models.AccountResponse{}, auditErr
	}
	if commitErr := commitAccountTx(ctx, tx); commitErr != nil {
		return models.AccountResponse{}, commitErr
	}
	return models.AccountDTO(*after), nil
}

func (s *AuthService) ResetAccountPassword(ctx context.Context, accountID int64, password string, meta audit.Meta) (models.AccountResponse, common.Error) {
	if validationErr := validatePassword(password); validationErr != nil {
		return models.AccountResponse{}, validationErr
	}
	hash, hashErr := s.createPasswordHash(password)
	if hashErr != nil {
		return models.AccountResponse{}, hashErr
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.AccountResponse{}, daoError(err)
	}
	defer tx.Rollback(ctx)
	queries := s.q.WithTx(tx)
	before, appErr := lockAccount(ctx, queries, accountID)
	if appErr != nil {
		return models.AccountResponse{}, appErr
	}
	row, err := queries.UpdateAdminAccountPassword(ctx, adminsqlc.UpdateAdminAccountPasswordParams{PasswordHash: hash, PasswordUpdatedAt: timestamp(time.Now()), AccountID: accountID})
	if err != nil {
		return models.AccountResponse{}, accountWriteError(err)
	}
	after, convertErr := accountFromValues(row.ID, row.Username, row.DisplayName, row.Role, row.Status,
		row.PasswordHash, row.SessionVersion, row.LastLoginAt, row.CreatedAt, row.UpdatedAt, row.PasswordUpdatedAt)
	if convertErr != nil {
		return models.AccountResponse{}, convertErr
	}
	if auditErr := s.audit.LogTx(ctx, tx, meta, "account.password_reset", "gfa_admin_account", accountID, accountAuditSnapshot(before), accountAuditSnapshot(after)); auditErr != nil {
		return models.AccountResponse{}, auditErr
	}
	if commitErr := commitAccountTx(ctx, tx); commitErr != nil {
		return models.AccountResponse{}, commitErr
	}
	return models.AccountDTO(*after), nil
}

func (s *AuthService) RevokeSessions(ctx context.Context, accountID int64, meta audit.Meta) (models.AccountResponse, common.Error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.AccountResponse{}, daoError(err)
	}
	defer tx.Rollback(ctx)
	queries := s.q.WithTx(tx)
	before, appErr := lockAccount(ctx, queries, accountID)
	if appErr != nil {
		return models.AccountResponse{}, appErr
	}
	row, err := queries.IncrementAdminSessionVersion(ctx, accountID)
	if err != nil {
		return models.AccountResponse{}, accountWriteError(err)
	}
	after, convertErr := accountFromValues(row.ID, row.Username, row.DisplayName, row.Role, row.Status,
		row.PasswordHash, row.SessionVersion, row.LastLoginAt, row.CreatedAt, row.UpdatedAt, row.PasswordUpdatedAt)
	if convertErr != nil {
		return models.AccountResponse{}, convertErr
	}
	if auditErr := s.audit.LogTx(ctx, tx, meta, "account.sessions_revoked", "gfa_admin_account", accountID, accountAuditSnapshot(before), accountAuditSnapshot(after)); auditErr != nil {
		return models.AccountResponse{}, auditErr
	}
	if commitErr := commitAccountTx(ctx, tx); commitErr != nil {
		return models.AccountResponse{}, commitErr
	}
	return models.AccountDTO(*after), nil
}

func lockAccount(ctx context.Context, queries *adminsqlc.Queries, accountID int64) (*models.AdminAccount, common.Error) {
	row, err := queries.LockAdminAccountByID(ctx, accountID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, accountNotFound()
	}
	if err != nil {
		return nil, daoError(err)
	}
	return accountFromValues(row.ID, row.Username, row.DisplayName, row.Role, row.Status,
		row.PasswordHash, row.SessionVersion, row.LastLoginAt, row.CreatedAt, row.UpdatedAt, row.PasswordUpdatedAt)
}

func commitAccountTx(ctx context.Context, tx pgx.Tx) common.Error {
	if err := tx.Commit(ctx); err != nil {
		return daoError(err)
	}
	return nil
}

func lastOwnerError() common.Error {
	return common.NewError(common.RETURN_FAILED, http.StatusConflict, "the last active owner cannot be disabled or demoted")
}
