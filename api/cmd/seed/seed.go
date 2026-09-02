package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	appdb "github.com/ddddami/laivan/internal/db"
	"github.com/ddddami/laivan/internal/storage"
	"github.com/jackc/pgx/v5"
)

func run(ctx context.Context, databaseURL string, maxConns int32, mediaUploader storage.Uploader) error {
	db, err := appdb.Open(ctx, databaseURL, maxConns)
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin seed transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	campusID, err := futaCampusID(ctx, tx)
	if err != nil {
		return err
	}

	if err := seedAgents(ctx, tx); err != nil {
		return err
	}
	if err := seedProperties(ctx, tx, campusID); err != nil {
		return err
	}
	if err := seedPropertyUnitTypes(ctx, tx); err != nil {
		return err
	}
	if err := seedAgentOffers(ctx, tx); err != nil {
		return err
	}
	if err := seedMedia(ctx, tx, mediaUploader); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit seed transaction: %w", err)
	}

	return nil
}

func futaCampusID(ctx context.Context, tx pgx.Tx) (string, error) {
	var campusID string
	if err := tx.QueryRow(ctx, `SELECT id FROM campuses WHERE slug = 'futa'`).Scan(&campusID); err != nil {
		return "", fmt.Errorf("get futa campus: %w", err)
	}

	return campusID, nil
}

func seedAgents(ctx context.Context, tx pgx.Tx) error {
	for _, agent := range agents {
		var userID string
		err := tx.QueryRow(ctx, `
			INSERT INTO users (email, display_name)
			VALUES ($1, $2)
			ON CONFLICT (email) DO UPDATE SET display_name = EXCLUDED.display_name
			RETURNING id
		`, agent.ID+"@seed.laivan.com", agent.DisplayName).Scan(&userID)
		if err != nil {
			return fmt.Errorf("seed user for agent %s: %w", agent.ID, err)
		}

		var seededAgentID string
		err = tx.QueryRow(ctx, `
			INSERT INTO agents (id, user_id, display_name, phone_number, whatsapp_number, status)
			VALUES ($1, $2, $3, $4, $5, 'active')
			ON CONFLICT (id) DO UPDATE SET
			  display_name = EXCLUDED.display_name,
			  phone_number = EXCLUDED.phone_number,
			  whatsapp_number = EXCLUDED.whatsapp_number,
			  status = EXCLUDED.status,
			  updated_at = now()
			WHERE agents.user_id = EXCLUDED.user_id
			RETURNING id
		`, agent.ID, userID, agent.DisplayName, agent.PhoneNumber, agent.WhatsAppNumber).Scan(&seededAgentID)
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("seed agent %s: existing agent is linked to another user", agent.ID)
		}
		if err != nil {
			return fmt.Errorf("seed agent %s: %w", agent.ID, err)
		}
	}

	return nil
}

func seedProperties(ctx context.Context, tx pgx.Tx, campusID string) error {
	for _, property := range properties {
		if _, err := tx.Exec(ctx, `
			INSERT INTO properties (id, campus_id, name, area, landmark, description, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
			ON CONFLICT (id) DO UPDATE SET
			  campus_id = EXCLUDED.campus_id,
			  name = EXCLUDED.name,
			  area = EXCLUDED.area,
			  landmark = EXCLUDED.landmark,
			  description = EXCLUDED.description,
			  created_at = EXCLUDED.created_at,
			  updated_at = now()
		`, property.ID, campusID, property.Name, property.Area, property.Landmark, property.Description, property.CreatedAt); err != nil {
			return fmt.Errorf("seed property %s: %w", property.ID, err)
		}
	}

	return nil
}

