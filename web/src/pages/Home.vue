<template>
  <div class="home-container">
    <div class="feature-container">
      <div class="hero-content unselectable">
        <h1 class="hero-title">essential living</h1>
        <div class="hero-buttons">
          <button class="btn-sm" :tabindex="0" @click="$router.push('/new')">shop new</button>
          <button class="btn-sm" :tabindex="0" @click="$router.push('/popular')">
            shop popular
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.home-container {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  background-color: #000;
  overflow: hidden;
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
  display: flex;
  align-items: center;
  justify-content: center;
}

.feature-image {
  width: 100vw;
  height: 100vh;
  object-fit: cover;
  position: absolute;
  filter: grayscale(0.3);
  top: 0;
  left: 0;
  z-index: -1;
}

.hero-content {
  position: relative;
  z-index: 1;
  text-align: center;
  color: white;
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.5);
}

.hero-title {
  font-size: 3rem;
  font-weight: 100;
  margin: 0;
  text-transform: uppercase;
  font-family: 'Josefin Sans', sans-serif;
  position: relative;
}

.hero-title::after {
  content: '';
  position: absolute;
  bottom: -12px;
  left: 50%;
  transform: translateX(-50%);
  width: 80px;
  height: 2px;
  background: white;
  opacity: 0.7;
}

.hero-buttons {
  display: flex;
  gap: 1rem;
  margin-top: 3rem;
  justify-content: center;
  align-items: center;
}

.hero-buttons button {
  color: #fff;
  background: #0000;
  border: 1px solid #fff;
  flex: none !important;
  display: inline-block !important;
}
button:hover {
  background: #ffffff;
  color: #000;
}
</style>
