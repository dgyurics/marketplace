-- Change image sort order (ascending)
CREATE OR REPLACE VIEW v_products AS
SELECT
    p.id,
    p.name,
    p.price,
    p.summary,
    COALESCE(p.description, '') AS description,
    p.details,
    p.category_id,
    p.inventory,
    p.cart_limit,
    COALESCE(p.tax_code, '') AS tax_code,
    p.featured,
    p.pickup_only,
    c.slug AS category_slug,
    COALESCE(imgs.images, '[]') AS images,
    COALESCE(order_stats.total_sold, 0) AS total_sold,
    p.created_at
FROM products p
LEFT JOIN categories c ON p.category_id = c.id
LEFT JOIN LATERAL (
    SELECT JSONB_AGG(
        JSONB_BUILD_OBJECT(
            'id', i.id::TEXT,
            'url', i.url,
            'type', i.type,
            'updated_at', i.updated_at,
            'alt_text', i.alt_text
        ) ORDER BY i.id ASC
    ) AS images
    FROM images i
    WHERE i.product_id = p.id
) imgs ON TRUE
LEFT JOIN (
    SELECT product_id, sum(oi.quantity) AS total_sold
    FROM order_items oi
    GROUP BY product_id
) order_stats ON order_stats.product_id = p.id
WHERE p.is_deleted = FALSE;