<p align="center">
  <img src="https://raw.githubusercontent.com/dgyurics/marketplace/main/logo.webp" alt="marketplace">
</p>

<h2>Marketplace</h2>
<p align="center">
Marketplace is a self-hosted e-commerce framework, designed for local businesses and community commerce. Built with minimal external dependencies and maximum self-reliance.
</p>

<h2>Features</h2>
<ul>
  <li>Product Management: CRUD operations for products, including categorization, basic inventory management, and support for product meta-data (e.g., material, weight, dimensions).</li>
  <li>Order Management: Create and manage orders, including order status tracking.</li>
  <li>Guest Checkout: Allow purchases without account creation.</li>
  <li>User Authentication: JWT-based user authorization and authentication.</li>
  <li>Shopping Cart: Manage customer shopping carts, calculate totals, and track item quantities.</li>
  <li>Payment Processing: Currently supports Stripe integration (being modified to support self-hosted alternatives).</li>
  <li>User Interface: An extensible user interface.</li>
  <li>Logging: Structured logging for monitoring and auditing.</li>
  <li>Email Notifications: Send order status emails to customers (currently via Mailjet, transitioning to self-hosted SMTP).</li>
  <li>Image Hosting: Self-hosted images using imgproxy. Automatic image resizing, compression, and format conversion.</li>
  <li>Internationalization and Localization: Support for multiple currencies and languages.</li>
</ul>

<h2>Active Development</h2>
<ul>
  <li>Self-Hosted SMTP:</string> Replace Mailjet with a self-hosted SMTP server.</li>
  <li>Native Payment Processing: Eliminate Stripe dependency.</li>
  <li>Admin Interface: An extensible admin interface for managing products, orders, and users.</li>
  <li>Backup and Recovery: Automated self-hosted backup solution.</li>
  <li>Caching: Implement API caching to improve performance.</li>
  <li>Rate Limiting: Protect the API from abuse with rate limiting.</li>
</ul>

<h2>Planned Enhancements</h2>
<ul>
  <li>Search: Search capabilities, allowing customer to find products through keyword and fuzzy searches.</li>
  <li>Refunds: Support for partial and full refunds.</li>
  <li>Product Variants: Support product variants (e.g., size, color) and manage inventory for each variant (SKU).</li>
</ul>

<h2>Future Considerations</h2>
<ul>
  <li>Local Delivery Zones: Define delivery areas and coordinate local delivery.</li>
  <li>Pickup Options: Allow customers to choose when to pick-up.</li>
  <li>Promotions and Discounts: Manage promotional codes.</li>
</ul>

<h2>Prerequisites</h2>
<ul>
  <li>Go 1.22 or higher</li>
  <li>Node.js 20.17 or higher</li>
  <li>Docker</li>
</ul>
