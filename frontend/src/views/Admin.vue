<template>
  <div class="space-y-10">
    <h1 class="text-2xl font-bold text-stone-800 sm:text-3xl">Admin</h1>

    <!-- Create screening -->
    <section class="rounded-xl border border-stone-200 bg-stone-50 p-6">
      <h2 class="mb-4 text-lg font-semibold text-stone-800">Create screening</h2>
      <form @submit.prevent="createScreening" class="flex flex-wrap items-end gap-3">
        <input
          v-model="form.movie_id"
          placeholder="Movie ID"
          required
          class="w-28 rounded-lg border border-stone-300 bg-white px-3 py-2 text-sm text-stone-800 placeholder-stone-400 outline-none focus:border-amber-500"
        />
        <input
          v-model="form.movie_name"
          placeholder="Movie name"
          required
          class="min-w-[160px] rounded-lg border border-stone-300 bg-white px-3 py-2 text-sm text-stone-800 placeholder-stone-400 outline-none focus:border-amber-500"
        />
        <input
          v-model="form.screen_at"
          type="datetime-local"
          required
          class="rounded-lg border border-stone-300 bg-white px-3 py-2 text-sm text-stone-800 outline-none focus:border-amber-500"
        />
        <input
          v-model.number="form.rows"
          type="number"
          min="1"
          placeholder="Rows"
          class="w-20 rounded-lg border border-stone-300 bg-white px-3 py-2 text-sm text-stone-800 placeholder-stone-400 outline-none focus:border-amber-500"
        />
        <input
          v-model.number="form.cols"
          type="number"
          min="1"
          placeholder="Cols"
          class="w-20 rounded-lg border border-stone-300 bg-white px-3 py-2 text-sm text-stone-800 placeholder-stone-400 outline-none focus:border-amber-500"
        />
        <button
          type="submit"
          :disabled="creating"
          class="rounded-lg bg-amber-500 px-4 py-2 text-sm font-medium text-stone-900 transition hover:bg-amber-400 disabled:opacity-60"
        >
          {{ creating ? 'Creating...' : 'Create' }}
        </button>
      </form>
      <p v-if="createMessage" class="mt-3 text-sm text-green-600">{{ createMessage }}</p>
    </section>

    <!-- Bookings -->
    <section class="rounded-xl border border-stone-200 bg-stone-50 p-6">
      <h2 class="mb-4 text-lg font-semibold text-stone-800">Bookings</h2>
      <div class="mb-4 flex flex-wrap gap-2">
        <input
          v-model="filters.user_id"
          placeholder="User ID"
          class="w-32 rounded-lg border border-stone-300 bg-white px-3 py-2 text-sm text-stone-800 placeholder-stone-400 outline-none focus:border-amber-500"
        />
        <input
          v-model="filters.screening_id"
          placeholder="Screening ID"
          class="w-32 rounded-lg border border-stone-300 bg-white px-3 py-2 text-sm text-stone-800 placeholder-stone-400 outline-none focus:border-amber-500"
        />
        <input
          v-model="filters.movie_id"
          placeholder="Movie ID"
          class="w-32 rounded-lg border border-stone-300 bg-white px-3 py-2 text-sm text-stone-800 placeholder-stone-400 outline-none focus:border-amber-500"
        />
        <button
          type="button"
          class="rounded-lg bg-stone-200 px-4 py-2 text-sm font-medium text-stone-700 transition hover:bg-stone-300"
          @click="loadBookings"
        >
          Apply
        </button>
      </div>
      <p v-if="bookingsLoading" class="flex items-center gap-2 text-stone-500">
        <span class="inline-block h-4 w-4 animate-spin rounded-full border-2 border-amber-500 border-t-transparent" />
        Loading...
      </p>
      <div v-else class="overflow-x-auto rounded-lg border border-stone-200">
        <table class="w-full min-w-[500px] text-left text-sm">
          <thead class="border-b border-stone-200 bg-stone-100">
            <tr>
              <th class="px-4 py-3 font-medium text-stone-700">Movie</th>
              <th class="px-4 py-3 font-medium text-stone-700">User ID</th>
              <th class="px-4 py-3 font-medium text-stone-700">Seat</th>
              <th class="px-4 py-3 font-medium text-stone-700">Status</th>
              <th class="px-4 py-3 font-medium text-stone-700">Created</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-stone-200">
            <tr v-for="row in bookings" :key="row.booking?.id" class="text-stone-600">
              <td class="px-4 py-3 text-stone-800">{{ row.movie_name || '-' }}</td>
              <td class="px-4 py-3">{{ row.booking?.user_id }}</td>
              <td class="px-4 py-3">{{ row.booking?.seat_row }}-{{ row.booking?.seat_col }}</td>
              <td class="px-4 py-3">{{ row.booking?.status }}</td>
              <td class="px-4 py-3">{{ formatDate(row.booking?.created_at) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <!-- Audit logs -->
    <section class="rounded-xl border border-stone-200 bg-stone-50 p-6">
      <div class="mb-4 flex items-center justify-between">
        <h2 class="text-lg font-semibold text-stone-800">Audit logs</h2>
        <button
          type="button"
          class="rounded-lg border border-stone-300 bg-white px-4 py-2 text-sm text-stone-700 transition hover:bg-stone-100"
          @click="loadLogs"
        >
          Refresh
        </button>
      </div>
      <p v-if="logsLoading" class="flex items-center gap-2 text-stone-500">
        <span class="inline-block h-4 w-4 animate-spin rounded-full border-2 border-amber-500 border-t-transparent" />
        Loading...
      </p>
      <ul v-else class="space-y-2">
        <li
          v-for="(log, i) in logs"
          :key="i"
          class="rounded-lg border border-stone-200 bg-white px-4 py-2 text-sm"
        >
          <strong class="text-amber-600">{{ log.event }}</strong>
          <span class="text-stone-600"> {{ JSON.stringify(log.payload) }}</span>
          <span class="ml-2 text-stone-400">{{ formatDate(log.created_at) }}</span>
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
