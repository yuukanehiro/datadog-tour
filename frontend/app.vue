<template>
  <div id="app">
    <nav class="navbar navbar-dark bg-dark">
      <div class="container-fluid">
        <a class="navbar-brand" href="/">
          <span class="fw-bold">🐕 Datadog Tour</span>
        </a>
        <span class="navbar-text text-white-50">
          User Management System
        </span>
      </div>
    </nav>

    <main class="container mt-4">
      <NuxtPage />
    </main>

    <footer class="mt-5 py-4 bg-light text-center">
      <div class="container">
        <p class="text-muted mb-0">
          Powered by Nuxt.js + Golang + Datadog
        </p>
      </div>
    </footer>
  </div>
</template>

<style>
#app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

main {
  flex: 1;
}

.navbar-brand {
  font-size: 1.5rem;
}
</style>
