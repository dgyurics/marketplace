<template>
  <div class="error">
    <h1>{{ $route.query['status'] || '500' }}</h1>
    <p>Something went wrong</p>
  </div>
</template>

<!-- TODO notify service of error -->

<style scoped>
.error {
  text-align: center;
  padding: 4rem 2rem;
  color: #555;
}

h1 {
  font-size: 2rem;
  margin-bottom: 1rem;
}
</style>
