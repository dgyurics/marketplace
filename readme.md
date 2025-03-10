<p align="center">
  <img src="https://raw.githubusercontent.com/dgyurics/marketplace/main/logo.webp" alt="marketplace">
</p>
<h1 align="center">Marketplace</h1>
<p align="center">
  An e-commerce API built with Go, designed to handle essential operations such as product and order management, user authentication, and shopping carts. A headless backend aims to serve as a robust foundation for any e-commerce application. Dependencies have been kept to a minimum.
</p>
<h2>Features</h2>
<ul>
  <li><strong>Product Management:</strong> CRUD operations for products, including categorization and inventory management.</li>
  <li><strong>User Authentication:</strong> JWT-based user authentication, supporting registration and login functionalities.</li>
  <li><strong>Shopping Cart:</strong> Manage user shopping carts, calculate totals, and track item quantities.</li>
  <li><strong>Logging:</strong> Structured logging for monitoring and debugging.</li>
  <li><strong>Order Management:</strong> Create and manage orders, including order status tracking.</li>
  <li><strong>Payments:</strong> Stripe integration for processing payments.</li>
</ul>

<h2>Planned Enhancements</h2>
<ul>
  <li><strong>Email Notifications:</strong> Send order status emails to users.</li>
  <li><strong>Search and Filtering:</strong> Advanced search capabilities, allowing users to find products through keyword and fuzzy searches.</li>
  <li><strong>Caching:</strong> Implement API caching to improve performance.</li>
  <li><strong>Rate Limiting:</strong> Protect the API from abuse with rate limiting.</li>
  <li><strong>Horizontal Database Scaling:</strong> Implement sharding and replication to scale the database horizontally.</li>
  <li><strong>Cart Item Limits:</strong> Implement cart item limts to prevent abuse.</li>
</ul>

<h2>Future Considerations</h2>
<ul>
  <li><strong>User Interface:</strong> Develop a user interface to interact with the API.</li>
  <li><strong>Shipping Management:</strong> Calculate shipping costs, manage shipping providers, and track shipments.</li>
  <li><strong>Role-Based Access Control:</strong> Implement role-based access control to restrict access to certain resources. E.g. admin, vendor, customer</li>
  <li><strong>Promotions and Discounts:</strong> Manage promotional codes, discounts, and sales events.</li>
  <li><strong>Inventory Management:</strong> Advanced inventory management with alerts for low stock and support for multiple warehouses.</li>
  <li><strong>Internationalization and Localization:</strong> support multiple currencies, languages, and localized product information to accommodate different regions.</li>
  <li><strong>Product Reviews:</strong> Allow users to leave reviews and ratings for products.</li>
</ul>

<h2>Prerequisites</h2>
<ul>
  <li>Go 1.22 or higher</li>
  <li>Docker and Docker Compose</li>
</ul>
