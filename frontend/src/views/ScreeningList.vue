<template>
  <div class="screenings">
    <h1>Screenings</h1>
    <p v-if="loading">Loading...</p>
    <ul v-else class="list">
      <li v-for="s in list" :key="s.id" class="card">
        <router-link :to="`/screenings/${s.id}`">
          <strong>{{ s.movie_name }}</strong>
          <span>{{ formatDate(s.screen_at) }}</span>
          <span>{{ s.rows }}×{{ s.cols }} seats</span>
        </router-link>
      </li>
    </ul>
    <p v-if="!loading && list.length === 0">No screenings. Admin can create one from Admin → Create screening.</p>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getScreenings } from '../api'

const list = ref([])
const loading = ref(true)

onMounted(async () => {
  try {
    list.value = await getScreenings()
  } catch {
    list.value = []
  } finally {
    loading.value = false
  }
})

function formatDate(d) {
  if (!d) return ''
  const dt = new Date(d)
  return dt.toLocaleString()
}
</script>

<style scoped>
.screenings h1 { margin-bottom: 1rem; }
.list { list-style: none; padding: 0; margin: 0; display: grid; gap: 0.75rem; }
.card { background: #1a1a20; border-radius: 8px; padding: 1rem; }
.card a { text-decoration: none; color: inherit; display: flex; flex-wrap: wrap; gap: 1rem; align-items: center; }
.card a strong { flex: 1 1 100%; }
.card span { color: #888; font-size: 0.9rem; }
</style>
