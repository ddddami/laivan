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

func (r *AgentApplicationRepository) SuspendAgent(ctx context.Context, agentID, operatorID domain.ID, operatorNote string) (domain.LinkedAgent, error) {
	return r.changeAgentStatus(ctx, agentID, operatorID, operatorNote, true)
}

func (r *AgentApplicationRepository) ReinstateAgent(ctx context.Context, agentID, operatorID domain.ID, operatorNote string) (domain.LinkedAgent, error) {
	return r.changeAgentStatus(ctx, agentID, operatorID, operatorNote, false)
}

func (r *AgentApplicationRepository) changeAgentStatus(ctx context.Context, agentID, operatorID domain.ID, operatorNote string, suspend bool) (domain.LinkedAgent, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	agentUUID, err := uuidParam(agentID)
	if err != nil {
		return domain.LinkedAgent{}, err
	}
	operatorUUID, err := uuidParam(operatorID)
	if err != nil {
		return domain.LinkedAgent{}, err
	}

	operation := "reinstate agent"
	if suspend {
		operation = "suspend agent"
	}
	tx, err := r.begin(ctx, operation)
	if err != nil {
		return domain.LinkedAgent{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)

	agent, err := queries.GetAgentForLifecycleUpdate(ctx, agentUUID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.LinkedAgent{}, ErrAgentNotFound
	}
	if err != nil {
		return domain.LinkedAgent{}, fmt.Errorf("load agent for lifecycle change: %w", err)
	}

	canManage, err := queries.CanManageAgent(ctx, generateddb.CanManageAgentParams{UserID: operatorUUID, AgentID: agentUUID})
	if err != nil {
		return domain.LinkedAgent{}, fmt.Errorf("check agent campus scope: %w", err)
	}
	if !canManage.Bool {
		return domain.LinkedAgent{}, ErrCampusForbidden
	}

	currentStatus := domain.AgentStatus(agent.Status)
	if suspend && currentStatus != domain.AgentStatusActive {
		return domain.LinkedAgent{}, ErrAgentLifecycleConflict
	}
	if !suspend && currentStatus != domain.AgentStatusSuspended {
		return domain.LinkedAgent{}, ErrAgentLifecycleConflict
	}

	var resultID pgtype.UUID
	var resultStatus string
	if suspend {
		result, err := queries.SuspendAgent(ctx, agentUUID)
		if err != nil {
			return domain.LinkedAgent{}, fmt.Errorf("suspend agent: %w", err)
		}
		resultID = result.ID
		resultStatus = result.Status
		if err := queries.SuspendActiveAgentApplications(ctx, agentUUID); err != nil {
			return domain.LinkedAgent{}, fmt.Errorf("suspend agent applications: %w", err)
		}
		if agent.UserID.Valid {
			revokedSessions, err := queries.RevokeUserSessions(ctx, agent.UserID)
			if err != nil {
				return domain.LinkedAgent{}, fmt.Errorf("revoke suspended agent sessions: %w", err)
			}
			for _, sessionID := range revokedSessions {
				if err := queries.CreateAuditEvent(ctx, generateddb.CreateAuditEventParams{
					ActorUserID:  operatorUUID,
					Action:       "session_revoked_by_agent_suspension",
					ResourceType: "session",
					ResourceID:   sessionID,
					Metadata:     []byte(`{"reason":"agent_suspended"}`),
				}); err != nil {
					return domain.LinkedAgent{}, fmt.Errorf("write session revocation audit event: %w", err)
				}
			}
		}
	} else {
		result, err := queries.ReinstateAgent(ctx, agentUUID)
		if err != nil {
			return domain.LinkedAgent{}, fmt.Errorf("reinstate agent: %w", err)
		}
		resultID = result.ID
		resultStatus = result.Status
		if err := queries.ReinstateSuspendedAgentApplications(ctx, agentUUID); err != nil {
			return domain.LinkedAgent{}, fmt.Errorf("reinstate agent applications: %w", err)
		}
	}

	action := "agent_reinstated"
	if suspend {
		action = "agent_suspended"
	}
	metadata, err := json.Marshal(map[string]any{
		"before":        map[string]string{"status": string(currentStatus)},
		"after":         map[string]string{"status": resultStatus},
		"operator_note": operatorNote,
	})
	if err != nil {
		return domain.LinkedAgent{}, fmt.Errorf("encode agent lifecycle audit metadata: %w", err)
	}
	if err := queries.CreateAuditEvent(ctx, generateddb.CreateAuditEventParams{
		ActorUserID:  operatorUUID,
		Action:       action,
		ResourceType: "agent",
		ResourceID:   agentUUID,
		Metadata:     metadata,
	}); err != nil {
		return domain.LinkedAgent{}, fmt.Errorf("write agent lifecycle audit event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.LinkedAgent{}, fmt.Errorf("commit agent lifecycle change: %w", err)
	}
	return domain.LinkedAgent{ID: domain.ID(uuidString(resultID)), Status: domain.AgentStatus(resultStatus)}, nil
}
