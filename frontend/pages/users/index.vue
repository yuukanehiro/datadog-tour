<template>
  <div>
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h1>User Management</h1>
      <NuxtLink to="/" class="btn btn-outline-secondary">
        ← Back to Home
      </NuxtLink>
    </div>

    <!-- Create User Form -->
    <div class="card shadow-sm mb-4">
      <div class="card-header bg-primary text-white">
        <h5 class="mb-0">Create New User</h5>
      </div>
      <div class="card-body">
        <form @submit.prevent="createUser">
          <div class="row g-3">
            <div class="col-md-5">
              <label for="name" class="form-label">Name</label>
              <input
                v-model="newUser.name"
                type="text"
                class="form-control"
                id="name"
                placeholder="Enter name"
                required
              >
            </div>
            <div class="col-md-5">
              <label for="email" class="form-label">Email</label>
              <input
                v-model="newUser.email"
                type="email"
                class="form-control"
                id="email"
                placeholder="Enter email"
                required
              >
            </div>
            <div class="col-md-2 d-flex align-items-end">
              <button type="submit" class="btn btn-primary w-100" :disabled="creating">
                <span v-if="creating">Creating...</span>
                <span v-else>Create</span>
              </button>
            </div>
          </div>
        </form>

        <div v-if="createMessage" class="alert alert-success mt-3" role="alert">
          {{ createMessage }}
        </div>
        <div v-if="createError" class="alert alert-danger mt-3" role="alert">
          {{ createError }}
        </div>
      </div>
    </div>

    <!-- Datadog Test Endpoints -->
    <div class="card shadow-sm mb-4">
      <div class="card-header bg-warning text-dark">
        <h5 class="mb-0">Datadog Test Endpoints</h5>
      </div>
      <div class="card-body">
        <p class="text-muted">
          These endpoints are for demonstrating Datadog APM features. Check traces in Datadog after calling them.
        </p>
        <div class="d-flex gap-2 flex-wrap">
          <button @click="testSlowEndpoint" class="btn btn-outline-warning" :disabled="testingSlowEndpoint">
            <span v-if="testingSlowEndpoint">Testing...</span>
            <span v-else>Test Slow Endpoint (2s delay)</span>
          </button>
          <button @click="testErrorEndpoint" class="btn btn-outline-danger" :disabled="testingErrorEndpoint">
            <span v-if="testingErrorEndpoint">Testing...</span>
            <span v-else>Test Error Endpoint</span>
          </button>
          <NuxtLink to="/test" class="btn btn-success">
            Go to Full API Test Page →
          </NuxtLink>
        </div>
        <div v-if="testSlowMessage" class="alert alert-info mt-3" role="alert">
          {{ testSlowMessage }}
        </div>
        <div v-if="testErrorMessage" class="alert alert-warning mt-3" role="alert">
          {{ testErrorMessage }}
        </div>
      </div>
    </div>

    <!-- Users List -->
    <div class="card shadow-sm">
      <div class="card-header bg-secondary text-white d-flex justify-content-between align-items-center">
        <h5 class="mb-0">Users List</h5>
        <button @click="fetchUsers" class="btn btn-light btn-sm" :disabled="loading">
          <span v-if="loading">Loading...</span>
          <span v-else>Refresh</span>
        </button>
      </div>
      <div class="card-body">
        <div v-if="loading && users.length === 0" class="text-center py-5">
          <div class="spinner-border text-primary" role="status">
            <span class="visually-hidden">Loading...</span>
          </div>
        </div>

        <div v-else-if="error" class="alert alert-danger" role="alert">
          {{ error }}
        </div>

        <div v-else-if="users.length === 0" class="alert alert-info" role="alert">
          No users found. Create your first user!
        </div>

        <div v-else class="table-responsive">
          <table class="table table-hover">
            <thead>
              <tr>
                <th>ID</th>
                <th>Name</th>
                <th>Email</th>
                <th>Created At</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="user in users" :key="user.id">
                <td>{{ user.id }}</td>
                <td>{{ user.name }}</td>
                <td>{{ user.email }}</td>
                <td>{{ formatDate(user.created_at) }}</td>
                <td>
                  <button @click="viewUser(user.id)" class="btn btn-sm btn-outline-primary">
                    View
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- User Detail Modal -->
    <div v-if="selectedUser" class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,0.5)">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">User Details</h5>
            <button type="button" class="btn-close" @click="selectedUser = null"></button>
          </div>
          <div class="modal-body">
            <dl class="row">
              <dt class="col-sm-3">ID:</dt>
              <dd class="col-sm-9">{{ selectedUser.id }}</dd>

              <dt class="col-sm-3">Name:</dt>
              <dd class="col-sm-9">{{ selectedUser.name }}</dd>

              <dt class="col-sm-3">Email:</dt>
              <dd class="col-sm-9">{{ selectedUser.email }}</dd>

              <dt class="col-sm-3">Created:</dt>
              <dd class="col-sm-9">{{ formatDate(selectedUser.created_at) }}</dd>
            </dl>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="selectedUser = null">
              Close
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
useHead({
  title: 'Users - Datadog Tour'
})

