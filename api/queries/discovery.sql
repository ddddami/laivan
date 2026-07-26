-- name: DiscoverProperties :many
WITH opportunities AS (
  SELECT
    p.id AS property_id,
    p.name AS property_name,
    p.area AS property_area,
    p.landmark AS property_landmark,
    put.id AS unit_type_id,
    put.category,
    put.name AS unit_type_name,
    put.description,
    put.notes,
    put.bedroom_count,
    put.has_parlour,
    put.bathroom_type,
    put.kitchen_type,
    MIN(ao.price_kobo)::integer AS lowest_price_kobo,
    COUNT(ao.id)::integer AS available_offer_count,
    (
      SELECT m.url
      FROM media m
      WHERE m.kind = 'image'
        AND (
          m.property_id = p.id
          OR m.property_unit_type_id = put.id
        )
      ORDER BY
        CASE WHEN m.property_unit_type_id = put.id THEN 0 ELSE 1 END,
        m.created_at,
        m.id
      LIMIT 1
    ) AS thumbnail_url,
    p.created_at,
    GREATEST(
      p.updated_at,
      put.updated_at,
      COALESCE(MAX(ao.updated_at), p.updated_at)
    ) AS updated_at,
    (
      CASE WHEN NULLIF(BTRIM(p.landmark), '') IS NOT NULL THEN 1 ELSE 0 END +
      CASE WHEN NULLIF(BTRIM(p.description), '') IS NOT NULL THEN 1 ELSE 0 END +
      CASE WHEN NULLIF(BTRIM(put.description), '') IS NOT NULL THEN 1 ELSE 0 END +
      CASE WHEN NULLIF(BTRIM(put.notes), '') IS NOT NULL THEN 1 ELSE 0 END +
      CASE WHEN put.bedroom_count IS NOT NULL THEN 1 ELSE 0 END +
      CASE WHEN put.bathroom_type IS NOT NULL AND put.bathroom_type <> 'unknown' THEN 1 ELSE 0 END +
      CASE WHEN put.kitchen_type IS NOT NULL AND put.kitchen_type <> 'unknown' THEN 1 ELSE 0 END
    )::integer AS completeness_score
  FROM properties p
  JOIN property_unit_types put ON put.property_id = p.id
  LEFT JOIN agent_offers ao
    ON ao.property_unit_type_id = put.id
    AND ao.status = 'available'
  WHERE p.campus_id = sqlc.arg(campus_id)
    AND (
      sqlc.arg(search)::text = ''
      OR STRPOS(LOWER(p.name), LOWER(sqlc.arg(search)::text)) > 0
      OR STRPOS(LOWER(p.area), LOWER(sqlc.arg(search)::text)) > 0
      OR STRPOS(LOWER(COALESCE(p.landmark, '')), LOWER(sqlc.arg(search)::text)) > 0
      OR STRPOS(LOWER(COALESCE(put.name, '')), LOWER(sqlc.arg(search)::text)) > 0
      OR STRPOS(
        REPLACE(LOWER(put.category), '_', ' '),
        REPLACE(LOWER(sqlc.arg(search)::text), '-', ' ')
      ) > 0
    )
    AND (
      cardinality(sqlc.arg(categories)::text[]) = 0
      OR put.category = ANY(sqlc.arg(categories)::text[])
    )
    AND (
      sqlc.arg(area)::text = ''
      OR p.area ILIKE '%' || sqlc.arg(area)::text || '%'
    )
    AND (
      sqlc.arg(bathroom_type)::text = ''
      OR put.bathroom_type = sqlc.arg(bathroom_type)::text
    )
    AND (
      sqlc.arg(kitchen_type)::text = ''
      OR put.kitchen_type = sqlc.arg(kitchen_type)::text
    )
    AND (
      NOT sqlc.arg(has_parlour_set)::boolean
      OR put.has_parlour = sqlc.arg(has_parlour)::boolean
    )
  GROUP BY p.id, put.id
),
filtered AS (
  SELECT *
  FROM opportunities
  WHERE (
      sqlc.arg(availability)::text = 'all'
      OR available_offer_count > 0
    )
    AND (
      NOT sqlc.arg(min_price_set)::boolean
      OR lowest_price_kobo >= sqlc.arg(min_price_kobo)::integer
    )
    AND (
      NOT sqlc.arg(max_price_set)::boolean
      OR lowest_price_kobo <= sqlc.arg(max_price_kobo)::integer
    )
)
SELECT
  property_id,
  property_name,
  property_area,
  property_landmark,
  unit_type_id,
  category,
  unit_type_name,
  description,
  notes,
  bedroom_count,
  has_parlour,
  bathroom_type,
  kitchen_type,
  COALESCE(lowest_price_kobo, 0)::integer AS lowest_price_kobo,
  available_offer_count,
  COALESCE(thumbnail_url, '')::text AS thumbnail_url,
  created_at,
  updated_at::timestamptz AS updated_at,
  COUNT(*) OVER() AS total_count
FROM filtered
ORDER BY
  CASE WHEN sqlc.arg(sort)::text = 'recommended' THEN (available_offer_count > 0)::integer END DESC,
  CASE WHEN sqlc.arg(sort)::text = 'recommended' THEN (thumbnail_url IS NOT NULL)::integer END DESC,
  CASE WHEN sqlc.arg(sort)::text = 'recommended' THEN completeness_score END DESC,
  CASE WHEN sqlc.arg(sort)::text = 'recommended' THEN updated_at END DESC,
  CASE WHEN sqlc.arg(sort)::text = 'recommended' THEN available_offer_count END DESC,
  CASE WHEN sqlc.arg(sort)::text = 'created_at' THEN created_at END,
  CASE WHEN sqlc.arg(sort)::text = '-created_at' THEN created_at END DESC,
  CASE WHEN sqlc.arg(sort)::text = 'lowest_price_kobo' THEN lowest_price_kobo END NULLS LAST,
  CASE WHEN sqlc.arg(sort)::text = '-lowest_price_kobo' THEN lowest_price_kobo END DESC NULLS LAST,
  unit_type_id
LIMIT sqlc.arg(result_limit)::integer
OFFSET sqlc.arg(result_offset)::integer;
