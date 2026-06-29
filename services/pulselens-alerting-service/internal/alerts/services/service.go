package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	alertmodels "github.com/Avv123/pulselens-alerting-service/internal/alerts/models"
	alertrepositories "github.com/Avv123/pulselens-alerting-service/internal/alerts/repositories"
	alertrequests "github.com/Avv123/pulselens-alerting-service/internal/alerts/requests"
	alertresponses "github.com/Avv123/pulselens-alerting-service/internal/alerts/responses"
	"github.com/Avv123/pulselens-alerting-service/pkg/postgres"
	commonauth "github.com/Avv123/pulselens-common/auth"
	"github.com/Avv123/pulselens-platform/errs"
	"github.com/Avv123/pulselens-platform/idgen"
	"github.com/Avv123/pulselens-platform/logging"
	"gorm.io/gorm"
)

type Service struct {
	repository *alertrepositories.Repository
}

func New() *Service {
	return &Service{repository: alertrepositories.NewRepository(postgres.Get())}
}

func (s *Service) CreateRule(ctx context.Context, claims *commonauth.Claims, request *alertrequests.CreateRuleRequest) (alertmodels.AlertRule, errs.CustomError) {
	if customError := validateRuleRequest(request.Name, request.SignalType, request.Comparator, request.WindowMinutes); customError.Exists() {
		return alertmodels.AlertRule{}, customError
	}

	policyID, customError := s.resolvePolicyID(ctx, claims.TenantID, request.PolicyID)
	if customError.Exists() {
		return alertmodels.AlertRule{}, customError
	}

	active := true
	if request.Active != nil {
		active = *request.Active
	}

	row := alertmodels.AlertRule{
		ID:              idgen.New("rule"),
		TenantID:        claims.TenantID,
		ServiceID:       request.ServiceID,
		PolicyID:        policyID,
		Name:            request.Name,
		Description:     request.Description,
		SignalType:      strings.ToLower(request.SignalType),
		MetricName:      request.MetricName,
		Severity:        strings.ToLower(request.Severity),
		Aggregation:     normalizeAggregation(request.Aggregation),
		Comparator:      request.Comparator,
		Threshold:       request.Threshold,
		WindowMinutes:   request.WindowMinutes,
		CooldownMinutes: maxInt(request.CooldownMinutes, 1),
		Active:          active,
	}

	if err := s.repository.CreateRule(ctx, &row); err != nil {
		return alertmodels.AlertRule{}, mapDBError(err)
	}
	return row, errs.CustomError{}
}

func (s *Service) UpdateRule(ctx context.Context, claims *commonauth.Claims, ruleID string, request *alertrequests.UpdateRuleRequest) (alertmodels.AlertRule, errs.CustomError) {
	row, err := s.repository.GetRule(ctx, claims.TenantID, ruleID)
	if err != nil {
		return alertmodels.AlertRule{}, mapDBError(err)
	}
	if customError := validateRuleRequest(request.Name, request.SignalType, request.Comparator, request.WindowMinutes); customError.Exists() {
		return alertmodels.AlertRule{}, customError
	}
	policyID, customError := s.resolvePolicyID(ctx, claims.TenantID, request.PolicyID)
	if customError.Exists() {
		return alertmodels.AlertRule{}, customError
	}

	row.PolicyID = policyID
	row.Name = request.Name
	row.Description = request.Description
	row.SignalType = strings.ToLower(request.SignalType)
	row.MetricName = request.MetricName
	row.Severity = strings.ToLower(request.Severity)
	row.Aggregation = normalizeAggregation(request.Aggregation)
	row.Comparator = request.Comparator
	row.Threshold = request.Threshold
	row.WindowMinutes = request.WindowMinutes
	row.CooldownMinutes = maxInt(request.CooldownMinutes, 1)
	if request.Active != nil {
		row.Active = *request.Active
	}

	if err = s.repository.UpdateRule(ctx, &row); err != nil {
		return alertmodels.AlertRule{}, mapDBError(err)
	}
	return row, errs.CustomError{}
}

func (s *Service) CreatePolicy(ctx context.Context, claims *commonauth.Claims, request *alertrequests.CreateAlertPolicyRequest) (alertmodels.AlertPolicy, errs.CustomError) {
	row, customError := buildPolicyRow(claims.TenantID, idgen.New("policy"), request.Name, request.Description, request.MaxDeliveryAttempts, request.DeliveryBackoffMillis, request.EscalationIntervalMinutes, request.MaxEscalations, request.RepeatNotificationMinutes, request.OpenChannelTypes, request.AckChannelTypes, request.ResolveChannelTypes, request.EscalationChannelTypes, request.Active)
	if customError.Exists() {
		return alertmodels.AlertPolicy{}, customError
	}
	if err := s.repository.CreatePolicy(ctx, &row); err != nil {
		return alertmodels.AlertPolicy{}, mapDBError(err)
	}
	return row, errs.CustomError{}
}

