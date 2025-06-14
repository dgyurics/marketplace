<p align="center">
  <img src="https://raw.githubusercontent.com/dgyurics/marketplace/main/logo.webp" alt="marketplace">
</p>
<h1 align="center">Marketplace</h1>
<p align="center">
  Marketplace is a self-hosted e-commerce framework built with Go, designed to help developers build and customize their own online store. Dependencies have been kept to a minimum.
</p>
<h2>Features</h2>
<ul>
  <li><strong>Product Management:</strong> CRUD operations for products, including categorization, basic inventory management, and support for product meta-data (e.g., material, weight, dimensions).</li>
  <li><strong>Order Management:</strong> Create and manage orders, including order status tracking.</li>
  <li><strong>User Authentication:</strong> JWT-based user authorization and authentication.</li>
  <li><strong>Shopping Cart:</strong> Manage user shopping carts, calculate totals, and track item quantities.</li>
  <li><strong>Payments:</strong> Stripe integration for calculating tax and processing payments.</li>
  <li><strong>Logging:</strong> Structured logging for monitoring and debugging in a distributed environment.</li>
  <li><strong>Internationalization and Localization:</strong> Support for multiple currencies and languages.</li>
  <li><strong>Email Notifications:</strong> Send order status emails to users.</li>
</ul>

<h2>Planned Enhancements</h2>
<ul>
  <li><strong>Search:</strong> Advanced search capabilities, allowing users to find products through keyword and fuzzy searches.</li>
  <li><strong>Caching:</strong> Implement API caching to improve performance.</li>
  <li><strong>Rate Limiting:</strong> Protect the API from abuse with rate limiting.</li>
  <li><strong>Image Hosting:</strong> Option to self-host product images (currently using Cloudinary).</li>
  <li><strong>SMTP:</strong> Self-host SMTP server (currently using Mailjet).</li>
  <li><strong>Product Variants:</strong> Support product variants (e.g., size, color) and manage inventory for each variant (SKU).</li>
  <li><strong>User Interface:</strong> An extensible user interface to be deployed alongside the API.</li>
  <li><strong>Refunds:</strong> Administrative endpoint for partial and full refunds.</li>
  <li><strong>Admin Interface:</strong> An extensible admin interface to manage products, inventory, and orders.</li>
</ul>

<h2>Future Considerations</h2>
<ul>
  <li><strong>Extensible payment provider:</strong> Stripe by default, others pluggable.</li>
  <li><strong>Shipping Management:</strong> Calculate shipping costs and track shipments.</li>
  <li><strong>Role-Based Access Control:</strong> Implement role-based access control to restrict access to certain resources. E.g. admin, vendor, customer</li>
  <li><strong>Promotions and Discounts:</strong> Manage promotional codes, discounts, and sale events.</li>
  <li><strong>Advanced Inventory Management:</strong> Inventory management with alerts for low stock, as well as support for multiple distributors.</li>
  <li><strong>Product Reviews:</strong> Allow users to leave reviews and ratings for products.</li>
  <li><strong>Pay with Crypto:</strong> Support for payments with cryptocurrency would eliminate need for 3rd party payment processors.</li>
</ul>

<h2>Prerequisites</h2>
<ul>
  <li>Go 1.22 or higher</li>
  <li>Docker</li>
</ul>

<h2>Installation</h2>
<p><a href="https://github.com/dgyurics/marketplace/wiki">This guide</a> walks you through setting up Marketplace on a machine running Ubuntu Linux. Although installing Marketplace is relatively straightforward, we recommend working knowledge of Go if you plan to modify the source code.
</p>
