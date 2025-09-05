<template>
  <div class="home-container">
    <div class="feature-container">
      <img
        class="feature-image autumn-image"
        src="@/assets/roberto-nickson-bg.webp"
        alt="Background Image"
        loading="eager"
        decoding="async"
      />
      <div class="hero-content unselectable">
        <h1 class="hero-title">essential living</h1>
        <div class="hero-links">
          <router-link to="/new" class="hero-link">shop new</router-link>
          <router-link to="/popular" class="hero-link">shop popular</router-link>
        </div>
      </div>
    </div>
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
  display: flex;
  align-items: center;
  justify-content: center;
}

.feature-image {
  width: 100vw;
  height: 100vh;
  object-fit: cover;
  position: absolute;
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
  font-size: 4rem;
  font-weight: 400;
  margin: 0;
  text-transform: uppercase;
  font-family: 'Bodoni Moda', serif;
  letter-spacing: 15px;
}

.hero-links {
  display: flex;
  flex-direction: row;
  gap: 2rem;
  align-items: center;
  justify-content: center;
}

.hero-link {
  color: white;
  text-decoration: none;
  font-size: 0.9rem;
  text-transform: lowercase;
  letter-spacing: 1px;
  border-bottom: 1px solid white;
  padding-bottom: 2px;
  transition: all 0.3s ease;
}

.hero-link:hover {
  color: rgba(255, 255, 255, 0.8);
  border-bottom-color: rgba(255, 255, 255, 0.8);
}
</style>