func (s *Service) UpdatePolicy(ctx context.Context, claims *commonauth.Claims, policyID string, request *alertrequests.UpdateAlertPolicyRequest) (alertmodels.AlertPolicy, errs.CustomError) {
	row, err := s.repository.GetPolicy(ctx, claims.TenantID, policyID)
	if err != nil {
		return alertmodels.AlertPolicy{}, mapDBError(err)
	}
	updated, customError := buildPolicyRow(claims.TenantID, row.ID, request.Name, request.Description, request.MaxDeliveryAttempts, request.DeliveryBackoffMillis, request.EscalationIntervalMinutes, request.MaxEscalations, request.RepeatNotificationMinutes, request.OpenChannelTypes, request.AckChannelTypes, request.ResolveChannelTypes, request.EscalationChannelTypes, request.Active)
	if customError.Exists() {
		return alertmodels.AlertPolicy{}, customError
	}
	row.Name = updated.Name
	row.Description = updated.Description
	row.MaxDeliveryAttempts = updated.MaxDeliveryAttempts
	row.DeliveryBackoffMillis = updated.DeliveryBackoffMillis
	row.EscalationIntervalMinutes = updated.EscalationIntervalMinutes
	row.MaxEscalations = updated.MaxEscalations
	row.RepeatNotificationMinutes = updated.RepeatNotificationMinutes
	row.OpenChannelTypes = updated.OpenChannelTypes
	row.AckChannelTypes = updated.AckChannelTypes
	row.ResolveChannelTypes = updated.ResolveChannelTypes
	row.EscalationChannelTypes = updated.EscalationChannelTypes
	row.Active = updated.Active
	if err = s.repository.UpdatePolicy(ctx, &row); err != nil {
		return alertmodels.AlertPolicy{}, mapDBError(err)
	}
	return row, errs.CustomError{}
}

func (s *Service) ListPolicies(ctx context.Context, claims *commonauth.Claims) ([]alertmodels.AlertPolicy, errs.CustomError) {
	rows, err := s.repository.ListPolicies(ctx, claims.TenantID)
	if err != nil {
		return nil, mapDBError(err)
	}
	return rows, errs.CustomError{}
}

func (s *Service) ListRules(ctx context.Context, claims *commonauth.Claims) ([]alertmodels.AlertRule, errs.CustomError) {
	rows, err := s.repository.ListRules(ctx, claims.TenantID)
	if err != nil {
		return nil, mapDBError(err)
	}
	return rows, errs.CustomError{}
}

func (s *Service) ListIncidents(ctx context.Context, claims *commonauth.Claims, filters alertrepositories.IncidentFilters) ([]alertmodels.Incident, errs.CustomError) {
	rows, err := s.repository.ListIncidents(ctx, claims.TenantID, filters)
	if err != nil {
		return nil, mapDBError(err)
	}
	return rows, errs.CustomError{}
}

func (s *Service) GetIncident(ctx context.Context, claims *commonauth.Claims, incidentID string) (alertmodels.Incident, errs.CustomError) {
	row, err := s.repository.GetIncident(ctx, claims.TenantID, incidentID)
	if err != nil {
		return alertmodels.Incident{}, mapDBError(err)
	}
	return row, errs.CustomError{}
}

func (s *Service) ListIncidentTimeline(ctx context.Context, claims *commonauth.Claims, incidentID string) ([]alertmodels.IncidentEvent, errs.CustomError) {
	rows, err := s.repository.ListIncidentEvents(ctx, claims.TenantID, incidentID)
	if err != nil {
		return nil, mapDBError(err)
	}
	return rows, errs.CustomError{}
}

func (s *Service) ListIncidentDeliveries(ctx context.Context, claims *commonauth.Claims, incidentID string) ([]alertmodels.NotificationDelivery, errs.CustomError) {
	rows, err := s.repository.ListIncidentDeliveries(ctx, claims.TenantID, incidentID)
	if err != nil {
		return nil, mapDBError(err)
	}
	return rows, errs.CustomError{}
}

