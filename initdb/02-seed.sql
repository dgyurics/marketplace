INSERT INTO categories (id, name, description) VALUES
('3d6f0c4a-75bf-4b9b-9f12-003d6f2f9a1f', 'Category 1', 'This is category 1'),
('81d29ba6-ff4c-4b48-93be-295f31864d5b', 'Category 2', 'This is category 2'),
('4b71dc4e-05e1-4b19-8307-d3dff67dc11f', 'Category 3', 'This is category 3'),
('7ae54a1e-4a4e-40e8-bb0f-c3096d41891f', 'Category 4', 'This is category 4'),
('58c9aaf6-490f-49b6-8c89-64cb7c5e31e3', 'Category 5', 'This is category 5');

INSERT INTO products (id, name, price, description) VALUES
('1c2d6b57-5e1b-4f29-bb38-dbb4b065e5e8', 'Product 1', 1000, 'This is product 1'),
('2a5d7f08-4d2b-4b0a-b8b7-e1bc9f01d898', 'Product 2', 2000, 'This is product 2'),
('3f9b8b1a-9d7c-4b37-8d35-ffa7e00c2a54', 'Product 3', 3000, 'This is product 3'),
('4c6b7d2b-6f0a-4e1d-8a5c-7a4e6f0c9f1b', 'Product 4', 4000, 'This is product 4'),
('5d7a8c1c-7e2c-4f1b-9a6d-8b5f7d0c1a2e', 'Product 5', 5032, 'This is product 5'),
('6e8b9d2d-8f3d-4f2c-a7b8-9c6f8e1d2b3f', 'Product 6', 6000, 'This is product 6'),
('7f9a0e3e-9d4e-4f3d-b8c9-a7e0f9f3c4d5', 'Product 7', 7077, 'This is product 7'),
('8a1b2c4f-ad5f-4b4e-c9d0-b8f1a0f4d6e6', 'Product 8', 8000, 'This is product 8'),
('9b2c3d56-be66-4c5f-da1e-c9e2f0f5e7f7', 'Product 9', 9099, 'This is product 9'),
('af3d4e6a-ce7a-4d6b-eb2f-d0f3a1b6c8a8', 'Product 10', 10000, 'This is product 10');

INSERT INTO inventory (product_id, quantity) VALUES
('1c2d6b57-5e1b-4f29-bb38-dbb4b065e5e8', 100),
('2a5d7f08-4d2b-4b0a-b8b7-e1bc9f01d898', 200),
('3f9b8b1a-9d7c-4b37-8d35-ffa7e00c2a54', 150),
('4c6b7d2b-6f0a-4e1d-8a5c-7a4e6f0c9f1b', 120),
('5d7a8c1c-7e2c-4f1b-9a6d-8b5f7d0c1a2e', 0),
('6e8b9d2d-8f3d-4f2c-a7b8-9c6f8e1d2b3f', 0),
('7f9a0e3e-9d4e-4f3d-b8c9-a7e0f9f3c4d5', 90),
('8a1b2c4f-ad5f-4b4e-c9d0-b8f1a0f4d6e6', 70),
('9b2c3d56-be66-4c5f-da1e-c9e2f0f5e7f7', 110),
('af3d4e6a-ce7a-4d6b-eb2f-d0f3a1b6c8a8', 80);

INSERT INTO product_categories (product_id, category_id) VALUES 
('1c2d6b57-5e1b-4f29-bb38-dbb4b065e5e8', '3d6f0c4a-75bf-4b9b-9f12-003d6f2f9a1f'),
('2a5d7f08-4d2b-4b0a-b8b7-e1bc9f01d898', '3d6f0c4a-75bf-4b9b-9f12-003d6f2f9a1f'),
('3f9b8b1a-9d7c-4b37-8d35-ffa7e00c2a54', '3d6f0c4a-75bf-4b9b-9f12-003d6f2f9a1f'),
('4c6b7d2b-6f0a-4e1d-8a5c-7a4e6f0c9f1b', '81d29ba6-ff4c-4b48-93be-295f31864d5b'),
('5d7a8c1c-7e2c-4f1b-9a6d-8b5f7d0c1a2e', '81d29ba6-ff4c-4b48-93be-295f31864d5b'),
('6e8b9d2d-8f3d-4f2c-a7b8-9c6f8e1d2b3f', '81d29ba6-ff4c-4b48-93be-295f31864d5b'),
('7f9a0e3e-9d4e-4f3d-b8c9-a7e0f9f3c4d5', '4b71dc4e-05e1-4b19-8307-d3dff67dc11f'),
('8a1b2c4f-ad5f-4b4e-c9d0-b8f1a0f4d6e6', '4b71dc4e-05e1-4b19-8307-d3dff67dc11f'),
('9b2c3d56-be66-4c5f-da1e-c9e2f0f5e7f7', '7ae54a1e-4a4e-40e8-bb0f-c3096d41891f'),
('af3d4e6a-ce7a-4d6b-eb2f-d0f3a1b6c8a8', '58c9aaf6-490f-49b6-8c89-64cb7c5e31e3');
