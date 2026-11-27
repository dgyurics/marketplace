ALTER TABLE products
ADD COLUMN IF NOT EXISTS member_price BIGINT;

DROP VIEW v_products;

CREATE OR REPLACE VIEW v_products AS
SELECT
    p.id,
    p.name,
    p.price,
    p.member_price,
    p.summary,
    p.description,
    p.details,
    p.category_id,
    p.inventory,
    p.cart_limit,
    p.tax_code,
    p.featured,
    p.negotiable,
    p.sort_order,
    c.slug AS category_slug,
    COALESCE(order_stats.total_sold, 0) AS total_sold,
    p.created_at,
    p.status
FROM products p
LEFT JOIN categories c ON p.category_id = c.id
LEFT JOIN (
    SELECT product_id, sum(oi.quantity) AS total_sold
    FROM order_items oi
    JOIN orders o ON o.id = oi.order_id
    WHERE o.status NOT IN ('requested', 'canceled', 'refunded')
    GROUP BY product_id
) order_stats ON order_stats.product_id = p.id
WHERE p.status <> 'archived';