func (s *Service) AcknowledgeIncident(ctx context.Context, claims *commonauth.Claims, incidentID string) (alertmodels.Incident, errs.CustomError) {
	row, err := s.repository.GetIncident(ctx, claims.TenantID, incidentID)
	if err != nil {
		return alertmodels.Incident{}, mapDBError(err)
	}
	now := time.Now().UTC()
	row.Status = "acknowledged"
	row.AcknowledgedBy = claims.UserID
	row.AcknowledgedAt = &now
	row.NextEscalationAt = nil
	if err = s.repository.Transaction(ctx, func(repo *alertrepositories.Repository) error {
		if saveErr := repo.SaveIncident(ctx, &row); saveErr != nil {
			return saveErr
		}
		return repo.CreateIncidentEvent(ctx, buildIncidentEvent(row, "incident.acknowledged", claims.UserID, "Incident acknowledged", map[string]any{
			"status":          row.Status,
			"acknowledged_by": claims.UserID,
		}))
	}); err != nil {
		return alertmodels.Incident{}, mapDBError(err)
	}
	_ = s.notifyIncident(ctx, row, "incident.acknowledged")
	return row, errs.CustomError{}
}

func (s *Service) ResolveIncident(ctx context.Context, claims *commonauth.Claims, incidentID string) (alertmodels.Incident, errs.CustomError) {
	row, err := s.repository.GetIncident(ctx, claims.TenantID, incidentID)
	if err != nil {
		return alertmodels.Incident{}, mapDBError(err)
	}
	now := time.Now().UTC()
	row.Status = "resolved"
	row.ResolvedAt = &now
	row.NextEscalationAt = nil
	if err = s.repository.Transaction(ctx, func(repo *alertrepositories.Repository) error {
		if saveErr := repo.SaveIncident(ctx, &row); saveErr != nil {
			return saveErr
		}
		return repo.CreateIncidentEvent(ctx, buildIncidentEvent(row, "incident.resolved", claims.UserID, "Incident resolved", map[string]any{
			"status": row.Status,
		}))
	}); err != nil {
		return alertmodels.Incident{}, mapDBError(err)
	}
	_ = s.notifyIncident(ctx, row, "incident.resolved")
	return row, errs.CustomError{}
}

func (s *Service) AssignIncident(ctx context.Context, claims *commonauth.Claims, incidentID string, request *alertrequests.AssignIncidentRequest) (alertmodels.Incident, errs.CustomError) {
	if strings.TrimSpace(request.AssignedTo) == "" {
		return alertmodels.Incident{}, errs.New("BAD_REQUEST", "assigned_to is required")
	}
	row, err := s.repository.GetIncident(ctx, claims.TenantID, incidentID)
	if err != nil {
		return alertmodels.Incident{}, mapDBError(err)
	}
	now := time.Now().UTC()
	row.AssignedTo = request.AssignedTo
	row.AssignedAt = &now
	if err = s.repository.Transaction(ctx, func(repo *alertrepositories.Repository) error {
		if saveErr := repo.SaveIncident(ctx, &row); saveErr != nil {
			return saveErr
		}
		return repo.CreateIncidentEvent(ctx, buildIncidentEvent(row, "incident.assigned", claims.UserID, "Incident assigned", map[string]any{
			"assigned_to": row.AssignedTo,
		}))
	}); err != nil {
		return alertmodels.Incident{}, mapDBError(err)
	}
	_ = s.notifyIncident(ctx, row, "incident.assigned")
	return row, errs.CustomError{}
}

func (s *Service) CreateNotificationChannel(ctx context.Context, claims *commonauth.Claims, request *alertrequests.CreateNotificationChannelRequest) (alertmodels.NotificationChannel, errs.CustomError) {
	if strings.TrimSpace(request.Name) == "" || strings.TrimSpace(request.Type) == "" {
		return alertmodels.NotificationChannel{}, errs.New("BAD_REQUEST", "name and type are required")
	}
	configMap, ok := request.Config.(map[string]interface{})
	if !ok {
		return alertmodels.NotificationChannel{}, errs.New("BAD_REQUEST", "config must be an object")
	}
	switch strings.ToLower(strings.TrimSpace(request.Type)) {
	case "webhook", "slack_webhook":
		if strings.TrimSpace(stringValue(configMap["url"])) == "" {
			return alertmodels.NotificationChannel{}, errs.New("BAD_REQUEST", "config.url is required")
		}
	case "email":
		recipients, ok := configMap["to"].([]interface{})
		if !ok || len(recipients) == 0 {
			return alertmodels.NotificationChannel{}, errs.New("BAD_REQUEST", "email config.to must contain at least one recipient")
		}
	}
	configPayload, err := json.Marshal(request.Config)
	if err != nil {
		return alertmodels.NotificationChannel{}, errs.New("BAD_REQUEST", err.Error())
	}
	active := true
	if request.Active != nil {
		active = *request.Active
	}
	row := alertmodels.NotificationChannel{
		ID:       idgen.New("channel"),
		TenantID: claims.TenantID,
		Name:     request.Name,
		Type:     strings.ToLower(request.Type),
		Config:   string(configPayload),
		Active:   active,
	}
	if err = s.repository.CreateNotificationChannel(ctx, &row); err != nil {
		return alertmodels.NotificationChannel{}, mapDBError(err)
	}
	return row, errs.CustomError{}
}

