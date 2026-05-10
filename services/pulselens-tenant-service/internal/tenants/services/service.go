package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	commonauth "github.com/omniful/pulselens-common/auth"
	pulsetenant "github.com/omniful/pulselens-common/tenant"
	"github.com/omniful/pulselens-platform/authz"
	"github.com/omniful/pulselens-platform/errs"
	"github.com/omniful/pulselens-platform/idgen"
	"github.com/omniful/pulselens-platform/logging"
	"github.com/omniful/pulselens-tenant-service/internal/tenants/models"
	"github.com/omniful/pulselens-tenant-service/internal/tenants/repositories"
	tenantrequests "github.com/omniful/pulselens-tenant-service/internal/tenants/requests"
	tenantresponses "github.com/omniful/pulselens-tenant-service/internal/tenants/responses"
	serviceauth "github.com/omniful/pulselens-tenant-service/pkg/auth"
	tenanterror "github.com/omniful/pulselens-tenant-service/pkg/error"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service struct {
	repository    *repositories.Repository
	jwtSecret     string
	jwtExpiryMins int
}

func New(repository *repositories.Repository, jwtSecret string, jwtExpiryMins int) *Service {
	return &Service{
		repository:    repository,
		jwtSecret:     jwtSecret,
		jwtExpiryMins: jwtExpiryMins,
	}
}

func (s *Service) CreateTenant(ctx context.Context, request *tenantrequests.CreateTenantRequest) (models.Tenant, models.User, errs.CustomError) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return models.Tenant{}, models.User{}, errs.New(tenanterror.InternalServer, err.Error())
	}

	tenantModel := models.Tenant{
		ID:            idgen.New("tenant"),
		Name:          request.Name,
		Slug:          strings.ToLower(request.Slug),
		Plan:          request.Plan,
		IngestQuota:   request.IngestQuota,
		RetentionDays: request.RetentionDays,
	}

	userModel := models.User{
		ID:           idgen.New("user"),
		TenantID:     tenantModel.ID,
		Name:         request.AdminName,
		Email:        strings.ToLower(request.AdminEmail),
		PasswordHash: string(passwordHash),
		Role:         authz.RoleTenantAdmin,
	}

	auditLog := &models.AuditLog{
		ID:           idgen.New("audit"),
		TenantID:     tenantModel.ID,
		ActorUserID:  userModel.ID,
		ActorType:    "internal",
		Action:       "tenant.created",
		ResourceType: "tenant",
		ResourceID:   tenantModel.ID,
		Payload:      marshalJSON(genericPayload{"slug": tenantModel.Slug, "plan": tenantModel.Plan}),
	}

	if err = s.repository.CreateTenant(ctx, &tenantModel, &userModel, auditLog); err != nil {
		return models.Tenant{}, models.User{}, mapDBError(err)
	}

	return tenantModel, userModel, errs.CustomError{}
}

func (s *Service) ListTenants(ctx context.Context) ([]models.Tenant, errs.CustomError) {
	rows, err := s.repository.ListTenants(ctx)
	if err != nil {
		return nil, mapDBError(err)
	}
	return rows, errs.CustomError{}
}

func (s *Service) GetTenant(ctx context.Context, tenantID string) (models.Tenant, errs.CustomError) {
	row, err := s.repository.GetTenant(ctx, tenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Tenant{}, errs.New(tenanterror.NotFound, "tenant not found")
		}
		return models.Tenant{}, mapDBError(err)
	}
	return row, errs.CustomError{}
}

func (s *Service) CreateService(ctx context.Context, tenantID string, actorUserID string, request *tenantrequests.CreateServiceRequest) (models.Service, errs.CustomError) {
	row := models.Service{
		ID:          idgen.New("svc"),
		TenantID:    tenantID,
		Name:        request.Name,
		Environment: request.Environment,
		Tags:        marshalJSON(request.Tags),
	}

	auditLog := &models.AuditLog{
		ID:           idgen.New("audit"),
		TenantID:     tenantID,
		ActorUserID:  actorUserID,
		ActorType:    actorType(actorUserID),
		Action:       "service.created",
		ResourceType: "service",
		ResourceID:   row.ID,
		Payload:      marshalJSON(genericPayload{"name": row.Name, "environment": row.Environment}),
	}

	if err := s.repository.CreateService(ctx, &row, auditLog); err != nil {
		return models.Service{}, mapDBError(err)
	}

	return row, errs.CustomError{}
}

