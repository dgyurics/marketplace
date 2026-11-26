<template>
  <section class="no-methods">
    <p class="no-methods-title">No payment methods available</p>
    <p class="no-methods-description">Order cannot be placed</p>
  </section>
</template>

<style scoped>
.no-methods {
  margin-top: 20px;
  margin-bottom: 12px;
  border: 1px solid var(--border-color);
  background: var(--subtle-background);
  border-radius: var(--panel-radius);
  padding: 16px 18px;
  max-width: 560px;
  margin-left: auto;
  margin-right: auto;
  text-align: center;
}

.no-methods-title {
  margin: 0;
  font-weight: 600;
  color: var(--text-color);
}

.no-methods-title::after {
  content: '';
  display: block;
  width: 40px;
  height: 1px;
  margin: 12px auto;
  background: var(--border-color);
}

.no-methods-description {
  margin: 0;
  color: var(--muted-text-color);
  line-height: 1.45;
}
</style>
