//go:build integration

package repo

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ddddami/laivan/internal/domain"
)

func TestInquiryRepositorySubmitPersistsAndDeduplicates(t *testing.T) {
	ctx := t.Context()
	db := openIntegrationDB(t, ctx)
	t.Cleanup(db.Close)

	truncateProperties(t, ctx, db)
	truncateAgents(t, ctx, db)
	t.Cleanup(func() {
		truncateProperties(t, context.Background(), db)
		truncateAgents(t, context.Background(), db)
	})

	campusID := testCampusID(t, ctx, db)
	propertyID := insertProperty(t, ctx, db, campusID, "Inquiry Lodge", time.Now().UTC())
	unitTypeID := insertPropertyUnitType(t, ctx, db, propertyID, "Self-contained")
	agentID := insertAgent(t, ctx, db, "Inquiry Agent")
	associateAgentWithCampus(t, ctx, db, agentID, campusID)
	agentUserID := agentUserID(t, ctx, db, agentID)
	studentUserID := insertUser(t, ctx, db, "Inquiry Student")
	offerID := insertAgentOffer(t, ctx, db, unitTypeID, agentID, "Fresh self-contained room", 35000000)

	if _, err := db.Exec(ctx, "UPDATE agents SET whatsapp_number = '+2348012345678' WHERE id = $1", string(agentID)); err != nil {
		t.Fatalf("set WhatsApp number: %v", err)
	}

	repository := NewInquiryRepository(db)
	submissionID := domain.ID("550e8400-e29b-41d4-a716-446655440020")
	created, handoff, isNew, err := repository.Submit(ctx, studentUserID, offerID, submissionID, "  Is the kitchen private?  ")
	if err != nil {
		t.Fatalf("submit inquiry: %v", err)
	}
	if !isNew || created.Status != domain.InquiryStatusOpen {
		t.Fatalf("created inquiry = %#v, isNew = %t", created, isNew)
	}
	if created.Message != "Is the kitchen private?" {
		t.Fatalf("stored message = %q, want trimmed message", created.Message)
	}
	if handoff.WhatsAppNumber != "+2348012345678" {
		t.Fatalf("WhatsApp number = %q, want override", handoff.WhatsAppNumber)
	}

	replayed, _, isNew, err := repository.Submit(ctx, studentUserID, offerID, submissionID, "A different retry message")
	if err != nil {
		t.Fatalf("replay inquiry: %v", err)
	}
	if isNew || replayed.ID != created.ID || replayed.Message != created.Message {
		t.Fatalf("replayed inquiry = %#v, isNew = %t", replayed, isNew)
	}

	var inquiryCount int
	if err := db.QueryRow(ctx, "SELECT count(*) FROM inquiries WHERE student_user_id = $1", string(studentUserID)).Scan(&inquiryCount); err != nil {
		t.Fatalf("count inquiries: %v", err)
	}
	if inquiryCount != 1 {
		t.Fatalf("inquiry count = %d, want 1 after replay", inquiryCount)
	}

	secondSubmissionID := domain.ID("550e8400-e29b-41d4-a716-446655440021")
	second, _, isNew, err := repository.Submit(ctx, studentUserID, offerID, secondSubmissionID, "When can I inspect it?")
	if err != nil {
		t.Fatalf("submit second inquiry: %v", err)
	}
	if !isNew || second.ID == created.ID {
		t.Fatalf("second inquiry = %#v, isNew = %t", second, isNew)
	}

	if _, err := db.Exec(ctx, "UPDATE agents SET whatsapp_number = NULL WHERE id = $1", string(agentID)); err != nil {
		t.Fatalf("clear WhatsApp number: %v", err)
	}
	fallback, handoff, _, err := repository.Submit(ctx, studentUserID, offerID, domain.ID("550e8400-e29b-41d4-a716-446655440022"), "Can I visit today?")
	if err != nil {
		t.Fatalf("submit fallback inquiry: %v", err)
	}
	if handoff.WhatsAppNumber != "" || handoff.PhoneNumber == "" || fallback.ID == "" {
		t.Fatalf("fallback handoff = %#v", handoff)
	}

	var concurrentIDs = make(chan domain.ID, 2)
	var concurrentErrors = make(chan error, 2)
	concurrentSubmissionID := domain.ID("550e8400-e29b-41d4-a716-446655440023")
	var waitGroup sync.WaitGroup
	waitGroup.Go(func() {
		inquiry, _, _, err := repository.Submit(ctx, studentUserID, offerID, concurrentSubmissionID, "Concurrent question")
		if err != nil {
			concurrentErrors <- err
			return
		}
		concurrentIDs <- inquiry.ID
	})
	waitGroup.Go(func() {
		inquiry, _, _, err := repository.Submit(ctx, studentUserID, offerID, concurrentSubmissionID, "Concurrent question")
		if err != nil {
			concurrentErrors <- err
			return
		}
		concurrentIDs <- inquiry.ID
	})
	waitGroup.Wait()
	close(concurrentIDs)
	close(concurrentErrors)
	for err := range concurrentErrors {
		t.Fatalf("concurrent inquiry: %v", err)
	}
	var ids []domain.ID
	for id := range concurrentIDs {
		ids = append(ids, id)
	}
	if len(ids) != 2 || ids[0] != ids[1] {
		t.Fatalf("concurrent inquiry IDs = %v, want two identical IDs", ids)
	}

	_, _, _, err = repository.Submit(ctx, agentUserID, offerID, domain.ID("550e8400-e29b-41d4-a716-446655440024"), "My own offer")
	if !errors.Is(err, ErrOwnOffer) {
		t.Fatalf("own offer error = %v, want %v", err, ErrOwnOffer)
	}

	if _, err := db.Exec(ctx, "UPDATE agent_offers SET status = 'paused' WHERE id = $1", string(offerID)); err != nil {
		t.Fatalf("pause offer: %v", err)
	}
	_, _, _, err = repository.Submit(ctx, studentUserID, offerID, domain.ID("550e8400-e29b-41d4-a716-446655440025"), "Paused offer")
	if !errors.Is(err, ErrOfferUnavailable) {
		t.Fatalf("paused offer error = %v, want %v", err, ErrOfferUnavailable)
	}

	otherStudentID := insertUser(t, ctx, db, "Other Inquiry Student")
	otherStudentInquiry, _, isNew, err := repository.Submit(ctx, otherStudentID, offerID, submissionID, "Can I see the kitchen?")
	if !errors.Is(err, ErrOfferUnavailable) {
		t.Fatalf("other student's paused offer error = %v, want %v", err, ErrOfferUnavailable)
	}
	if otherStudentInquiry.ID != "" || isNew {
		t.Fatalf("other student's paused inquiry = %#v, isNew = %t", otherStudentInquiry, isNew)
	}

	if _, err := db.Exec(ctx, "UPDATE agent_offers SET status = 'available' WHERE id = $1", string(offerID)); err != nil {
		t.Fatalf("restore offer availability: %v", err)
	}
	otherStudentInquiry, _, isNew, err = repository.Submit(ctx, otherStudentID, offerID, submissionID, "Can I see the kitchen?")
	if err != nil {
		t.Fatalf("submit same ID for another student: %v", err)
	}
	if !isNew || otherStudentInquiry.ID == created.ID || otherStudentInquiry.StudentUserID != otherStudentID {
		t.Fatalf("other student's inquiry = %#v, isNew = %t", otherStudentInquiry, isNew)
	}

	deniedCases := []struct {
		name         string
		offerID      domain.ID
		submissionID domain.ID
		prepare      func(t *testing.T)
		want         error
	}{
		{
			name:         "missing offer",
			offerID:      "550e8400-e29b-41d4-a716-44665544002b",
			submissionID: "550e8400-e29b-41d4-a716-446655440026",
			want:         ErrNotFound,
		},
		{
			name:         "unavailable offer",
			submissionID: "550e8400-e29b-41d4-a716-446655440027",
			prepare: func(t *testing.T) {
				if _, err := db.Exec(ctx, "UPDATE agent_offers SET status = 'unavailable' WHERE id = $1", string(offerID)); err != nil {
					t.Fatalf("make offer unavailable: %v", err)
				}
			},
			want: ErrOfferUnavailable,
		},
		{
			name:         "archived offer",
			submissionID: "550e8400-e29b-41d4-a716-446655440028",
			prepare: func(t *testing.T) {
				if _, err := db.Exec(ctx, "UPDATE agent_offers SET status = 'available', archived_at = now() WHERE id = $1", string(offerID)); err != nil {
					t.Fatalf("archive offer: %v", err)
				}
			},
			want: ErrNotFound,
		},
		{
			name:         "suspended agent",
			submissionID: "550e8400-e29b-41d4-a716-446655440029",
			prepare: func(t *testing.T) {
				if _, err := db.Exec(ctx, "UPDATE agent_offers SET archived_at = NULL WHERE id = $1", string(offerID)); err != nil {
					t.Fatalf("restore offer: %v", err)
				}
				if _, err := db.Exec(ctx, "UPDATE agents SET status = 'suspended' WHERE id = $1", string(agentID)); err != nil {
					t.Fatalf("suspend agent: %v", err)
				}
			},
			want: ErrOfferUnavailable,
		},
		{
			name:         "agent no longer in campus",
			submissionID: "550e8400-e29b-41d4-a716-44665544002a",
			prepare: func(t *testing.T) {
				if _, err := db.Exec(ctx, "UPDATE agents SET status = 'active' WHERE id = $1", string(agentID)); err != nil {
					t.Fatalf("reactivate agent: %v", err)
				}
				if _, err := db.Exec(ctx, "DELETE FROM agent_campuses WHERE agent_id = $1 AND campus_id = $2", string(agentID), string(campusID)); err != nil {
					t.Fatalf("remove agent campus association: %v", err)
				}
			},
			want: ErrOfferUnavailable,
		},
	}

	for _, testCase := range deniedCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.prepare != nil {
				testCase.prepare(t)
			}
			targetOfferID := offerID
			if testCase.offerID != "" {
				targetOfferID = testCase.offerID
			}
			_, _, _, err := repository.Submit(ctx, studentUserID, targetOfferID, testCase.submissionID, "Is this available?")
			if !errors.Is(err, testCase.want) {
				t.Fatalf("submit error = %v, want %v", err, testCase.want)
			}
			var count int
			if err := db.QueryRow(ctx, "SELECT count(*) FROM inquiries WHERE student_user_id = $1 AND submission_id = $2", string(studentUserID), string(testCase.submissionID)).Scan(&count); err != nil {
				t.Fatalf("count denied inquiries: %v", err)
			}
			if count != 0 {
				t.Fatalf("denied inquiry count = %d, want 0", count)
			}
		})
	}
}
