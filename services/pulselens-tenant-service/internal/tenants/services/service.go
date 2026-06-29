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

	commonauth "github.com/Avv123/pulselens-common/auth"
	pulsetenant "github.com/Avv123/pulselens-common/tenant"
	"github.com/Avv123/pulselens-platform/authz"
	"github.com/Avv123/pulselens-platform/errs"
	"github.com/Avv123/pulselens-platform/idgen"
	"github.com/Avv123/pulselens-platform/logging"
	"github.com/Avv123/pulselens-tenant-service/constants"
	"github.com/Avv123/pulselens-tenant-service/internal/tenants/models"
	"github.com/Avv123/pulselens-tenant-service/internal/tenants/repositories"
	tenantrequests "github.com/Avv123/pulselens-tenant-service/internal/tenants/requests"
	tenantresponses "github.com/Avv123/pulselens-tenant-service/internal/tenants/responses"
	serviceauth "github.com/Avv123/pulselens-tenant-service/pkg/auth"
	tenanterror "github.com/Avv123/pulselens-tenant-service/pkg/error"
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
		Action:       constants.ActionTenantCreated,
		ResourceType: constants.ResourceTenant,
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

func (s *Service) ListAPIKeysForClaims(ctx context.Context, tenantID string) ([]models.APIKey, errs.CustomError) {
	return s.ListAPIKeys(ctx, tenantID)
}

func (s *Service) ResolveAPIKey(ctx context.Context, rawKey string) (pulsetenant.ResolvedAPIKey, errs.CustomError) {
	keyHash := hashAPIKey(rawKey)

	// B1: read-through Redis cache — avoids 3 sequential DB calls on every ingest request.
	if cached, ok := getResolvedFromCache(ctx, keyHash); ok {
		return cached, errs.CustomError{}
	}

	keyModel, err := s.repository.GetAPIKeyByHash(ctx, keyHash)
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

	// B13: fire last_used_at update in background — never blocks the hot path.
	touchAPIKeyAsync(nil, s.repository, keyModel.ID)

	scopes := make([]pulsetenant.APIKeyScope, 0)
	_ = json.Unmarshal([]byte(keyModel.Scopes), &scopes)

	resolved := pulsetenant.ResolvedAPIKey{
		KeyID:       keyModel.ID,
		TenantID:    tenantModel.ID,
		TenantName:  tenantModel.Name,
		Plan:        string(tenantModel.Plan),
		IngestQuota: tenantModel.IngestQuota,
		ServiceID:   serviceModel.ID,
		ServiceName: serviceModel.Name,
		Environment: serviceModel.Environment,
		Scopes:      scopes,
		Active:      keyModel.Active,
	}

	// Populate cache for subsequent requests within the TTL window.
	setResolvedInCache(ctx, keyHash, resolved)
	return resolved, errs.CustomError{}
}

