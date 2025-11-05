<template>
  <div>
    <div class="row">
      <div class="col-md-12">
        <div class="card shadow-sm">
          <div class="card-body">
            <h1 class="card-title mb-4">API Test Endpoints</h1>
            <p class="lead">
              Test various API endpoints to see how errors and traces appear in Datadog
            </p>

            <div class="row mt-4">
              <!-- Slow Endpoint -->
              <div class="col-md-6 mb-3">
                <div class="card">
                  <div class="card-header bg-info text-white">
                    <h5 class="mb-0">Slow Request</h5>
                  </div>
                  <div class="card-body">
                    <p>Simulates a slow API call (2 second delay)</p>
                    <button
                      class="btn btn-info"
                      @click="callEndpoint('/api/slow')"
                      :disabled="loading"
                    >
                      {{ loading && currentEndpoint === '/api/slow' ? 'Loading...' : 'Test Slow API' }}
                    </button>
                  </div>
                </div>
              </div>

              <!-- Error Endpoint -->
              <div class="col-md-6 mb-3">
                <div class="card">
                  <div class="card-header bg-danger text-white">
                    <h5 class="mb-0">Unexpected Error</h5>
                  </div>
                  <div class="card-body">
                    <p>Triggers a system error (will trigger alerts)</p>
                    <button
                      class="btn btn-danger"
                      @click="callEndpoint('/api/error')"
                      :disabled="loading"
                    >
                      {{ loading && currentEndpoint === '/api/error' ? 'Loading...' : 'Trigger Error' }}
                    </button>
                  </div>
                </div>
              </div>

              <!-- Expected Error Endpoint -->
              <div class="col-md-6 mb-3">
                <div class="card">
                  <div class="card-header bg-warning text-dark">
                    <h5 class="mb-0">Expected Error (No Alert)</h5>
                  </div>
                  <div class="card-body">
                    <p>Triggers an expected error (validation error, no alert)</p>
                    <button
                      class="btn btn-warning"
                      @click="callEndpoint('/api/expected-error')"
                      :disabled="loading"
                    >
                      {{ loading && currentEndpoint === '/api/expected-error' ? 'Loading...' : 'Expected Error' }}
                    </button>
                  </div>
                </div>
              </div>

              <!-- Unexpected Error Endpoint -->
              <div class="col-md-6 mb-3">
                <div class="card">
                  <div class="card-header bg-danger text-white">
                    <h5 class="mb-0">System Error (Alert)</h5>
                  </div>
                  <div class="card-body">
                    <p>Triggers a system error (will trigger alerts)</p>
                    <button
                      class="btn btn-danger"
                      @click="callEndpoint('/api/unexpected-error')"
                      :disabled="loading"
                    >
                      {{ loading && currentEndpoint === '/api/unexpected-error' ? 'Loading...' : 'System Error' }}
                    </button>
                  </div>
                </div>
              </div>

              <!-- Warning Endpoint -->
              <div class="col-md-6 mb-3">
                <div class="card">
                  <div class="card-header bg-warning text-dark">
                    <h5 class="mb-0">Warning Log</h5>
                  </div>
                  <div class="card-body">
                    <p>Logs a warning message (performance degradation)</p>
                    <button
                      class="btn btn-warning"
                      @click="callEndpoint('/api/warn')"
                      :disabled="loading"
                    >
                      {{ loading && currentEndpoint === '/api/warn' ? 'Loading...' : 'Test Warning' }}
                    </button>
                  </div>
                </div>
              </div>
            </div>

            <!-- Response Display -->
            <div v-if="response" class="mt-4">
              <div class="card">
                <div class="card-header" :class="{
                  'bg-success text-white': response.status >= 200 && response.status < 300,
                  'bg-warning text-dark': response.status >= 400 && response.status < 500,
                  'bg-danger text-white': response.status >= 500
                }">
                  <h5 class="mb-0">Response (Status: {{ response.status }})</h5>
                </div>
                <div class="card-body">
                  <div v-if="response.headers" class="mb-3">
                    <h6>Headers:</h6>
                    <ul class="list-unstyled">
                      <li v-if="response.headers['x-datadog-trace-id']">
                        <strong>X-Datadog-Trace-Id:</strong> {{ response.headers['x-datadog-trace-id'] }}
                      </li>
                      <li v-if="response.headers['x-datadog-span-id']">
                        <strong>X-Datadog-Span-Id:</strong> {{ response.headers['x-datadog-span-id'] }}
                      </li>
                    </ul>
                  </div>
                  <h6>Body:</h6>
                  <pre class="bg-light p-3 rounded"><code>{{ JSON.stringify(response.data, null, 2) }}</code></pre>
                </div>
              </div>
            </div>

            <div class="mt-4">
              <NuxtLink to="/" class="btn btn-secondary">
                ← Back to Home
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
  title: 'API Test - Datadog Tour'
})

const config = useRuntimeConfig()
const apiBase = config.public.apiBase

const loading = ref(false)
const currentEndpoint = ref('')
const response = ref<{
  status: number
  data: any
  headers: Record<string, string>
} | null>(null)

async function callEndpoint(endpoint: string) {
  loading.value = true
  currentEndpoint.value = endpoint
  response.value = null

  try {
    const res = await fetch(`${apiBase}${endpoint}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    })

    const data = await res.json()

    response.value = {
      status: res.status,
      data: data,
      headers: {
        'x-datadog-trace-id': res.headers.get('x-datadog-trace-id') || '',
        'x-datadog-span-id': res.headers.get('x-datadog-span-id') || ''
      }
    }
  } catch (error: any) {
    response.value = {
      status: 0,
      data: { error: error.message || 'Network error' },
      headers: {}
    }
  } finally {
    loading.value = false
    currentEndpoint.value = ''
  }
}
</script>
