package repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	generateddb "github.com/ddddami/laivan/internal/db/generated"
	"github.com/ddddami/laivan/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type AgentApplicationRepository struct {
	queries *generateddb.Queries
	db      generateddb.DBTX
}

func NewAgentApplicationRepository(database generateddb.DBTX) *AgentApplicationRepository {
	return &AgentApplicationRepository{queries: generateddb.New(database), db: database}
}

func (r *AgentApplicationRepository) GetEffectiveAccess(ctx context.Context, userID domain.ID) (domain.EffectiveAccess, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	uuid, err := uuidParam(userID)
	if err != nil {
		return domain.EffectiveAccess{}, err
	}

	globalAdmin, err := r.queries.IsGlobalAdmin(ctx, uuid)
	if err != nil {
		return domain.EffectiveAccess{}, fmt.Errorf("get global admin role: %w", err)
	}
	operatorCampuses, err := r.queries.ListCampusOperatorCampuses(ctx, uuid)
	if err != nil {
		return domain.EffectiveAccess{}, fmt.Errorf("list campus operator roles: %w", err)
	}

	access := domain.EffectiveAccess{GlobalAdmin: globalAdmin, Roles: make([]string, 0, 3)}
	if globalAdmin {
		access.Roles = append(access.Roles, "global_admin")
	}
	if len(operatorCampuses) > 0 {
		access.Roles = append(access.Roles, "campus_operator")
		access.CampusOperatorIDs = make([]domain.ID, 0, len(operatorCampuses))
		for _, campusID := range operatorCampuses {
			access.CampusOperatorIDs = append(access.CampusOperatorIDs, domain.ID(uuidString(campusID)))
		}
	}

	agent, err := r.queries.GetLinkedAgentAccess(ctx, uuid)
	if err == nil {
		access.Agent = &domain.LinkedAgent{ID: domain.ID(uuidString(agent.ID)), Status: domain.AgentStatus(agent.Status)}
		if access.Agent.Status == domain.AgentStatusActive {
			access.Roles = append(access.Roles, "active_agent")
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return domain.EffectiveAccess{}, fmt.Errorf("get linked agent access: %w", err)
	}

	return access, nil
}

func (r *AgentApplicationRepository) CreateApplication(ctx context.Context, application domain.AgentApplication) (domain.AgentApplication, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	applicantID, err := uuidParam(application.ApplicantUserID)
	if err != nil {
		return domain.AgentApplication{}, err
	}
	campusID, err := uuidParam(application.CampusID)
	if err != nil {
		return domain.AgentApplication{}, err
	}

	beginner, ok := r.db.(interface {
		Begin(context.Context) (pgx.Tx, error)
	})
	if !ok {
		return domain.AgentApplication{}, errors.New("agent application repository database does not support transactions")
	}
	tx, err := beginner.Begin(ctx)
	if err != nil {
		return domain.AgentApplication{}, fmt.Errorf("begin create agent application transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	queries := r.queries.WithTx(tx)
	if _, err := queries.LockUserForAgentApplication(ctx, applicantID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AgentApplication{}, ErrNotFound
		}
		return domain.AgentApplication{}, fmt.Errorf("lock applicant user: %w", err)
	}
	if _, err := queries.GetLinkedAgentAccess(ctx, applicantID); err == nil {
		return domain.AgentApplication{}, ErrAlreadyLinked
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return domain.AgentApplication{}, fmt.Errorf("check linked agent: %w", err)
	}

	id, err := queries.CreateAgentApplication(ctx, generateddb.CreateAgentApplicationParams{
		ApplicantUserID: applicantID,
		CampusID:        campusID,
		Name:            application.Name,
		PhoneNumber:     application.PhoneNumber,
	})
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return domain.AgentApplication{}, ErrNotFound
		case isUniqueViolation(err):
			return domain.AgentApplication{}, ErrApplicationConflict
		case isForeignKeyViolation(err):
			return domain.AgentApplication{}, ErrNotFound
		default:
			return domain.AgentApplication{}, fmt.Errorf("create agent application: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.AgentApplication{}, fmt.Errorf("commit create agent application: %w", err)
	}

	return r.GetApplication(ctx, domain.ID(uuidString(id)))
}

func (r *AgentApplicationRepository) GetApplication(ctx context.Context, id domain.ID) (domain.AgentApplication, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	uuid, err := uuidParam(id)
	if err != nil {
		return domain.AgentApplication{}, err
	}
	row, err := r.queries.GetAgentApplication(ctx, uuid)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.AgentApplication{}, ErrNotFound
	}
	if err != nil {
		return domain.AgentApplication{}, fmt.Errorf("get agent application: %w", err)
	}
	return agentApplicationFromRow(row), nil
}

func (r *AgentApplicationRepository) ListApplications(ctx context.Context, applicantID domain.ID) ([]domain.AgentApplication, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	uuid, err := uuidParam(applicantID)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListAgentApplicationsByApplicant(ctx, uuid)
	if err != nil {
		return nil, fmt.Errorf("list applicant agent applications: %w", err)
	}
	applications := make([]domain.AgentApplication, 0, len(rows))
	for _, row := range rows {
		applications = append(applications, agentApplicationFromApplicantRow(row))
	}
	return applications, nil
}

func (r *AgentApplicationRepository) ListOperatorApplications(ctx context.Context, operatorID, campusID domain.ID, status domain.AgentApplicationStatus) ([]domain.AgentApplication, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	operatorUUID, err := uuidParam(operatorID)
	if err != nil {
		return nil, err
	}
	campusUUID, err := uuidParam(campusID)
	if err != nil {
		return nil, err
	}
	canReview, err := r.queries.CanReviewCampus(ctx, generateddb.CanReviewCampusParams{UserID: operatorUUID, CampusID: campusUUID})
	if err != nil {
		return nil, fmt.Errorf("check campus operator scope: %w", err)
	}
	if !canReview.Bool {
		return nil, ErrCampusForbidden
	}
	rows, err := r.queries.ListAgentApplicationsByCampus(ctx, generateddb.ListAgentApplicationsByCampusParams{CampusID: campusUUID, Column2: string(status)})
	if err != nil {
		return nil, fmt.Errorf("list campus agent applications: %w", err)
	}
	applications := make([]domain.AgentApplication, 0, len(rows))
	for _, row := range rows {
		applications = append(applications, agentApplicationFromCampusRow(row))
	}
	return applications, nil
}

func (r *AgentApplicationRepository) ActivateApplication(ctx context.Context, applicationID, operatorID domain.ID, legacyAgentID *domain.ID, operatorNote string) (domain.AgentApplication, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	applicationUUID, err := uuidParam(applicationID)
	if err != nil {
		return domain.AgentApplication{}, err
	}
	operatorUUID, err := uuidParam(operatorID)
	if err != nil {
		return domain.AgentApplication{}, err
	}

	tx, err := r.begin(ctx, "activate agent application")
	if err != nil {
		return domain.AgentApplication{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)

	row, err := queries.GetAgentApplicationForUpdate(ctx, applicationUUID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.AgentApplication{}, ErrNotFound
	}
	if err != nil {
		return domain.AgentApplication{}, fmt.Errorf("load agent application for activation: %w", err)
	}
	canReview, err := queries.CanReviewCampus(ctx, generateddb.CanReviewCampusParams{UserID: operatorUUID, CampusID: row.CampusID})
	if err != nil {
		return domain.AgentApplication{}, fmt.Errorf("check activation campus scope: %w", err)
	}
	if !canReview.Bool {
		return domain.AgentApplication{}, ErrCampusForbidden
	}
	if row.Status != string(domain.AgentApplicationStatusPending) {
		return domain.AgentApplication{}, ErrApplicationResolved
	}
	if _, err := queries.LockUserForAgentApplication(ctx, row.ApplicantUserID); err != nil {
		return domain.AgentApplication{}, fmt.Errorf("lock applicant user for activation: %w", err)
	}
	if _, err := queries.GetLinkedAgentAccess(ctx, row.ApplicantUserID); err == nil {
		return domain.AgentApplication{}, ErrAlreadyLinked
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return domain.AgentApplication{}, fmt.Errorf("check linked agent before activation: %w", err)
	}

	var agentID pgtype.UUID
	if legacyAgentID != nil {
		legacyUUID, err := uuidParam(*legacyAgentID)
		if err != nil {
			return domain.AgentApplication{}, err
		}
		legacy, err := queries.GetAgent(ctx, legacyUUID)
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AgentApplication{}, ErrNotFound
		}
		if err != nil {
			return domain.AgentApplication{}, fmt.Errorf("load legacy agent: %w", err)
		}
		if legacy.UserID.Valid || legacy.Status != string(domain.AgentStatusActive) {
			return domain.AgentApplication{}, ErrLegacyAgentConflict
		}
		if _, err := queries.LinkAgentToUser(ctx, generateddb.LinkAgentToUserParams{ID: legacyUUID, UserID: row.ApplicantUserID}); err != nil {
			if errors.Is(err, pgx.ErrNoRows) || isUniqueViolation(err) {
				return domain.AgentApplication{}, ErrLegacyAgentConflict
			}
			return domain.AgentApplication{}, fmt.Errorf("link legacy agent: %w", err)
		}
		agentID = legacyUUID
	} else {
		created, err := queries.CreateAgent(ctx, generateddb.CreateAgentParams{UserID: row.ApplicantUserID, DisplayName: row.Name, PhoneNumber: row.PhoneNumber})
		if err != nil {
			if isUniqueViolation(err) {
				return domain.AgentApplication{}, ErrPhoneConflict
			}
			return domain.AgentApplication{}, fmt.Errorf("create agent from application: %w", err)
		}
		agentID = created.ID
	}

	if _, err := queries.ActivateAgentApplication(ctx, generateddb.ActivateAgentApplicationParams{
		ID:             applicationUUID,
		ReviewerUserID: operatorUUID,
		AgentID:        agentID,
		Column4:        operatorNote,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AgentApplication{}, ErrApplicationResolved
		}
		return domain.AgentApplication{}, fmt.Errorf("activate agent application: %w", err)
	}

	metadata, err := json.Marshal(map[string]any{
		"before":          map[string]string{"status": string(domain.AgentApplicationStatusPending)},
		"after":           map[string]string{"status": string(domain.AgentApplicationStatusActive)},
		"agent_id":        uuidString(agentID),
		"legacy_agent_id": idOrEmpty(legacyAgentID),
	})
	if err != nil {
		return domain.AgentApplication{}, fmt.Errorf("encode activation audit metadata: %w", err)
	}
	action := "agent_application_activated"
	if legacyAgentID != nil {
		action = "agent_application_linked"
	}
	if err := queries.CreateAuditEvent(ctx, generateddb.CreateAuditEventParams{
		ActorUserID:  operatorUUID,
		Action:       action,
		ResourceType: "agent_application",
		ResourceID:   applicationUUID,
		Metadata:     metadata,
	}); err != nil {
		return domain.AgentApplication{}, fmt.Errorf("write activation audit event: %w", err)
	}

	result, err := queries.GetAgentApplication(ctx, applicationUUID)
	if err != nil {
		return domain.AgentApplication{}, fmt.Errorf("load activated application: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.AgentApplication{}, fmt.Errorf("commit agent application activation: %w", err)
	}
	return agentApplicationFromRow(result), nil
}

func (r *AgentApplicationRepository) DeclineApplication(ctx context.Context, applicationID, operatorID domain.ID, operatorNote string) (domain.AgentApplication, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	applicationUUID, err := uuidParam(applicationID)
	if err != nil {
		return domain.AgentApplication{}, err
	}
	operatorUUID, err := uuidParam(operatorID)
	if err != nil {
		return domain.AgentApplication{}, err
	}
	tx, err := r.begin(ctx, "decline agent application")
	if err != nil {
		return domain.AgentApplication{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)

	row, err := queries.GetAgentApplicationForUpdate(ctx, applicationUUID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.AgentApplication{}, ErrNotFound
	}
	if err != nil {
		return domain.AgentApplication{}, fmt.Errorf("load agent application for decline: %w", err)
	}
	canReview, err := queries.CanReviewCampus(ctx, generateddb.CanReviewCampusParams{UserID: operatorUUID, CampusID: row.CampusID})
	if err != nil {
		return domain.AgentApplication{}, fmt.Errorf("check decline campus scope: %w", err)
	}
	if !canReview.Bool {
		return domain.AgentApplication{}, ErrCampusForbidden
	}
	if row.Status != string(domain.AgentApplicationStatusPending) {
		return domain.AgentApplication{}, ErrApplicationResolved
	}
	if _, err := queries.DeclineAgentApplication(ctx, generateddb.DeclineAgentApplicationParams{ID: applicationUUID, ReviewerUserID: operatorUUID, Column3: operatorNote}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AgentApplication{}, ErrApplicationResolved
		}
		return domain.AgentApplication{}, fmt.Errorf("decline agent application: %w", err)
	}
	metadata, err := json.Marshal(map[string]any{
		"before": map[string]string{"status": string(domain.AgentApplicationStatusPending)},
		"after":  map[string]string{"status": string(domain.AgentApplicationStatusDeclined)},
	})
	if err != nil {
		return domain.AgentApplication{}, fmt.Errorf("encode decline audit metadata: %w", err)
	}
	if err := queries.CreateAuditEvent(ctx, generateddb.CreateAuditEventParams{
		ActorUserID:  operatorUUID,
		Action:       "agent_application_declined",
		ResourceType: "agent_application",
		ResourceID:   applicationUUID,
		Metadata:     metadata,
	}); err != nil {
		return domain.AgentApplication{}, fmt.Errorf("write decline audit event: %w", err)
	}
	result, err := queries.GetAgentApplication(ctx, applicationUUID)
	if err != nil {
		return domain.AgentApplication{}, fmt.Errorf("load declined application: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.AgentApplication{}, fmt.Errorf("commit agent application decline: %w", err)
	}
	return agentApplicationFromRow(result), nil
}

func (r *AgentApplicationRepository) begin(ctx context.Context, operation string) (pgx.Tx, error) {
	beginner, ok := r.db.(interface {
		Begin(context.Context) (pgx.Tx, error)
	})
	if !ok {
		return nil, fmt.Errorf("%s database does not support transactions", operation)
	}
	tx, err := beginner.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin %s transaction: %w", operation, err)
	}
	return tx, nil
}

func agentApplicationFromRow(row generateddb.GetAgentApplicationRow) domain.AgentApplication {
	return domain.AgentApplication{
		ID:                   domain.ID(uuidString(row.ID)),
		ApplicantUserID:      domain.ID(uuidString(row.ApplicantUserID)),
		ApplicantEmail:       row.ApplicantEmail,
		ApplicantDisplayName: row.ApplicantDisplayName,
		CampusID:             domain.ID(uuidString(row.CampusID)),
		Name:                 row.Name,
		PhoneNumber:          row.PhoneNumber,
		Status:               domain.AgentApplicationStatus(row.Status),
		ReviewerUserID:       nullableUUID(row.ReviewerUserID),
		AgentID:              nullableUUID(row.AgentID),
		OperatorNote:         textString(row.OperatorNote),
		CreatedAt:            row.CreatedAt.Time,
		UpdatedAt:            row.UpdatedAt.Time,
		DecidedAt:            timestamptzPointer(row.DecidedAt),
	}
}

func agentApplicationFromApplicantRow(row generateddb.ListAgentApplicationsByApplicantRow) domain.AgentApplication {
	return agentApplicationFromRow(generateddb.GetAgentApplicationRow(row))
}

func agentApplicationFromCampusRow(row generateddb.ListAgentApplicationsByCampusRow) domain.AgentApplication {
	return agentApplicationFromRow(generateddb.GetAgentApplicationRow(row))
}

func nullableUUID(value pgtype.UUID) *domain.ID {
	if !value.Valid {
		return nil
	}
	id := domain.ID(uuidString(value))
	return &id
}

func idOrEmpty(value *domain.ID) string {
	if value == nil {
		return ""
	}
	return string(*value)
}
