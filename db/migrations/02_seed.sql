INSERT INTO categories (id, name, description) VALUES
('526402777421709313', 'Category 1', 'category-1', 'This is category 1'),
('526403107899310082', 'Category 2', 'category-2', 'This is category 2'),
('526403265579974659', 'Category 3', 'category-3', 'This is category 3'),
('526403439761031172', 'Category 4', 'category-4', 'This is category 4'),
('526403643973304325', 'Category 5', 'category-5', 'This is category 5');

INSERT INTO products (id, name, price, description) VALUES
('526403779902308358', 'Product 1', 1000, 'This is product 1'),
('526403958856482823', 'Product 2', 2000, 'This is product 2'),
('526404087537729544', 'Product 3', 3000, 'This is product 3'),
('526404262373097481', 'Product 4', 4000, 'This is product 4'),
('526404379536785418', 'Product 5', 5032, 'This is product 5'),
('526404490752950283', 'Product 6', 6000, 'This is product 6'),
('526404634617577484', 'Product 7', 7077, 'This is product 7'),
('526404777643343885', 'Product 8', 8000, 'This is product 8'),
('526404888909840398', 'Product 9', 9099, 'This is product 9'),
('526404991661899791', 'Product 10', 10000, 'This is product 10');

INSERT INTO images (id, product_id, image_url, image_type, format, animated, display_order, alt_text) VALUES
('576264336620650498', '526403779902308358', 'https://picsum.photos/seed/product1-main/800/600', 'main', 'avif', false, 0, 'Main image of Product 1'),
('576264336620650499', '526403779902308358', 'https://picsum.photos/seed/product1-thumb/150/150', 'thumbnail', 'avif', false, 1, 'Thumbnail image of Product 1'),
('576264336620650500', '526403779902308358', 'https://picsum.photos/seed/product1-gallery1/800/600', 'gallery', 'avif', false, 2, 'Gallery image of Product 1'),
('576264336620650501', '526403958856482823', 'https://picsum.photos/seed/product2-main/800/600', 'main', 'avif', false, 0, 'Main image of Product 2'),
('576264336620650502', '526403958856482823', 'https://picsum.photos/seed/product2-thumb/150/150', 'thumbnail', 'avif', false, 1, 'Thumbnail image of Product 2'),
('576264336620650503', '526404087537729544', 'https://picsum.photos/seed/product3-main/800/600', 'main', 'avif', false, 0, 'Main image of Product 3'),
('576264336620650504', '526404087537729544', 'https://picsum.photos/seed/product3-thumb/150/150', 'thumbnail', 'avif', false, 1, 'Thumbnail image of Product 3'),
('576264336620650505', '526404087537729544', 'https://picsum.photos/seed/product3-gallery1/800/600', 'gallery', 'avif', false, 2, 'Gallery image of Product 3'),
('576264336620650506', '526404262373097481', 'https://picsum.photos/seed/product4-main/800/600', 'main', 'avif', false, 0, 'Main image of Product 4'),
('576264336620650507', '526404262373097481', 'https://picsum.photos/seed/product4-thumb/150/150', 'thumbnail', 'avif', false, 1, 'Thumbnail image of Product 4'),
('576264336620650508', '526404262373097481', 'https://picsum.photos/seed/product4-gallery1/800/600', 'gallery', 'avif', false, 2, 'Gallery image of Product 4'),
('576264336620650509', '526404379536785418', 'https://picsum.photos/seed/product5-main/800/600', 'main', 'avif', false, 0, 'Main image of Product 5'),
('576264336620650510', '526404379536785418', 'https://picsum.photos/seed/product5-thumb/150/150', 'thumbnail', 'avif', false, 1, 'Thumbnail image of Product 5'),
('576264336620650511', '526404379536785418', 'https://picsum.photos/seed/product5-zoom/1600/1200', 'zoom', 'avif', false, 2, 'Zoom image of Product 5'),
('576264336620650512', '526404490752950283', 'https://picsum.photos/seed/product6-main/800/600', 'main', 'avif', false, 0, 'Main image of Product 6'),
('576264336620650513', '526404490752950283', 'https://picsum.photos/seed/product6-thumb/150/150', 'thumbnail', 'avif', false, 1, 'Thumbnail image of Product 6'),
('576264336620650514', '526404634617577484', 'https://picsum.photos/seed/product7-main/800/600', 'main', 'avif', false, 0, 'Main image of Product 7'),
('576264336620650515', '526404634617577484', 'https://picsum.photos/seed/product7-thumb/150/150', 'thumbnail', 'avif', false, 1, 'Thumbnail image of Product 7'),
('576264336620650516', '526404634617577484', 'https://picsum.photos/seed/product7-hero/1920/1080', 'hero', 'avif', false, 2, 'Hero image of Product 7'),
('576264336620650517', '526404777643343885', 'https://picsum.photos/seed/product8-main/800/600', 'main', 'avif', false, 0, 'Main image of Product 8'),
('576264336620650518', '526404777643343885', 'https://picsum.photos/seed/product8-thumb/150/150', 'thumbnail', 'avif', false, 1, 'Thumbnail image of Product 8'),
('576264336620650519', '526404888909840398', 'https://picsum.photos/seed/product9-main/800/600', 'main', 'avif', false, 0, 'Main image of Product 9'),
('576264336620650520', '526404888909840398', 'https://picsum.photos/seed/product9-thumb/150/150', 'thumbnail', 'avif', false, 1, 'Thumbnail image of Product 9'),
('576264336620650521', '526404888909840398', 'https://picsum.photos/seed/product9-zoom/1600/1200', 'zoom', 'avif', false, 2, 'Zoom image of Product 9'),
('576264336620650522', '526404991661899791', 'https://picsum.photos/seed/product10-main/800/600', 'main', 'avif', false, 0, 'Main image of Product 10'),
('576264336620650523', '526404991661899791', 'https://picsum.photos/seed/product10-thumb/150/150', 'thumbnail', 'avif', false, 1, 'Thumbnail image of Product 10'),
('576264336620650524', '526404991661899791', 'https://fakegifhost.com/product10.gif', 'gallery', 'gif', true, 2, 'Animated GIF for Product 10');

INSERT INTO inventory (product_id, quantity) VALUES
('526403779902308358', 100),
('526403958856482823', 200),
('526404087537729544', 150),
('526404262373097481', 120),
('526404379536785418', 0),
('526404490752950283', 0),
('526404634617577484', 90),
('526404777643343885', 70),
('526404888909840398', 110),
('526404991661899791', 80);

INSERT INTO product_categories (product_id, category_id) VALUES 
('526403779902308358', '526402777421709313'),
('526403958856482823', '526402777421709313'),
('526404087537729544', '526402777421709313'),
('526404262373097481', '526403107899310082'),
('526404379536785418', '526403107899310082'),
('526404490752950283', '526403107899310082'),
('526404634617577484', '526403265579974659'),
('526404777643343885', '526403265579974659'),
('526404888909840398', '526403439761031172'),
('526404991661899791', '526403643973304325');
