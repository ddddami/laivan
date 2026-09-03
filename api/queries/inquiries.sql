-- name: FindInquiryByStudentSubmission :one
SELECT id, student_user_id, agent_offer_id, submission_id, message, status, created_at, updated_at
FROM inquiries
WHERE student_user_id = $1 AND submission_id = $2;

-- name: CreateInquiryForAvailableOffer :one
INSERT INTO inquiries (student_user_id, agent_offer_id, submission_id, message)
SELECT sqlc.arg(student_user_id), ao.id, sqlc.arg(submission_id), btrim(sqlc.arg(message))
FROM agent_offers ao
JOIN property_unit_types put ON put.id = ao.property_unit_type_id
JOIN properties p ON p.id = put.property_id
JOIN agents a ON a.id = ao.agent_id
WHERE ao.id = sqlc.arg(agent_offer_id)
  AND ao.status = 'available'
  AND ao.archived_at IS NULL
  AND a.status = 'active'
  AND a.user_id <> sqlc.arg(student_user_id)
  AND EXISTS (
      SELECT 1
      FROM agent_campuses ac
      WHERE ac.agent_id = a.id AND ac.campus_id = p.campus_id
  )
RETURNING id, student_user_id, agent_offer_id, submission_id, message, status, created_at, updated_at;

-- name: GetInquiryOfferState :one
SELECT ao.archived_at, a.user_id AS agent_user_id
FROM agent_offers ao
JOIN agents a ON a.id = ao.agent_id
WHERE ao.id = $1;

-- name: GetInquiryHandoff :one
SELECT a.display_name AS agent_display_name,
       a.phone_number,
       a.whatsapp_number,
       ao.title AS offer_title,
       p.name AS property_name,
       put.name AS unit_name,
       p.area
FROM inquiries i
JOIN agent_offers ao ON ao.id = i.agent_offer_id
JOIN agents a ON a.id = ao.agent_id
JOIN property_unit_types put ON put.id = ao.property_unit_type_id
JOIN properties p ON p.id = put.property_id
WHERE i.id = $1;