func seedPropertyUnitTypes(ctx context.Context, tx pgx.Tx) error {
	for _, unitType := range unitTypes {
		if _, err := tx.Exec(ctx, `
			INSERT INTO property_unit_types (id, property_id, category, name, description, notes, bedroom_count, has_parlour, bathroom_type, kitchen_type, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $11)
			ON CONFLICT (id) DO UPDATE SET
			  property_id = EXCLUDED.property_id,
			  category = EXCLUDED.category,
			  name = EXCLUDED.name,
			  description = EXCLUDED.description,
			  notes = EXCLUDED.notes,
			  bedroom_count = EXCLUDED.bedroom_count,
			  has_parlour = EXCLUDED.has_parlour,
			  bathroom_type = EXCLUDED.bathroom_type,
			  kitchen_type = EXCLUDED.kitchen_type,
			  created_at = EXCLUDED.created_at,
			  updated_at = now()
		`, unitType.ID, unitType.PropertyID, unitType.Category, unitType.Name, unitType.Description, unitType.Notes, unitType.BedroomCount, unitType.HasParlour, unitType.BathroomType, unitType.KitchenType, unitType.CreatedAt); err != nil {
			return fmt.Errorf("seed property unit type %s: %w", unitType.ID, err)
		}
	}

	return nil
}

func seedAgentOffers(ctx context.Context, tx pgx.Tx) error {
	for _, offer := range agentOffers {
		if _, err := tx.Exec(ctx, `
			INSERT INTO agent_offers (id, property_unit_type_id, agent_id, title, description, notes, price_kobo, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
			ON CONFLICT (property_unit_type_id, agent_id) DO UPDATE SET
			  title = EXCLUDED.title,
			  description = EXCLUDED.description,
			  notes = EXCLUDED.notes,
			  price_kobo = EXCLUDED.price_kobo,
			  status = EXCLUDED.status,
			  created_at = EXCLUDED.created_at,
			  updated_at = now()
		`, offer.ID, offer.UnitTypeID, offer.AgentID, offer.Title, offer.Description, offer.Notes, offer.PriceKobo, offer.Status, offer.CreatedAt); err != nil {
			return fmt.Errorf("seed agent offer %s: %w", offer.ID, err)
		}
	}

	return nil
}

func seedMedia(ctx context.Context, tx pgx.Tx, mediaUploader storage.Uploader) error {
	if mediaUploader == nil {
		for _, media := range mediaItems {
			if _, err := tx.Exec(ctx, `DELETE FROM media WHERE id = $1`, media.ID); err != nil {
				return fmt.Errorf("remove disabled seed media %s: %w", media.ID, err)
			}
		}
		return nil
	}

	for _, media := range mediaItems {
		data, err := seedMediaAssets.ReadFile(media.AssetPath)
		if err != nil {
			return fmt.Errorf("read seed media %s: %w", media.ID, err)
		}
		url, err := mediaUploader.Upload(ctx, storage.UploadInput{
			Key:         media.ObjectKey,
			Body:        bytes.NewReader(data),
			ContentType: media.ContentType,
		})
		if err != nil {
			return fmt.Errorf("upload seed media %s: %w", media.ID, err)
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO media (id, property_id, property_unit_type_id, agent_offer_id, uploaded_by_agent_id, url, object_key, kind, caption, content_type, size_bytes, created_at)
			VALUES ($1, NULLIF($2, '')::uuid, NULLIF($3, '')::uuid, NULLIF($4, '')::uuid, $5, $6, $7, $8, $9, $10, $11, $12)
			ON CONFLICT (id) DO UPDATE SET
			  property_id = EXCLUDED.property_id,
			  property_unit_type_id = EXCLUDED.property_unit_type_id,
			  agent_offer_id = EXCLUDED.agent_offer_id,
			  uploaded_by_agent_id = EXCLUDED.uploaded_by_agent_id,
			  url = EXCLUDED.url,
			  object_key = EXCLUDED.object_key,
			  kind = EXCLUDED.kind,
			  caption = EXCLUDED.caption,
			  content_type = EXCLUDED.content_type,
			  size_bytes = EXCLUDED.size_bytes,
			  created_at = EXCLUDED.created_at
		`, media.ID, media.PropertyID, media.UnitTypeID, media.AgentOfferID, media.UploadedByAgentID, url, media.ObjectKey, media.Kind, media.Caption, media.ContentType, len(data), media.CreatedAt); err != nil {
			return fmt.Errorf("seed media %s: %w", media.ID, err)
		}
	}

	return nil
}
