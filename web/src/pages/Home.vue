<template>
  <div class="home-container">
    <div class="feature-container"></div>
  </div>
</template>

<style scoped>
.home-container {
  position: relative;
  width: 100%;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  z-index: 1;
}

.feature-container {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  z-index: -1;
  overflow: hidden;
  background-image: url('@/assets/roberto-nickson-bg.webp');
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
}

.feature-image {
  width: 100vw;
  height: 100vh;
  object-fit: cover;
  position: absolute;
  top: 0;
  left: 0;
}
</style>