func (s *Service) ListNotificationChannels(ctx context.Context, claims *commonauth.Claims) ([]alertmodels.NotificationChannel, errs.CustomError) {
	rows, err := s.repository.ListNotificationChannels(ctx, claims.TenantID)
	if err != nil {
		return nil, mapDBError(err)
	}
	return rows, errs.CustomError{}
}

func (s *Service) ListNotificationDeliveries(ctx context.Context, claims *commonauth.Claims) ([]alertmodels.NotificationDelivery, errs.CustomError) {
	rows, err := s.repository.ListNotificationDeliveries(ctx, claims.TenantID)
	if err != nil {
		return nil, mapDBError(err)
	}
	return rows, errs.CustomError{}
}

func (s *Service) AddIncidentComment(ctx context.Context, claims *commonauth.Claims, incidentID string, request *alertrequests.CreateIncidentCommentRequest) (alertmodels.IncidentComment, errs.CustomError) {
	if strings.TrimSpace(request.Body) == "" {
		return alertmodels.IncidentComment{}, errs.New("BAD_REQUEST", "body is required")
	}
	if _, err := s.repository.GetIncident(ctx, claims.TenantID, incidentID); err != nil {
		return alertmodels.IncidentComment{}, mapDBError(err)
	}
	row := alertmodels.IncidentComment{
		ID:         idgen.New("comment"),
		IncidentID: incidentID,
		TenantID:   claims.TenantID,
		AuthorID:   claims.UserID,
		Body:       request.Body,
	}
	if err := s.repository.Transaction(ctx, func(repo *alertrepositories.Repository) error {
		if createErr := repo.CreateIncidentComment(ctx, &row); createErr != nil {
			return createErr
		}
		incident, incidentErr := repo.GetIncident(ctx, claims.TenantID, incidentID)
		if incidentErr != nil {
			return incidentErr
		}
		return repo.CreateIncidentEvent(ctx, buildIncidentEvent(incident, "incident.comment_added", claims.UserID, "Incident comment added", map[string]any{
			"comment_id": row.ID,
			"body":       row.Body,
		}))
	}); err != nil {
		return alertmodels.IncidentComment{}, mapDBError(err)
	}
	return row, errs.CustomError{}
}

func (s *Service) ListIncidentComments(ctx context.Context, claims *commonauth.Claims, incidentID string) ([]alertmodels.IncidentComment, errs.CustomError) {
	rows, err := s.repository.ListIncidentComments(ctx, claims.TenantID, incidentID)
	if err != nil {
		return nil, mapDBError(err)
	}
	return rows, errs.CustomError{}
}