func (s *Service) ListServices(ctx context.Context, tenantID string) ([]models.Service, errs.CustomError) {
	rows, err := s.repository.ListServices(ctx, tenantID)
	if err != nil {
		return nil, mapDBError(err)
	}
	return rows, errs.CustomError{}
}

func (s *Service) CreateAPIKey(ctx context.Context, actorUserID string, request *tenantrequests.CreateAPIKeyRequest) (tenantresponses.CreateAPIKeyResponse, errs.CustomError) {
	serviceModel, err := s.repository.GetService(ctx, request.ServiceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tenantresponses.CreateAPIKeyResponse{}, errs.New(tenanterror.NotFound, "service not found")
		}
		return tenantresponses.CreateAPIKeyResponse{}, mapDBError(err)
	}

	if serviceModel.TenantID != request.TenantID {
		return tenantresponses.CreateAPIKeyResponse{}, errs.New(tenanterror.Forbidden, "service does not belong to tenant")
	}

	rawKey, err := generateAPIKey()
	if err != nil {
		return tenantresponses.CreateAPIKeyResponse{}, errs.New(tenanterror.InternalServer, err.Error())
	}

	row := models.APIKey{
		ID:        idgen.New("key"),
		TenantID:  request.TenantID,
		ServiceID: request.ServiceID,
		Name:      request.Name,
		KeyPrefix: rawKey[:16],
		KeyHash:   hashAPIKey(rawKey),
		Scopes:    marshalJSON(request.Scopes),
		Active:    true,
	}

	auditLog := &models.AuditLog{
		ID:           idgen.New("audit"),
		TenantID:     request.TenantID,
		ActorUserID:  actorUserID,
		ActorType:    actorType(actorUserID),
		Action:       "api_key.created",
		ResourceType: "api_key",
		ResourceID:   row.ID,
		Payload:      marshalJSON(genericPayload{"name": row.Name, "service_id": row.ServiceID}),
	}

	if err = s.repository.CreateAPIKey(ctx, &row, auditLog); err != nil {
		return tenantresponses.CreateAPIKeyResponse{}, mapDBError(err)
	}

	return tenantresponses.CreateAPIKeyResponse{
		ID:        row.ID,
		TenantID:  row.TenantID,
		ServiceID: row.ServiceID,
		Name:      row.Name,
		Key:       rawKey,
		KeyPrefix: row.KeyPrefix,
		Scopes:    request.Scopes,
	}, errs.CustomError{}
}

func (s *Service) ListAPIKeys(ctx context.Context, tenantID string) ([]models.APIKey, errs.CustomError) {
	rows, err := s.repository.ListAPIKeys(ctx, tenantID)
	if err != nil {
		return nil, mapDBError(err)
	}
	return rows, errs.CustomError{}
}

func (s *Service) ResolveAPIKey(ctx context.Context, rawKey string) (pulsetenant.ResolvedAPIKey, errs.CustomError) {
	keyModel, err := s.repository.GetAPIKeyByHash(ctx, hashAPIKey(rawKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return pulsetenant.ResolvedAPIKey{}, errs.New(tenanterror.Unauthorized, "invalid api key")
		}
		return pulsetenant.ResolvedAPIKey{}, mapDBError(err)
	}

	tenantModel, err := s.repository.GetTenant(ctx, keyModel.TenantID)
	if err != nil {
		return pulsetenant.ResolvedAPIKey{}, mapDBError(err)
	}

	serviceModel, err := s.repository.GetService(ctx, keyModel.ServiceID)
	if err != nil {
		return pulsetenant.ResolvedAPIKey{}, mapDBError(err)
	}

	_ = s.repository.TouchAPIKey(ctx, keyModel.ID)

	scopes := make([]pulsetenant.APIKeyScope, 0)
	_ = json.Unmarshal([]byte(keyModel.Scopes), &scopes)

	return pulsetenant.ResolvedAPIKey{
		KeyID:       keyModel.ID,
		TenantID:    tenantModel.ID,
		TenantName:  tenantModel.Name,
		Plan:        tenantModel.Plan,
		IngestQuota: tenantModel.IngestQuota,
		ServiceID:   serviceModel.ID,
		ServiceName: serviceModel.Name,
		Environment: serviceModel.Environment,
		Scopes:      scopes,
		Active:      keyModel.Active,
	}, errs.CustomError{}
}

