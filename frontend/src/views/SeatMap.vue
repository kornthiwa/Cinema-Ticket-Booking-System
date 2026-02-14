<template>
  <div>
    <h1 class="text-2xl font-bold text-stone-800 sm:text-3xl">{{ screening?.movie_name }}</h1>
    <p v-if="screening" class="mt-1 text-stone-500">{{ formatDate(screening.screen_at) }}</p>
    <p v-if="!loading && !screening" class="mt-4 text-stone-500">Screening not found.</p>

    <template v-else-if="screening">
      <div class="mb-6 mt-6 flex flex-wrap gap-6 text-sm text-stone-600">
        <span class="flex items-center gap-2">
          <span class="h-4 w-4 rounded bg-green-500" /> Available
        </span>
        <span class="flex items-center gap-2">
          <span class="h-4 w-4 rounded bg-amber-500" /> Locked
        </span>
        <span class="flex items-center gap-2">
          <span class="h-4 w-4 rounded bg-stone-500" /> Booked
        </span>
      </div>

      <div
        class="mb-8 inline-grid gap-1.5"
        :style="{ gridTemplateColumns: `repeat(${screening.cols}, minmax(0, 1fr))` }"
      >
        <button
          v-for="seat in flatSeats"
          :key="`${seat.row}-${seat.col}`"
          type="button"
          class="seat h-9 w-9 rounded-lg border text-xs font-medium transition sm:h-10 sm:w-10"
          :class="{
            'cursor-pointer border-green-600 bg-green-500 text-stone-900 hover:bg-green-400': seat.status === 'AVAILABLE' && !(myLock && myLock.row === seat.row && myLock.col === seat.col),
            'cursor-pointer border-amber-600 bg-amber-500 text-stone-900 opacity-90 hover:opacity-100': seat.status === 'LOCKED',
            'cursor-pointer border-stone-500 bg-stone-500 text-stone-200 hover:bg-stone-600': seat.status === 'BOOKED',
            'cursor-not-allowed': seat.status === 'AVAILABLE' && myLock && myLock.row === seat.row && myLock.col === seat.col,
            'ring-2 ring-amber-500 ring-offset-2 ring-offset-white': myLock && myLock.row === seat.row && myLock.col === seat.col,
          }"
          :disabled="seat.status === 'AVAILABLE' && myLock && myLock.row === seat.row && myLock.col === seat.col"
          :title="seat.status !== 'AVAILABLE' ? 'คลิกดูรายละเอียด' : ''"
          @click="onSeat(seat)"
        >
          {{ seat.row + 1 }}-{{ seat.col + 1 }}
        </button>
      </div>

      <!-- คลิกดูรายละเอียดที่นั่ง ล็อก/จอง -->
      <div
        v-if="selectedSeat"
        class="mb-6 rounded-xl border border-stone-200 bg-stone-50 p-4"
      >
        <div class="mb-2 flex items-center justify-between">
          <span class="font-medium text-stone-800">
            ที่นั่ง {{ selectedSeat.row + 1 }}-{{ selectedSeat.col + 1 }}
            <span class="ml-2 rounded px-2 py-0.5 text-xs" :class="selectedSeat.status === 'LOCKED' ? 'bg-amber-100 text-amber-800' : 'bg-stone-200 text-stone-700'">
              {{ selectedSeat.status === 'LOCKED' ? 'ล็อก' : 'จองแล้ว' }}
            </span>
          </span>
          <button
            type="button"
            class="rounded p-1 text-stone-400 hover:bg-stone-200 hover:text-stone-600"
            aria-label="ปิด"
            @click="selectedSeat = null"
          >
            ✕
          </button>
        </div>
        <p class="text-sm text-stone-600">
          <span class="font-medium">ผู้{{ selectedSeat.status === 'LOCKED' ? 'ล็อก' : 'จอง' }}:</span>
          {{ selectedSeat.user_id }}
        </p>
        <p v-if="selectedSeat.status === 'LOCKED'" class="mt-1 text-sm text-stone-600">
          <span class="font-medium">ล็อกเมื่อ:</span> {{ formatDate(selectedSeat.locked_at) }}
        </p>
        <p v-if="selectedSeat.status === 'LOCKED'" class="mt-0.5 text-sm text-stone-600">
          <span class="font-medium">ปลดล็อคเมื่อ:</span> {{ formatDate(selectedSeat.unlocks_at) }}
        </p>
        <p v-if="selectedSeat.status === 'BOOKED' && selectedSeat.booked_at" class="mt-1 text-sm text-stone-600">
          <span class="font-medium">จองเมื่อ:</span> {{ formatDate(selectedSeat.booked_at) }}
        </p>
      </div>

      <div
        v-if="myLock"
        class="mb-6 rounded-xl border border-stone-200 bg-stone-50 p-5"
      >
        <p class="mb-4 text-stone-700">
          Seat <strong>{{ myLock.row + 1 }}-{{ myLock.col + 1 }}</strong> locked. Pay within 5 minutes.
        </p>
        <div class="flex flex-wrap gap-3">
          <button
            type="button"
            :disabled="confirming"
            class="rounded-lg bg-amber-500 px-4 py-2 text-sm font-medium text-stone-900 transition hover:bg-amber-400 disabled:opacity-60"
            @click="confirmPay"
          >
            {{ confirming ? 'Processing...' : 'Confirm payment (mock)' }}
          </button>
          <button
            type="button"
            class="rounded-lg border border-stone-300 bg-white px-4 py-2 text-sm text-stone-700 transition hover:bg-stone-100"
            @click="myLock = null"
          >
            Cancel
          </button>
        </div>
      </div>

      <p
        v-if="message"
        class="rounded-lg px-4 py-2 text-sm"
        :class="{
          'bg-green-100 text-green-800': messageType === 'success',
          'bg-red-100 text-red-800': messageType === 'error',
          'bg-sky-100 text-sky-800': messageType === 'info',
        }"
      >
        {{ message }}
      </p>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { getScreening, getSeatMap, getSeatDetails, lockSeat, confirmPayment, wsUrl } from '../api'