func (s *Service) EvaluateRule(ctx context.Context, rule alertmodels.AlertRule) (alertresponses.EvaluationResult, error) {
	observedValue, err := s.observedValue(ctx, rule)
	if err != nil {
		return alertresponses.EvaluationResult{}, err
	}

	now := time.Now().UTC()
	breached := compare(observedValue, rule.Comparator, rule.Threshold)
	openIncident, incidentErr := s.repository.GetOpenIncident(ctx, rule.ID)
	incidentStatus := "none"
	if incidentErr == nil {
		incidentStatus = openIncident.Status
	}

	if breached {
		if shouldTrigger(rule, now) {
			triggeredAt := now
			if incidentErr == nil {
				openIncident.ObservedValue = observedValue
				openIncident.Threshold = rule.Threshold
				openIncident.Summary = buildIncidentSummary(rule, observedValue)
				if err = s.repository.Transaction(ctx, func(repo *alertrepositories.Repository) error {
					if saveErr := repo.SaveIncident(ctx, &openIncident); saveErr != nil {
						return saveErr
					}
					return repo.CreateIncidentEvent(ctx, buildIncidentEvent(openIncident, "incident.retriggered", "", "Incident observed value updated", map[string]any{
						"observed_value": openIncident.ObservedValue,
						"threshold":      openIncident.Threshold,
					}))
				}); err != nil {
					return alertresponses.EvaluationResult{}, err
				}
				incidentStatus = openIncident.Status
			} else if errors.Is(incidentErr, gorm.ErrRecordNotFound) {
				incident := alertmodels.Incident{
					ID:              idgen.New("incident"),
					AlertRuleID:     rule.ID,
					TenantID:        rule.TenantID,
					ServiceID:       rule.ServiceID,
					Severity:        strings.ToLower(rule.Severity),
					Status:          "open",
					EscalationLevel: 0,
					EscalationCount: 0,
					Title:           rule.Name,
					Summary:         buildIncidentSummary(rule, observedValue),
					ObservedValue:   observedValue,
					Threshold:       rule.Threshold,
					TriggeredAt:     now,
				}
				policy := s.policyForRule(ctx, rule)
				nextEscalationAt := now.Add(time.Duration(maxInt(policy.EscalationIntervalMinutes, 1)) * time.Minute)
				incident.NextEscalationAt = &nextEscalationAt
				if err = s.repository.Transaction(ctx, func(repo *alertrepositories.Repository) error {
					if createErr := repo.CreateIncident(ctx, &incident); createErr != nil {
						return createErr
					}
					return repo.CreateIncidentEvent(ctx, buildIncidentEvent(incident, "incident.opened", "", "Incident opened", map[string]any{
						"observed_value": incident.ObservedValue,
						"threshold":      incident.Threshold,
						"severity":       incident.Severity,
					}))
				}); err != nil {
					return alertresponses.EvaluationResult{}, err
				}
				_ = s.notifyIncident(ctx, incident, "incident.opened")
				incidentStatus = incident.Status
			} else {
				return alertresponses.EvaluationResult{}, incidentErr
			}

			if err = s.repository.MarkRuleEvaluated(ctx, rule.ID, now, &triggeredAt); err != nil {
				return alertresponses.EvaluationResult{}, err
			}
		} else {
			if err = s.repository.MarkRuleEvaluated(ctx, rule.ID, now, nil); err != nil {
				return alertresponses.EvaluationResult{}, err
			}
		}
	} else {
		if incidentErr == nil {
			openIncident.Status = "resolved"
			openIncident.ResolvedAt = &now
			openIncident.NextEscalationAt = nil
			openIncident.Summary = buildResolutionSummary(rule, observedValue)
			if err = s.repository.Transaction(ctx, func(repo *alertrepositories.Repository) error {
				if saveErr := repo.SaveIncident(ctx, &openIncident); saveErr != nil {
					return saveErr
				}
				return repo.CreateIncidentEvent(ctx, buildIncidentEvent(openIncident, "incident.resolved", "", "Incident auto-resolved", map[string]any{
					"observed_value": observedValue,
				}))
			}); err != nil {
				return alertresponses.EvaluationResult{}, err
			}
			_ = s.notifyIncident(ctx, openIncident, "incident.resolved")
			incidentStatus = openIncident.Status
		}
		if err = s.repository.MarkRuleEvaluated(ctx, rule.ID, now, nil); err != nil {
			return alertresponses.EvaluationResult{}, err
		}
	}

	return alertresponses.EvaluationResult{
		RuleID:         rule.ID,
		Breached:       breached,
		ObservedValue:  observedValue,
		Comparator:     rule.Comparator,
		Threshold:      rule.Threshold,
		IncidentStatus: incidentStatus,
	}, nil
}

