<p align="center">
  <img src="https://raw.githubusercontent.com/dgyurics/marketplace/main/logo.webp" alt="marketplace">
</p>
<h1 align="center">Marketplace</h1>
<p align="center">
  An e-commerce backend API built with Go, designed to handle essential operations such as product management, user authentication, and shopping carts. The project follows a clean architecture approach, making it easy to extend and maintain. This backend aims to serve as a robust foundation for any e-commerce application. Dependencies have been kept to a minimum.
</p>
<h2>Features</h2>
<ul>
  <li><strong>Product Management:</strong> CRUD operations for products, including categorization and inventory management.</li>
  <li><strong>User Authentication:</strong> JWT-based user authentication, supporting registration and login functionalities.</li>
  <li><strong>Shopping Cart:</strong> Manage user shopping carts, calculate totals, and track item quantities.</li>
  <li><strong>Transaction Handling:</strong> Ensure data consistency with transaction support for complex operations.</li>
  <li><strong>Repository Pattern:</strong> Abstraction layer for database operations, improving testability and flexibility.</li>
</ul>

<h2>Planned Enhancements</h2>
<ul>
  <li><strong>Order Management:</strong> Implement order creation and tracking, with status updates and order history.</li>
  <li><strong>Payment Integration:</strong> Integrate with a single payment prover, Stripe.</li>
  <li><strong>Email Notifications:</strong> Send order confirmations to users.</li>
  <li><strong>Search and Filtering:</strong> Advanced search capabilities, allowing users to find products through keyword and fuzzy searches.</li>
</ul>

<h2>Future Considerations</h2>
<ul>
  <li><strong>Shipping Management:</strong> Calculate shipping costs, manage shipping providers, and track shipments.</li>
  <li><strong>Promotions and Discounts:</strong> Manage promotional codes, discounts, and sales events.</li>
  <li><strong>Inventory Management:</strong> Advanced inventory management with alerts for low stock and support for multiple warehouses.</li>
  <li><strong>Internationalization and Localization:</strong> Plan for potential support of multiple currencies, languages, and localized product information to accommodate different countries and regions.</li>
  <li><strong>Multiple Payment Providers:</strong> Support multiple payment providers, such as PayPal and Apple Pay.</li>
</ul>

<h2>Prerequisites</h2>
<ul>
  <li>Go 1.20 or higher</li>
  <li>Docker and Docker Compose</li>
</ul>
