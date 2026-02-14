<template>
  <div class="seat-map">
    <h1>{{ screening?.movie_name }}</h1>
    <p v-if="screening">{{ formatDate(screening.screen_at) }}</p>
    <p v-if="!loading && !screening">Screening not found.</p>
    <template v-else-if="screening">
      <div class="legend">
        <span><span class="box available"></span> Available</span>
        <span><span class="box locked"></span> Locked</span>
        <span><span class="box booked"></span> Booked</span>
      </div>
      <div class="grid" :style="{ gridTemplateColumns: `repeat(${screening.cols}, 1fr)` }">
        <button
          v-for="seat in flatSeats"
          :key="`${seat.row}-${seat.col}`"
          class="seat"
          :class="seat.status.toLowerCase()"
          :disabled="seat.status !== 'AVAILABLE' || (myLock && myLock.row === seat.row && myLock.col === seat.col)"
          @click="onSeat(seat)"
        >
          {{ seat.row + 1 }}-{{ seat.col + 1 }}
        </button>
      </div>
      <div v-if="myLock" class="actions">
        <p>Seat {{ myLock.row + 1 }}-{{ myLock.col + 1 }} locked. Pay within 5 minutes.</p>
        <button @click="confirmPay" :disabled="confirming">Confirm payment (mock)</button>
        <button class="secondary" @click="myLock = null">Cancel</button>
      </div>
      <p v-if="message" class="message" :class="messageType">{{ message }}</p>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { getScreening, getSeatMap, lockSeat, confirmPayment, wsUrl } from '../api'

const route = useRoute()
const screening = ref(null)
const seats = ref([])
const loading = ref(true)
const myLock = ref(null)
const confirming = ref(false)
const message = ref('')
const messageType = ref('info')
let ws = null

const flatSeats = computed(() => {
  const s = seats.value
  if (!s || !s.length) return []
  return s.flat()
})

onMounted(async () => {
  const id = route.params.id
  try {
    screening.value = await getScreening(id)
    const data = await getSeatMap(id)
    seats.value = data.seats || []
    connectWs(id)
  } catch (e) {
    message.value = e.message
    messageType.value = 'error'
  } finally {
    loading.value = false
  }
})

onUnmounted(() => {
  if (ws) ws.close()
})

function connectWs(id) {
  const url = wsUrl(id)
  ws = new WebSocket(url)
  ws.onmessage = (ev) => {
    try {
      const msg = JSON.parse(ev.data)
      if (msg.type === 'SEAT_UPDATE' && msg.payload) {
        updateSeat(msg.payload)
      }
    } catch (_) {}
  }
  ws.onclose = () => {
    ws = null
  }
}

function updateSeat(payload) {
  const { row, col } = payload
  if (!seats.value[row]) return
  if (!seats.value[row][col]) return
  seats.value[row][col] = { ...seats.value[row][col], ...payload }
}

async function onSeat(seat) {
  if (seat.status !== 'AVAILABLE') return
  message.value = ''
  try {
    const res = await lockSeat(route.params.id, seat.row, seat.col)
    myLock.value = { row: seat.row, col: seat.col, bookingId: res.booking_id }
    setMessage('Seat locked. Confirm payment within 5 minutes.', 'success')
    const data = await getSeatMap(route.params.id)
    seats.value = data.seats || []
  } catch (e) {
    setMessage(e.message, 'error')
  }
}

async function confirmPay() {
  if (!myLock.value?.bookingId) return
  confirming.value = true
  message.value = ''
  try {
    await confirmPayment(myLock.value.bookingId)
    setMessage('Booking confirmed.', 'success')
    myLock.value = null
    const data = await getSeatMap(route.params.id)
    seats.value = data.seats || []
  } catch (e) {
    setMessage(e.message, 'error')
  } finally {
    confirming.value = false
  }
}

function setMessage(text, type) {
  message.value = text
  messageType.value = type || 'info'
  if (text) setTimeout(() => { message.value = '' }, 5000)
}

function formatDate(d) {
  if (!d) return ''
  return new Date(d).toLocaleString()
}
</script>

<style scoped>
.seat-map h1 { margin-bottom: 0.25rem; }
.seat-map > p { color: #888; margin-bottom: 1rem; }
.legend { display: flex; gap: 1.5rem; margin-bottom: 1rem; font-size: 0.9rem; }
.box { display: inline-block; width: 1rem; height: 1rem; margin-right: 0.35rem; vertical-align: middle; border-radius: 3px; }
.available { background: #22c55e; }
.locked { background: #eab308; }
.booked { background: #64748b; }
.grid { display: grid; gap: 6px; max-width: max-content; margin-bottom: 1.5rem; }
.seat { width: 36px; height: 36px; border-radius: 6px; border: 1px solid #333; cursor: pointer; font-size: 0.7rem; padding: 0; }
.seat.available { background: #22c55e; color: #0f0f12; }
.seat.available:hover:not(:disabled) { filter: brightness(1.1); }
.seat.locked { background: #eab308; color: #0f0f12; cursor: not-allowed; }
.seat.booked { background: #64748b; color: #94a3b8; cursor: not-allowed; }
.seat:disabled { cursor: not-allowed; opacity: 0.9; }
.actions { padding: 1rem; background: #1a1a20; border-radius: 8px; margin-bottom: 1rem; }
.actions button { margin-right: 0.5rem; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer; border: none; background: #2563eb; color: #fff; }
.actions button.secondary { background: #333; }
.actions button:disabled { opacity: 0.6; cursor: not-allowed; }
.message { padding: 0.5rem; border-radius: 4px; }
.message.success { background: #14532d; color: #86efac; }
.message.error { background: #450a0a; color: #fca5a5; }
.message.info { background: #1e3a5f; color: #93c5fd; }
</style>