// EvaluateAll evaluates every active alert rule concurrently.
// B12: a bounded goroutine pool (semaphore size 10) ensures that a slow rule or
// a slow DB call does not starve the evaluation of every other rule in the tick.
func (s *Service) EvaluateAll(ctx context.Context) error {
	rules, err := s.repository.ListActiveRules(ctx)
	if err != nil {
		return err
	}

	const concurrency = 10
	sem := make(chan struct{}, concurrency)
	done := make(chan struct{})

	go func() {
		defer close(done)
		for _, rule := range rules {
			sem <- struct{}{}
			go func(r alertmodels.AlertRule) {
				defer func() { <-sem }()
				if _, evalErr := s.EvaluateRule(ctx, r); evalErr != nil {
					logging.Errorf("failed to evaluate rule=%s err=%v", r.ID, evalErr)
				}
			}(rule)
		}
		// drain the semaphore to wait for all goroutines
		for i := 0; i < concurrency; i++ {
			sem <- struct{}{}
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case <-done:
		return nil
	}
}

func (s *Service) EvaluateEscalations(ctx context.Context) error {
	rows, err := s.repository.ListEscalationCandidates(ctx, 50)
	if err != nil {
		return err
	}
	for _, incident := range rows {
		now := time.Now().UTC()
		rule, ruleErr := s.repository.GetRule(ctx, incident.TenantID, incident.AlertRuleID)
		if ruleErr != nil {
			logging.Errorf("failed to load rule for escalation incident=%s err=%v", incident.ID, ruleErr)
			continue
		}
		policy := s.policyForRule(ctx, rule)
		if incident.EscalationLevel >= maxInt(policy.MaxEscalations, 1) {
			incident.NextEscalationAt = nil
			if saveErr := s.repository.SaveIncident(ctx, &incident); saveErr != nil {
				logging.Errorf("failed to clamp escalation incident=%s err=%v", incident.ID, saveErr)
			}
			continue
		}
		incident.EscalationLevel++
		incident.EscalationCount++
		incident.LastEscalatedAt = &now
		nextEscalationAt := now.Add(time.Duration(maxInt(policy.EscalationIntervalMinutes, 1)) * time.Minute)
		incident.NextEscalationAt = &nextEscalationAt
		incident.Summary = fmt.Sprintf("%s | escalated to level %d", incident.Summary, incident.EscalationLevel)
		if saveErr := s.repository.Transaction(ctx, func(repo *alertrepositories.Repository) error {
			if err := repo.SaveIncident(ctx, &incident); err != nil {
				return err
			}
			return repo.CreateIncidentEvent(ctx, buildIncidentEvent(incident, "incident.escalated", "", "Incident escalated", map[string]any{
				"escalation_level": incident.EscalationLevel,
				"escalation_count": incident.EscalationCount,
			}))
		}); saveErr != nil {
			logging.Errorf("failed to save escalation incident=%s err=%v", incident.ID, saveErr)
			continue
		}
		_ = s.notifyIncident(ctx, incident, "incident.escalated")
	}
	return nil
}

func (s *Service) observedValue(ctx context.Context, rule alertmodels.AlertRule) (float64, error) {
	since := time.Now().UTC().Add(-time.Duration(maxInt(rule.WindowMinutes, 1)) * time.Minute)
	switch rule.SignalType {
	case "metric":
		return s.repository.AggregateMetric(ctx, rule.TenantID, rule.ServiceID, rule.MetricName, normalizeAggregation(rule.Aggregation), since)
	case "trace":
		return s.repository.CountTraceErrors(ctx, rule.TenantID, rule.ServiceID, since)
	default:
		return s.repository.CountLogEvents(ctx, rule.TenantID, rule.ServiceID, rule.Severity, since)
	}
}

func validateRuleRequest(name string, signalType string, comparator string, windowMinutes int) errs.CustomError {
	if strings.TrimSpace(name) == "" {
		return errs.New("BAD_REQUEST", "name is required")
	}
	switch strings.ToLower(signalType) {
	case "log", "metric", "trace":
	default:
		return errs.New("BAD_REQUEST", "signal_type must be log, metric, or trace")
	}
	switch comparator {
	case ">", ">=", "<", "<=":
	default:
		return errs.New("BAD_REQUEST", "comparator must be one of >, >=, <, <=")
	}
	if windowMinutes <= 0 {
		return errs.New("BAD_REQUEST", "window_minutes must be greater than zero")
	}
	return errs.CustomError{}
}

func normalizeAggregation(value string) string {
	switch strings.ToLower(value) {
	case "max", "sum", "count":
		return strings.ToLower(value)
	default:
		return "avg"
	}
}

func compare(left float64, comparator string, right float64) bool {
	switch comparator {
	case ">":
		return left > right
	case ">=":
		return left >= right
	case "<":
		return left < right
	case "<=":
		return left <= right
	default:
		return false
	}
}

func shouldTrigger(rule alertmodels.AlertRule, now time.Time) bool {
	if rule.LastTriggeredAt == nil {
		return true
	}
	return rule.LastTriggeredAt.Add(time.Duration(maxInt(rule.CooldownMinutes, 1)) * time.Minute).Before(now)
}

func buildIncidentSummary(rule alertmodels.AlertRule, observedValue float64) string {
	return fmt.Sprintf("%s breached: observed %.2f %s %.2f", rule.Name, observedValue, rule.Comparator, rule.Threshold)
}

func buildResolutionSummary(rule alertmodels.AlertRule, observedValue float64) string {
	return fmt.Sprintf("%s recovered: observed %.2f", rule.Name, observedValue)
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}

func mapDBError(err error) errs.CustomError {
	if err == nil {
		return errs.CustomError{}
	}
	if err == gorm.ErrRecordNotFound {
		return errs.New("NOT_FOUND", "record not found")
	}
	return errs.New("INTERNAL_SERVER_ERROR", err.Error())
}

func (s *Service) notifyIncident(ctx context.Context, incident alertmodels.Incident, eventType string) error {
	rule, err := s.repository.GetRule(ctx, incident.TenantID, incident.AlertRuleID)
	if err != nil {
		return err
	}
	policy := s.policyForRule(ctx, rule)
	channels, err := s.repository.ListActiveNotificationChannels(ctx, incident.TenantID)
	if err != nil {
		return err
	}
	channelTypes := policyChannelTypes(policy, eventType)
	payload := map[string]interface{}{
		"incident_id": incident.ID,
		"title":       incident.Title,
		"status":      incident.Status,
		"event_type":  eventType,
		"summary":     incident.Summary,
	}
	payloadBytes, _ := json.Marshal(payload)
	for _, channel := range channels {
		if len(channelTypes) > 0 && !containsIgnoreCase(channelTypes, channel.Type) {
			continue
		}
		delivery := alertmodels.NotificationDelivery{
			ID:           idgen.New("delivery"),
			TenantID:     incident.TenantID,
			IncidentID:   incident.ID,
			ChannelID:    channel.ID,
			EventType:    eventType,
			Status:       "pending",
			AttemptCount: 0,
			Payload:      string(payloadBytes),
		}
		if err = s.repository.CreateNotificationDelivery(ctx, &delivery); err != nil {
			return err
		}
		_ = s.repository.CreateIncidentEvent(ctx, buildIncidentEvent(incident, "notification.delivery_attempted", "", "Notification delivery attempted", map[string]any{
			"delivery_id": delivery.ID,
			"channel_id":  channel.ID,
			"event_type":  eventType,
		}))
		status, response, deliveredAt := deliverNotificationWithPolicy(ctx, channel, payloadBytes, policy.MaxDeliveryAttempts, policy.DeliveryBackoffMillis)
		delivery.Status = status
		delivery.AttemptCount = maxInt(policy.MaxDeliveryAttempts, 1)
		delivery.Response = response
		delivery.DeliveredAt = deliveredAt
		if err = s.repository.UpdateNotificationDelivery(ctx, &delivery); err != nil {
			return err
		}
		eventName := "notification.delivered"
		eventSummary := "Notification delivered"
		if status != "delivered" {
			eventName = "notification.delivery_failed"
			eventSummary = "Notification delivery failed"
		}
		_ = s.repository.CreateIncidentEvent(ctx, buildIncidentEvent(incident, eventName, "", eventSummary, map[string]any{
			"delivery_id":   delivery.ID,
			"channel_id":    channel.ID,
			"status":        delivery.Status,
			"attempt_count": delivery.AttemptCount,
			"response":      delivery.Response,
			"notification":  eventType,
			"delivered_at":  delivery.DeliveredAt,
		}))
	}
	return nil
}

func (s *Service) resolvePolicyID(ctx context.Context, tenantID string, requestedPolicyID string) (string, errs.CustomError) {
	if strings.TrimSpace(requestedPolicyID) != "" {
		row, err := s.repository.GetPolicy(ctx, tenantID, requestedPolicyID)
		if err != nil {
			return "", mapDBError(err)
		}
		return row.ID, errs.CustomError{}
	}
	row, err := s.repository.FindDefaultPolicy(ctx, tenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			row = defaultPolicyRow(tenantID)
			if createErr := s.repository.CreatePolicy(ctx, &row); createErr != nil {
				return "", mapDBError(createErr)
			}
			return row.ID, errs.CustomError{}
		}
		return "", mapDBError(err)
	}
	return row.ID, errs.CustomError{}
}

