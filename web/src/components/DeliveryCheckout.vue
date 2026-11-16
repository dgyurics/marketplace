<template>
  <form @submit.prevent="submit">
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
        <input v-model="acknowledged" type="checkbox" />
        <span>
          I understand payment is collected on delivery and that refusing the package may limit
          future pay-on-delivery orders.
        </span>
      </label>
    </section>

    <button
      type="submit"
      class="btn-full-width mt-30"
      :disabled="isSubmitting || !acknowledged"
      :tabindex="0"
    >
      Place Order
    </button>
  </form>
</template>

<script setup lang="ts">
import { useDeliveryCheckout } from '@/composables/useDeliveryCheckout'

const { acknowledged, isSubmitting, submit } = useDeliveryCheckout()
</script>

<style scoped>
.delivery-info {
  font-size: 14px;
  margin-top: 20px;
  margin-bottom: 12px;
  border: 1px solid var(--border-color);
  background: var(--subtle-background);
  border-radius: var(--panel-radius);
  padding: 16px 18px;
  margin-left: auto;
  margin-right: auto;
}

.delivery-info-description {
  margin: 0;
  color: var(--muted-text-color);
  line-height: 1.45;
}

.delivery-steps {
  margin: 12px 0 0;
  padding-left: 18px;
  color: var(--muted-text-color);
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
  color: var(--muted-text-color);
  cursor: pointer;
}

.delivery-ack input {
  margin-top: 2px;
}
</style>
