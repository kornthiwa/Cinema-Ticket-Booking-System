<template>
  <div class="admin">
    <h1>Admin</h1>
    <section>
      <h2>Create screening</h2>
      <form @submit.prevent="createScreening" class="form">
        <input v-model="form.movie_id" placeholder="Movie ID" required />
        <input v-model="form.movie_name" placeholder="Movie name" required />
        <input v-model="form.screen_at" type="datetime-local" required />
        <input v-model.number="form.rows" type="number" min="1" placeholder="Rows" />
        <input v-model.number="form.cols" type="number" min="1" placeholder="Cols" />
        <button type="submit" :disabled="creating">Create</button>
      </form>
      <p v-if="createMessage" class="message">{{ createMessage }}</p>
    </section>
    <section>
      <h2>Bookings</h2>
      <div class="filters">
        <input v-model="filters.user_id" placeholder="User ID" />
        <input v-model="filters.screening_id" placeholder="Screening ID" />
        <input v-model="filters.movie_id" placeholder="Movie ID" />
        <button @click="loadBookings">Apply</button>
      </div>
      <p v-if="bookingsLoading">Loading...</p>
      <div v-else class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Movie</th>
              <th>User ID</th>
              <th>Seat</th>
              <th>Status</th>
              <th>Created</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in bookings" :key="row.booking?.id">
              <td>{{ row.movie_name || '-' }}</td>
              <td>{{ row.booking?.user_id }}</td>
              <td>{{ row.booking?.seat_row }}-{{ row.booking?.seat_col }}</td>
              <td>{{ row.booking?.status }}</td>
              <td>{{ formatDate(row.booking?.created_at) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
    <section>
      <h2>Audit logs</h2>
      <button @click="loadLogs">Refresh</button>
      <p v-if="logsLoading">Loading...</p>
      <ul v-else class="logs">
        <li v-for="(log, i) in logs" :key="i">
          <strong>{{ log.event }}</strong> {{ JSON.stringify(log.payload) }}
          <span class="muted">{{ formatDate(log.created_at) }}</span>
        </li>
      </ul>
    </section>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { adminBookings, adminAuditLogs, createScreening as createScreeningApi } from '../api'

const form = ref({ movie_id: '', movie_name: '', screen_at: '', rows: 5, cols: 8 })
const creating = ref(false)
const createMessage = ref('')
const bookings = ref([])
const bookingsLoading = ref(false)
const filters = ref({ user_id: '', screening_id: '', movie_id: '' })
const logs = ref([])
const logsLoading = ref(false)

onMounted(() => {
  loadBookings()
  loadLogs()
})

async function createScreening() {
  creating.value = true
  createMessage.value = ''
  try {
    const d = new Date(form.value.screen_at)
    await createScreeningApi({
      movie_id: form.value.movie_id,
      movie_name: form.value.movie_name,
      screen_at: d.toISOString(),
      rows: form.value.rows || 5,
      cols: form.value.cols || 8,
    })
    createMessage.value = 'Screening created.'
    form.value = { movie_id: '', movie_name: '', screen_at: '', rows: 5, cols: 8 }
  } catch (e) {
    createMessage.value = e.message
  } finally {
    creating.value = false
  }
}

async function loadBookings() {
  bookingsLoading.value = true
  try {
    const q = {}
    if (filters.value.user_id) q.user_id = filters.value.user_id
    if (filters.value.screening_id) q.screening_id = filters.value.screening_id
    if (filters.value.movie_id) q.movie_id = filters.value.movie_id
    bookings.value = await adminBookings(q)
  } catch {
    bookings.value = []
  } finally {
    bookingsLoading.value = false
  }
}

async function loadLogs() {
  logsLoading.value = true
  try {
    const res = await adminAuditLogs()
    logs.value = Array.isArray(res) ? res.reverse() : []
  } catch {
    logs.value = []
  } finally {
    logsLoading.value = false
  }
}

function formatDate(d) {
  if (!d) return ''
  return new Date(d).toLocaleString()
}
</script>

<style scoped>
.admin section { margin-bottom: 2rem; }
.admin h2 { margin-bottom: 0.75rem; font-size: 1.1rem; }
.form { display: flex; flex-wrap: wrap; gap: 0.5rem; align-items: center; margin-bottom: 0.5rem; }
.form input { padding: 0.4rem; border: 1px solid #333; border-radius: 4px; background: #0f0f12; color: #e0e0e0; }
.form button { padding: 0.5rem 1rem; background: #2563eb; color: #fff; border: none; border-radius: 4px; cursor: pointer; }
.filters { display: flex; flex-wrap: wrap; gap: 0.5rem; margin-bottom: 0.75rem; }
.filters input { padding: 0.4rem; width: 140px; border: 1px solid #333; border-radius: 4px; background: #0f0f12; color: #e0e0e0; }
.table-wrap { overflow-x: auto; }
table { width: 100%; border-collapse: collapse; }
th, td { padding: 0.5rem; text-align: left; border-bottom: 1px solid #333; }
.logs { list-style: none; padding: 0; font-size: 0.9rem; }
.logs li { padding: 0.35rem 0; border-bottom: 1px solid #222; }
.muted { color: #666; margin-left: 0.5rem; }
.message { margin-top: 0.5rem; color: #86efac; }
</style>