func (s *Service) policyForRule(ctx context.Context, rule alertmodels.AlertRule) alertmodels.AlertPolicy {
	if strings.TrimSpace(rule.PolicyID) != "" {
		if row, err := s.repository.GetPolicy(ctx, rule.TenantID, rule.PolicyID); err == nil {
			return row
		}
	}
	return defaultPolicyRow(rule.TenantID)
}

func buildPolicyRow(tenantID string, policyID string, name string, description string, maxDeliveryAttempts int, deliveryBackoffMillis int, escalationIntervalMinutes int, maxEscalations int, repeatNotificationMinutes int, openChannelTypes []string, ackChannelTypes []string, resolveChannelTypes []string, escalationChannelTypes []string, active *bool) (alertmodels.AlertPolicy, errs.CustomError) {
	if strings.TrimSpace(name) == "" {
		return alertmodels.AlertPolicy{}, errs.New("BAD_REQUEST", "policy name is required")
	}
	isActive := true
	if active != nil {
		isActive = *active
	}
	row := defaultPolicyRow(tenantID)
	row.ID = policyID
	row.Name = name
	row.Description = description
	row.MaxDeliveryAttempts = maxInt(maxDeliveryAttempts, 1)
	row.DeliveryBackoffMillis = maxInt(deliveryBackoffMillis, 200)
	row.EscalationIntervalMinutes = maxInt(escalationIntervalMinutes, 1)
	row.MaxEscalations = maxInt(maxEscalations, 1)
	row.RepeatNotificationMinutes = maxInt(repeatNotificationMinutes, 1)
	row.OpenChannelTypes = marshalJSON(normalizeChannelTypes(openChannelTypes, []string{"webhook", "slack_webhook", "email"}))
	row.AckChannelTypes = marshalJSON(normalizeChannelTypes(ackChannelTypes, []string{"webhook"}))
	row.ResolveChannelTypes = marshalJSON(normalizeChannelTypes(resolveChannelTypes, []string{"webhook"}))
	row.EscalationChannelTypes = marshalJSON(normalizeChannelTypes(escalationChannelTypes, []string{"slack_webhook", "webhook"}))
	row.Active = isActive
	return row, errs.CustomError{}
}

