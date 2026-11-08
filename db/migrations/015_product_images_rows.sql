-- Product images are loaded directly from images by the repositories.
DROP VIEW v_products;

CREATE VIEW v_products AS
SELECT
	p.id,
	p.name,
	p.price,
	COALESCE(p.summary, '') AS summary,
	COALESCE(p.description, '') AS description,
	p.details,
	p.category_id,
	p.inventory,
	p.cart_limit,
	COALESCE(p.tax_code, '') AS tax_code,
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
	GROUP BY product_id
) order_stats ON order_stats.product_id = p.id
WHERE p.status <> 'archived';
