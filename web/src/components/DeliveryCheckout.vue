<template>
  <div>
    <section class="delivery-info">
      <p class="delivery-info-description">
        Nothing is charged now. Payment is collected when the order arrives.
      </p>

      <ul class="delivery-steps">
        <li>We reserve your items and dispatch the order.</li>
        <li>You inspect the package at the door.</li>
        <li>You pay the courier by card or cash.</li>
      </ul>

      <label class="delivery-ack">
        <input type="checkbox" />
        <span>
          I understand payment is collected on delivery and that refusing the package may limit
          future pay-on-delivery orders.
        </span>
      </label>
    </section>

    <button type="button" class="btn-full-width mt-30" disabled :tabindex="0">Place Order</button>
  </div>
</template>

<style scoped>
.delivery-info {
  margin-top: 20px;
  margin-bottom: 12px;
  border: 1px solid #e9e9e9;
  background: #fafafa;
  border-radius: 12px;
  padding: 16px 18px;
  max-width: 560px;
  margin-left: auto;
  margin-right: auto;
}

.delivery-info-description {
  margin: 0;
  color: #444;
  line-height: 1.45;
}

.delivery-steps {
  margin: 12px 0 0;
  padding-left: 18px;
  color: #444;
  font-size: 13px;
  line-height: 1.6;
}

.delivery-ack {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-top: 14px;
  font-size: 12px;
  line-height: 1.5;
  color: #555;
  cursor: pointer;
}

.delivery-ack input {
  margin-top: 2px;
}
</style>