func defaultPolicyRow(tenantID string) alertmodels.AlertPolicy {
	return alertmodels.AlertPolicy{
		ID:                        idgen.New("policy"),
		TenantID:                  tenantID,
		Name:                      "Default Policy",
		MaxDeliveryAttempts:       3,
		DeliveryBackoffMillis:     200,
		EscalationIntervalMinutes: 5,
		MaxEscalations:            3,
		RepeatNotificationMinutes: 5,
		OpenChannelTypes:          marshalJSON([]string{"webhook", "slack_webhook", "email"}),
		AckChannelTypes:           marshalJSON([]string{"webhook"}),
		ResolveChannelTypes:       marshalJSON([]string{"webhook"}),
		EscalationChannelTypes:    marshalJSON([]string{"slack_webhook", "webhook"}),
		Active:                    true,
	}
}

func policyChannelTypes(policy alertmodels.AlertPolicy, eventType string) []string {
	switch eventType {
	case "incident.acknowledged":
		return unmarshalChannelTypes(policy.AckChannelTypes)
	case "incident.resolved":
		return unmarshalChannelTypes(policy.ResolveChannelTypes)
	case "incident.escalated":
		return unmarshalChannelTypes(policy.EscalationChannelTypes)
	default:
		return unmarshalChannelTypes(policy.OpenChannelTypes)
	}
}

func normalizeChannelTypes(values []string, defaults []string) []string {
	if len(values) == 0 {
		return defaults
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value != "" {
			result = append(result, value)
		}
	}
	if len(result) == 0 {
		return defaults
	}
	return result
}

func unmarshalChannelTypes(value string) []string {
	rows := make([]string, 0)
	_ = json.Unmarshal([]byte(value), &rows)
	return rows
}

func containsIgnoreCase(values []string, target string) bool {
	target = strings.ToLower(strings.TrimSpace(target))
	for _, value := range values {
		if strings.ToLower(strings.TrimSpace(value)) == target {
			return true
		}
	}
	return false
}

func marshalJSON(value interface{}) string {
	payload, _ := json.Marshal(value)
	return string(payload)
}

func stringValue(value interface{}) string {
	if typed, ok := value.(string); ok {
		return typed
	}
	return ""
}

func buildIncidentEvent(incident alertmodels.Incident, eventType string, actorID string, summary string, metadata map[string]any) *alertmodels.IncidentEvent {
	return &alertmodels.IncidentEvent{
		ID:         idgen.New("incident_event"),
		IncidentID: incident.ID,
		TenantID:   incident.TenantID,
		EventType:  eventType,
		ActorID:    actorID,
		Summary:    summary,
		Metadata:   marshalJSON(metadata),
	}
}