const route = useRoute()
const screening = ref(null)
const seats = ref([])
const loading = ref(true)
const myLock = ref(null)
const confirming = ref(false)
const message = ref('')
const messageType = ref('info')
const seatDetails = ref({ locked: [], booked: [] })
const selectedSeat = ref(null)
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
    const [mapData, detailsData] = await Promise.all([
      getSeatMap(id),
      getSeatDetails(id).catch(() => ({ locked: [], booked: [] })),
    ])
    seats.value = mapData.seats || []
    seatDetails.value = { locked: detailsData.locked || [], booked: detailsData.booked || [] }
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
  if (seat.status === 'LOCKED') {
    const d = seatDetails.value.locked.find((x) => x.row === seat.row && x.col === seat.col)
    selectedSeat.value = d ? { ...d, status: 'LOCKED' } : { row: seat.row, col: seat.col, status: 'LOCKED', user_id: seat.user_id || '-' }
    return
  }
  if (seat.status === 'BOOKED') {
    const d = seatDetails.value.booked.find((x) => x.row === seat.row && x.col === seat.col)
    selectedSeat.value = d ? { ...d, status: 'BOOKED' } : { row: seat.row, col: seat.col, status: 'BOOKED', user_id: seat.user_id || '-' }
    return
  }
  if (seat.status !== 'AVAILABLE') return
  message.value = ''
  selectedSeat.value = null
  try {
    const res = await lockSeat(route.params.id, seat.row, seat.col)
    myLock.value = { row: seat.row, col: seat.col, bookingId: res.booking_id }
    setMessage('Seat locked. Confirm payment within 5 minutes.', 'success')
    const data = await getSeatMap(route.params.id)
    seats.value = data.seats || []
    const detailsData = await getSeatDetails(route.params.id).catch(() => ({}))
    seatDetails.value = { locked: detailsData.locked || [], booked: detailsData.booked || [] }
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
    selectedSeat.value = null
    const [mapData, detailsData] = await Promise.all([
      getSeatMap(route.params.id),
      getSeatDetails(route.params.id).catch(() => ({ locked: [], booked: [] })),
    ])
    seats.value = mapData.seats || []
    seatDetails.value = { locked: detailsData.locked || [], booked: detailsData.booked || [] }
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
