<template>
  <div>
    <div class="row">
      <div class="col-md-12">
        <div class="card shadow-sm">
          <div class="card-body">
            <h1 class="card-title mb-4">Welcome to Datadog Tour</h1>
            <p class="lead">
              This is a demonstration application for learning Datadog APM, Logging, and RUM.
            </p>

            <div class="row mt-4">
              <div class="col-md-4">
                <div class="card text-bg-primary mb-3">
                  <div class="card-body">
                    <h5 class="card-title">APM</h5>
                    <p class="card-text">
                      Distributed tracing across Golang API, MySQL, and Redis
                    </p>
                  </div>
                </div>
              </div>

              <div class="col-md-4">
                <div class="card text-bg-success mb-3">
                  <div class="card-body">
                    <h5 class="card-title">Logs</h5>
                    <p class="card-text">
                      Structured JSON logs with trace correlation
                    </p>
                  </div>
                </div>
              </div>

              <div class="col-md-4">
                <div class="card text-bg-info mb-3">
                  <div class="card-body">
                    <h5 class="card-title">Metrics</h5>
                    <p class="card-text">
                      Custom business metrics with DogStatsD
                    </p>
                  </div>
                </div>
              </div>
            </div>

            <div class="mt-4">
              <NuxtLink to="/users" class="btn btn-primary btn-lg me-2">
                Go to User Management →
              </NuxtLink>
              <NuxtLink to="/test" class="btn btn-success btn-lg">
                API Test Endpoints →
              </NuxtLink>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
useHead({
  title: 'Home - Datadog Tour'
})
</script>
