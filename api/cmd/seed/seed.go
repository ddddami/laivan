package main

import (
	"context"
	"fmt"

	appdb "github.com/ddddami/laivan/internal/db"
	"github.com/jackc/pgx/v5"
)

func run(ctx context.Context, databaseURL string) error {
	db, err := appdb.Open(ctx, databaseURL)
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
		if _, err := tx.Exec(ctx, `
			INSERT INTO agents (id, display_name, phone_number, whatsapp_number)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO UPDATE SET
			  display_name = EXCLUDED.display_name,
			  phone_number = EXCLUDED.phone_number,
			  whatsapp_number = EXCLUDED.whatsapp_number,
			  updated_at = now()
		`, agent.ID, agent.DisplayName, agent.PhoneNumber, agent.WhatsAppNumber); err != nil {
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
			INSERT INTO property_unit_types (id, property_id, category, name, description, bedroom_count, has_parlour, bathroom_type, kitchen_type, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10)
			ON CONFLICT (id) DO UPDATE SET
			  property_id = EXCLUDED.property_id,
			  category = EXCLUDED.category,
			  name = EXCLUDED.name,
			  description = EXCLUDED.description,
			  bedroom_count = EXCLUDED.bedroom_count,
			  has_parlour = EXCLUDED.has_parlour,
			  bathroom_type = EXCLUDED.bathroom_type,
			  kitchen_type = EXCLUDED.kitchen_type,
			  created_at = EXCLUDED.created_at,
			  updated_at = now()
		`, unitType.ID, unitType.PropertyID, unitType.Category, unitType.Name, unitType.Description, unitType.BedroomCount, unitType.HasParlour, unitType.BathroomType, unitType.KitchenType, unitType.CreatedAt); err != nil {
			return fmt.Errorf("seed property unit type %s: %w", unitType.ID, err)
		}
	}

	return nil
}

func seedAgentOffers(ctx context.Context, tx pgx.Tx) error {
	for _, offer := range agentOffers {
		if _, err := tx.Exec(ctx, `
			INSERT INTO agent_offers (id, property_unit_type_id, agent_id, title, description, price_kobo, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
			ON CONFLICT (property_unit_type_id, agent_id) DO UPDATE SET
			  title = EXCLUDED.title,
			  description = EXCLUDED.description,
			  price_kobo = EXCLUDED.price_kobo,
			  status = EXCLUDED.status,
			  created_at = EXCLUDED.created_at,
			  updated_at = now()
		`, offer.ID, offer.UnitTypeID, offer.AgentID, offer.Title, offer.Description, offer.PriceKobo, offer.Status, offer.CreatedAt); err != nil {
			return fmt.Errorf("seed agent offer %s: %w", offer.ID, err)
		}
	}

	return nil
}
