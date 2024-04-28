<p align="center">
  <img src="https://raw.githubusercontent.com/dgyurics/marketplace/main/logo.webp" alt="marketplace">
</p>
<ul>
  <li>
    <strong>Product Endpoints</strong>
    <ul>
      <li><strong>GET /products</strong> - Fetch a list of all products</li>
      <li><strong>GET /products/{id}</strong> - Fetch details of a single product by its ID</li>
      <li><strong>POST /products</strong> - Create a new product (admin only)</li>
      <li><strong>PUT /products/{id}</strong> - Update an existing product by its ID (admin only)</li>
      <li><strong>DELETE /products/{id}</strong> - Delete a product by its ID (admin only)</li>
      <li><strong>GET /products/search</strong> - Search for products by name, description, or other attributes</li>
      <li><strong>PATCH /products/{id}</strong> - Partially update product details (admin only)</li>
    </ul>
  </li>
  <li>
    <strong>Category Endpoints</strong>
    <ul>
      <li><strong>GET /categories</strong> - Fetch a list of all product categories</li>
      <li><strong>GET /categories/{id}</strong> - Fetch details of a single category by its ID</li>
      <li><strong>POST /categories</strong> - Create a new category (admin only)</li>
      <li><strong>PUT /categories/{id}</strong> - Update an existing category by its ID (admin only)</li>
      <li><strong>DELETE /categories/{id}</strong> - Delete a category by its ID (admin only)</li>
      <li><strong>GET /categories/{id}/products</strong> - Fetch all products in a specific category</li>
    </ul>
  </li>
  <li>
    <strong>Cart Endpoints</strong>
    <ul>
      <li><strong>GET /cart</strong> - Fetch the current user's cart</li>
      <li><strong>POST /cart</strong> - Add an item to the cart</li>
      <li><strong>PUT /cart/{itemId}</strong> - Update the quantity of a cart item</li>
      <li><strong>DELETE /cart/{itemId}</strong> - Remove an item from the cart</li>
      <li><strong>PATCH /cart/{itemId}</strong> - Partially update cart item details</li>
      <li><strong>POST /cart/checkout</strong> - Initiate the checkout process</li>
      <li><strong>GET /cart/total</strong> - Fetch the total price of the cart</li>
      <li><strong>POST /cart/clear</strong> - Clear all items from the cart</li>
    </ul>
  </li>
  <li>
    <strong>Order Endpoints</strong>
    <ul>
      <li><strong>POST /orders</strong> - Place a new order</li>
      <li><strong>GET /orders/{id}</strong> - Fetch details of a specific order</li>
      <li><strong>GET /orders</strong> - Fetch all orders for the authenticated user</li>
      <li><strong>GET /orders</strong> - Fetch all orders (admin only)</li>
      <li><strong>PUT /orders/{id}</strong> - Update the status of an order (admin only)</li>
      <li><strong>DELETE /orders/{id}</strong> - Cancel an order (admin only)</li>
    </ul>
  </li>
  <li>
    <strong>Authentication Endpoints</strong>
    <ul>
      <li><strong>POST auth/register</strong> - Create a new user account</li>
      <li><strong>POST auth/login</strong> - Authenticate user and generate JWT</li>
      <li><strong>POST auth/refresh-token</strong> - Generate new JWT using refresh token</li>
      <li><strong>POST auth/logout</strong> - Invalidate refresh token</li>
      <li><strong>GET auth/profile</strong> - Fetch authenticated user's profile</li>
      <li><strong>POST auth/update-profile</strong> - Update the authenticated user's profile information</li>
      <li><strong>POST auth/change-password</strong> - Change user's password</li>
      <li><strong>POST auth/forgot-password</strong> - Initiate password reset process</li>
      <li><strong>POST auth/reset-password</strong> - Reset user's password using a token</li>
      <li><strong>GET auth/users</strong> (admin only) - Fetch a list of all users</li>
      <li><strong>DELETE auth/users/{id}</strong> (admin only) - Delete a user account by ID</li>
    </ul>    
  </li>
</ul>