func (s *Service) RotateAPIKey(ctx context.Context, actorUserID string, tenantID string, keyID string, name string) (tenantresponses.CreateAPIKeyResponse, errs.CustomError) {
	keyModel, err := s.repository.GetAPIKeyByID(ctx, keyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tenantresponses.CreateAPIKeyResponse{}, errs.New(tenanterror.NotFound, "api key not found")
		}
		return tenantresponses.CreateAPIKeyResponse{}, mapDBError(err)
	}
	if keyModel.TenantID != tenantID {
		return tenantresponses.CreateAPIKeyResponse{}, errs.New(tenanterror.Forbidden, "tenant access denied")
	}
	if !keyModel.Active {
		return tenantresponses.CreateAPIKeyResponse{}, errs.New(tenanterror.BadRequest, "api key is already inactive")
	}

	rawKey, err := generateAPIKey()
	if err != nil {
		return tenantresponses.CreateAPIKeyResponse{}, errs.New(tenanterror.InternalServer, err.Error())
	}

	replacement := models.APIKey{
		ID:         idgen.New("key"),
		TenantID:   keyModel.TenantID,
		ServiceID:  keyModel.ServiceID,
		Name:       firstNonEmpty(strings.TrimSpace(name), keyModel.Name+"-rotated"),
		KeyPrefix:  rawKey[:16],
		KeyHash:    hashAPIKey(rawKey),
		Scopes:     keyModel.Scopes,
		Active:     true,
		RotatedAt:  nil,
		RevokedAt:  nil,
		ReplacedBy: "",
	}
	scopes := make([]pulsetenant.APIKeyScope, 0)
	_ = json.Unmarshal([]byte(keyModel.Scopes), &scopes)
	auditLog := &models.AuditLog{
		ID:           idgen.New("audit"),
		TenantID:     keyModel.TenantID,
		ActorUserID:  actorUserID,
		ActorType:    actorType(actorUserID),
		Action:       "api_key.rotated",
		ResourceType: "api_key",
		ResourceID:   keyModel.ID,
		Payload:      marshalJSON(genericPayload{"replaced_by": replacement.ID, "service_id": keyModel.ServiceID}),
	}
	if err = s.repository.RotateAPIKey(ctx, &keyModel, &replacement, auditLog); err != nil {
		return tenantresponses.CreateAPIKeyResponse{}, mapDBError(err)
	}
	// B1: invalidate the old key's cache entry immediately on rotation.
	invalidateResolvedCache(ctx, keyModel.KeyHash)
	return tenantresponses.CreateAPIKeyResponse{
		ID:        replacement.ID,
		TenantID:  replacement.TenantID,
		ServiceID: replacement.ServiceID,
		Name:      replacement.Name,
		Key:       rawKey,
		KeyPrefix: replacement.KeyPrefix,
		Scopes:    scopes,
	}, errs.CustomError{}
}

func (s *Service) RevokeAPIKey(ctx context.Context, actorUserID string, tenantID string, keyID string) errs.CustomError {
	keyModel, err := s.repository.GetAPIKeyByID(ctx, keyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.New(tenanterror.NotFound, "api key not found")
		}
		return mapDBError(err)
	}
	if keyModel.TenantID != tenantID {
		return errs.New(tenanterror.Forbidden, "tenant access denied")
	}
	if !keyModel.Active {
		return errs.New(tenanterror.BadRequest, "api key is already inactive")
	}
	auditLog := &models.AuditLog{
		ID:           idgen.New("audit"),
		TenantID:     keyModel.TenantID,
		ActorUserID:  actorUserID,
		ActorType:    actorType(actorUserID),
		Action:       "api_key.revoked",
		ResourceType: "api_key",
		ResourceID:   keyModel.ID,
		Payload:      marshalJSON(genericPayload{"service_id": keyModel.ServiceID}),
	}
	if err = s.repository.RevokeAPIKey(ctx, keyModel.ID, auditLog); err != nil {
		return mapDBError(err)
	}
	// B1: explicitly invalidate the cache so the revoked key is never served again.
	invalidateResolvedCache(ctx, keyModel.KeyHash)
	return errs.CustomError{}
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

// B7: explicit actor type constants — distinguishes system bootstrap, api_key
// operations, and user-driven actions for forensic audit log analysis.
const (
	ActorTypeUser    = "user"
	ActorTypeInternal = "internal"
)

func actorType(actorUserID string) string {
	if strings.TrimSpace(actorUserID) == "" {
		return ActorTypeInternal
	}
	return ActorTypeUser
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func mapDBError(err error) errs.CustomError {
	// B11: log internal error detail but return an opaque code to callers
	// so raw SQL/constraint names are never exposed in API responses.
	logging.Errorf("tenant-service db error: %v", err)
	lowered := strings.ToLower(err.Error())
	if strings.Contains(lowered, "duplicate") || strings.Contains(lowered, "unique") {
		return errs.New(tenanterror.Conflict, "resource already exists")
	}
	return errs.New(tenanterror.InternalServer, "an internal error occurred")
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