func (s *Service) Login(ctx context.Context, request *tenantrequests.LoginRequest) (tenantresponses.LoginResponse, errs.CustomError) {
	userModel, err := s.repository.GetUserByEmail(ctx, strings.ToLower(request.Email))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tenantresponses.LoginResponse{}, errs.New(tenanterror.Unauthorized, "invalid credentials")
		}
		return tenantresponses.LoginResponse{}, mapDBError(err)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(userModel.PasswordHash), []byte(request.Password)); err != nil {
		return tenantresponses.LoginResponse{}, errs.New(tenanterror.Unauthorized, "invalid credentials")
	}

	claims := serviceauth.NewClaims(userModel.ID, userModel.TenantID, userModel.Email, userModel.Role, s.jwtExpiryMins)
	token, err := serviceauth.GenerateToken(s.jwtSecret, claims)
	if err != nil {
		return tenantresponses.LoginResponse{}, errs.New(tenanterror.InternalServer, err.Error())
	}

	return tenantresponses.LoginResponse{
		Token:    token,
		UserID:   userModel.ID,
		TenantID: userModel.TenantID,
		Email:    userModel.Email,
		Role:     userModel.Role,
	}, errs.CustomError{}
}

func (s *Service) Me(ctx context.Context, claims *commonauth.Claims) (models.User, errs.CustomError) {
	userModel, err := s.repository.GetUserByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, errs.New(tenanterror.NotFound, "user not found")
		}
		return models.User{}, mapDBError(err)
	}
	return userModel, errs.CustomError{}
}

func (s *Service) ListAuditLogs(ctx context.Context, tenantID string) ([]models.AuditLog, errs.CustomError) {
	rows, err := s.repository.ListAuditLogs(ctx, tenantID)
	if err != nil {
		return nil, mapDBError(err)
	}
	return rows, errs.CustomError{}
}

func (s *Service) CreateUser(ctx context.Context, tenantID string, actorUserID string, request *tenantrequests.CreateUserRequest) (models.User, errs.CustomError) {
	if strings.TrimSpace(request.Name) == "" || strings.TrimSpace(request.Email) == "" || strings.TrimSpace(request.Password) == "" {
		return models.User{}, errs.New(tenanterror.BadRequest, "name, email, and password are required")
	}
	if !validRole(request.Role) {
		return models.User{}, errs.New(tenanterror.BadRequest, "invalid role")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, errs.New(tenanterror.InternalServer, err.Error())
	}

	row := models.User{
		ID:           idgen.New("user"),
		TenantID:     tenantID,
		Name:         strings.TrimSpace(request.Name),
		Email:        strings.ToLower(strings.TrimSpace(request.Email)),
		PasswordHash: string(passwordHash),
		Role:         strings.TrimSpace(request.Role),
	}
	auditLog := &models.AuditLog{
		ID:           idgen.New("audit"),
		TenantID:     tenantID,
		ActorUserID:  actorUserID,
		ActorType:    actorType(actorUserID),
		Action:       "user.created",
		ResourceType: "user",
		ResourceID:   row.ID,
		Payload:      marshalJSON(genericPayload{"email": row.Email, "role": row.Role}),
	}
	if err = s.repository.CreateUser(ctx, &row, auditLog); err != nil {
		return models.User{}, mapDBError(err)
	}
	return row, errs.CustomError{}
}

func (s *Service) ListUsers(ctx context.Context, tenantID string) ([]models.User, errs.CustomError) {
	rows, err := s.repository.ListUsers(ctx, tenantID)
	if err != nil {
		return nil, mapDBError(err)
	}
	return rows, errs.CustomError{}
}

type genericPayload map[string]interface{}

func marshalJSON(payload interface{}) string {
	bytes, _ := json.Marshal(payload)
	return string(bytes)
}

func actorType(actorUserID string) string {
	if actorUserID == "" {
		return "internal"
	}
	return "user"
}

func mapDBError(err error) errs.CustomError {
	logging.Errorf("tenant-service db error: %v", err)
	lowered := strings.ToLower(err.Error())
	if strings.Contains(lowered, "duplicate") || strings.Contains(lowered, "unique") {
		return errs.New(tenanterror.Conflict, err.Error())
	}
	return errs.New(tenanterror.InternalServer, err.Error())
}

func generateAPIKey() (string, error) {
	randomBytes := make([]byte, 24)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("pls_%s", hex.EncodeToString(randomBytes)), nil
}

func hashAPIKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func validRole(role string) bool {
	switch strings.TrimSpace(role) {
	case authz.RoleTenantAdmin, authz.RoleViewer, authz.RoleOperator, authz.RoleAlertManager, authz.RoleServiceOwner:
		return true
	default:
		return false
	}
}
