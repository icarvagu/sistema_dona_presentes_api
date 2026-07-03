-- +goose Up
UPDATE purchase_orders po
SET general_number = po.sale_id::text,
    has_engraving = EXISTS (
        SELECT 1
        FROM sale_items si
        WHERE si.sale_id = po.sale_id
          AND (COALESCE(si.engravings, '[]'::jsonb) <> '[]'::jsonb OR COALESCE(si.personalization_type, '') <> '')
    ),
    corel_required = EXISTS (
        SELECT 1
        FROM sale_items si
        WHERE si.sale_id = po.sale_id
          AND (COALESCE(si.engravings, '[]'::jsonb) <> '[]'::jsonb OR COALESCE(si.personalization_type, '') <> '')
    ),
    updated_at = NOW();

-- +goose Down
-- A numeração compartilhada entre Venda e Compra é definitiva e não deve ser revertida.
