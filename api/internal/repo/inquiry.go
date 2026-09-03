package repo

import (
	"context"
	"errors"
	"fmt"

	generateddb "github.com/ddddami/laivan/internal/db/generated"
	"github.com/ddddami/laivan/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type InquiryRepository struct {
	queries *generateddb.Queries
}

func NewInquiryRepository(database generateddb.DBTX) *InquiryRepository {
	return &InquiryRepository{queries: generateddb.New(database)}
}

func (r *InquiryRepository) Submit(ctx context.Context, studentUserID, agentOfferID, submissionID domain.ID, message string) (domain.Inquiry, domain.InquiryHandoff, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	studentUUID, err := uuidParam(studentUserID)
	if err != nil {
		return domain.Inquiry{}, domain.InquiryHandoff{}, false, err
	}
	offerUUID, err := uuidParam(agentOfferID)
	if err != nil {
		return domain.Inquiry{}, domain.InquiryHandoff{}, false, err
	}
	submissionUUID, err := uuidParam(submissionID)
	if err != nil {
		return domain.Inquiry{}, domain.InquiryHandoff{}, false, err
	}

	inquiryRow, err := r.queries.FindInquiryByStudentSubmission(ctx, generateddb.FindInquiryByStudentSubmissionParams{
		StudentUserID: studentUUID,
		SubmissionID:  submissionUUID,
	})
	created := true
	if errors.Is(err, pgx.ErrNoRows) {
		inquiryRow, err = r.queries.CreateInquiryForAvailableOffer(ctx, generateddb.CreateInquiryForAvailableOfferParams{
			StudentUserID: studentUUID,
			AgentOfferID:  offerUUID,
			SubmissionID:  submissionUUID,
			Message:       message,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Inquiry{}, domain.InquiryHandoff{}, false, r.classifyOffer(ctx, offerUUID, studentUUID)
		}
		if isUniqueViolation(err) {
			inquiryRow, err = r.queries.FindInquiryByStudentSubmission(ctx, generateddb.FindInquiryByStudentSubmissionParams{
				StudentUserID: studentUUID,
				SubmissionID:  submissionUUID,
			})
			created = false
		}
	} else if err == nil {
		created = false
	}
	if err != nil {
		return domain.Inquiry{}, domain.InquiryHandoff{}, false, fmt.Errorf("submit inquiry: %w", err)
	}

	inquiry := inquiryFromRow(inquiryRow)
	handoffRow, err := r.queries.GetInquiryHandoff(ctx, inquiryRow.ID)
	if err != nil {
		return domain.Inquiry{}, domain.InquiryHandoff{}, false, fmt.Errorf("get inquiry handoff: %w", err)
	}

	return inquiry, inquiryHandoffFromRow(handoffRow), created, nil
}

func (r *InquiryRepository) classifyOffer(ctx context.Context, offerID, studentUserID pgtype.UUID) error {
	state, err := r.queries.GetInquiryOfferState(ctx, offerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("classify inquiry offer: %w", err)
	}
	if state.ArchivedAt.Valid {
		return ErrNotFound
	}
	if state.AgentUserID.Valid && state.AgentUserID == studentUserID {
		return ErrOwnOffer
	}
	// Every existing offer that failed the atomic eligibility predicate is unavailable.
	return ErrOfferUnavailable
}

func inquiryFromRow(row generateddb.Inquiry) domain.Inquiry {
	return domain.Inquiry{
		ID:            domain.ID(uuidString(row.ID)),
		StudentUserID: domain.ID(uuidString(row.StudentUserID)),
		AgentOfferID:  domain.ID(uuidString(row.AgentOfferID)),
		SubmissionID:  domain.ID(uuidString(row.SubmissionID)),
		Message:       row.Message,
		Status:        domain.InquiryStatus(row.Status),
		Timestamps: domain.Timestamps{
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		},
	}
}

func inquiryHandoffFromRow(row generateddb.GetInquiryHandoffRow) domain.InquiryHandoff {
	return domain.InquiryHandoff{
		AgentDisplayName: row.AgentDisplayName,
		PhoneNumber:      row.PhoneNumber,
		WhatsAppNumber:   textString(row.WhatsappNumber),
		OfferTitle:       row.OfferTitle,
		PropertyName:     row.PropertyName,
		UnitName:         row.UnitName,
		Area:             row.Area,
	}
}