interface User {
  id: number
  name: string
  email: string
  created_at: string
}

const config = useRuntimeConfig()
const apiBase = config.public.apiBase

const users = ref<User[]>([])
const loading = ref(false)
const error = ref('')

const newUser = ref({
  name: '',
  email: ''
})
const creating = ref(false)
const createMessage = ref('')
const createError = ref('')

const selectedUser = ref<User | null>(null)

// Test endpoints
const testingSlowEndpoint = ref(false)
const testingErrorEndpoint = ref(false)
const testSlowMessage = ref('')
const testErrorMessage = ref('')

// Test slow endpoint
const testSlowEndpoint = async () => {
  testingSlowEndpoint.value = true
  testSlowMessage.value = ''

  try {
    const startTime = Date.now()
    const response = await fetch(`${apiBase}/api/slow`)
    const endTime = Date.now()
    const data = await response.json()

    if (data.success) {
      testSlowMessage.value = `Slow endpoint responded in ${endTime - startTime}ms. Check Datadog APM for trace details!`
      setTimeout(() => {
        testSlowMessage.value = ''
      }, 5000)
    }
  } catch (err) {
    testSlowMessage.value = 'Failed to call slow endpoint'
    console.error(err)
  } finally {
    testingSlowEndpoint.value = false
  }
}

// Test error endpoint
const testErrorEndpoint = async () => {
  testingErrorEndpoint.value = true
  testErrorMessage.value = ''

  try {
    const response = await fetch(`${apiBase}/api/error`)
    const data = await response.json()

    // Error endpoint always returns error
    testErrorMessage.value = `Error endpoint called successfully. Error trace sent to Datadog: "${data.message}"`
    setTimeout(() => {
      testErrorMessage.value = ''
    }, 5000)
  } catch (err) {
    testErrorMessage.value = 'Failed to call error endpoint'
    console.error(err)
  } finally {
    testingErrorEndpoint.value = false
  }
}

// Fetch users
const fetchUsers = async () => {
  loading.value = true
  error.value = ''

  try {
    const response = await fetch(`${apiBase}/api/users`)
    const data = await response.json()

    if (data.success) {
      users.value = data.data || []
    } else {
      error.value = data.message || 'Failed to fetch users'
    }
  } catch (err) {
    error.value = 'Failed to connect to API'
    console.error(err)
  } finally {
    loading.value = false
  }
}

// Create user
const createUser = async () => {
  creating.value = true
  createMessage.value = ''
  createError.value = ''

  try {
    const response = await fetch(`${apiBase}/api/users`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(newUser.value)
    })

    const data = await response.json()

    if (data.success) {
      createMessage.value = `User "${newUser.value.name}" created successfully!`
      newUser.value = { name: '', email: '' }
      await fetchUsers()

      setTimeout(() => {
        createMessage.value = ''
      }, 3000)
    } else {
      createError.value = data.message || 'Failed to create user'
    }
  } catch (err) {
    createError.value = 'Failed to connect to API'
    console.error(err)
  } finally {
    creating.value = false
  }
}

// View user detail
const viewUser = async (id: number) => {
  try {
    const response = await fetch(`${apiBase}/api/users/${id}`)
    const data = await response.json()

    if (data.success) {
      selectedUser.value = data.data
    }
  } catch (err) {
    console.error(err)
  }
}

// Format date
const formatDate = (dateString: string) => {
  const date = new Date(dateString)
  return date.toLocaleString()
}

// Fetch users on mount
onMounted(() => {
  fetchUsers()
})
</script>

<style scoped>
.modal.show {
  display: block !important;
}
</style>
