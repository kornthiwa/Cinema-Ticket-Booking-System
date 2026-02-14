<template>
  <div>
    <h1 class="mb-6 text-2xl font-bold text-stone-800 sm:text-3xl">Screenings</h1>

    <p v-if="loading" class="flex items-center gap-2 text-stone-500">
      <span class="inline-block h-5 w-5 animate-spin rounded-full border-2 border-amber-500 border-t-transparent" />
      Loading...
    </p>

    <ul v-else class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <li
        v-for="s in list"
        :key="s.id"
        class="rounded-xl border border-stone-200 bg-stone-50 transition hover:border-amber-400 hover:bg-white"
      >
        <div class="p-5">
          <router-link :to="`/screenings/${s.id}`" class="group block">
            <strong class="block text-lg font-semibold text-stone-800 group-hover:text-amber-600">
              {{ s.movie_name }}
            </strong>
            <span class="mt-2 block text-sm text-stone-500">{{ formatDate(s.screen_at) }}</span>
            <span class="mt-1 inline-block rounded-full bg-stone-200 px-3 py-0.5 text-xs text-stone-600">
              {{ s.rows }}×{{ s.cols }} seats
            </span>
          </router-link>
          <button
            type="button"
            class="mt-3 w-full rounded-lg border border-stone-300 bg-white py-2 text-sm font-medium text-stone-600 transition hover:bg-stone-100"
            :disabled="detailsLoading === s.id"
            @click="toggleDetails(s)"
          >
            {{ expandedId === s.id ? 'ซ่อนรายการ' : 'ดูรายการ ล็อก/จอง' }}
          </button>
        </div>

        <!-- รายการ ล็อก / จอง -->
        <div
          v-if="expandedId === s.id"
          class="border-t border-stone-200 bg-white px-5 pb-5 pt-2"
        >
          <p v-if="detailsLoading === s.id" class="flex items-center gap-2 py-2 text-sm text-stone-500">
            <span class="inline-block h-4 w-4 animate-spin rounded-full border-2 border-amber-500 border-t-transparent" />
            โหลดรายการ...
          </p>
          <template v-else-if="details[s.id]">
            <div v-if="details[s.id].locked.length || details[s.id].booked.length" class="space-y-4">
              <!-- ที่นั่งถูกล็อก -->
              <div v-if="details[s.id].locked.length">
                <h4 class="mb-2 text-sm font-semibold text-amber-700">ที่นั่งถูกล็อก</h4>
                <div class="overflow-x-auto rounded-lg border border-stone-200">
                  <table class="w-full min-w-[320px] text-left text-sm">
                    <thead class="bg-stone-100">
                      <tr>
                        <th class="px-3 py-2 font-medium text-stone-700">แถว-คอลัมน์</th>
                        <th class="px-3 py-2 font-medium text-stone-700">ผู้ล็อก (User ID)</th>
                        <th class="px-3 py-2 font-medium text-stone-700">ล็อกเมื่อ</th>
                        <th class="px-3 py-2 font-medium text-stone-700">ปลดล็อคเมื่อ</th>
                      </tr>
                    </thead>
                    <tbody class="divide-y divide-stone-100">
                      <tr v-for="x in details[s.id].locked" :key="`lock-${x.row}-${x.col}`" class="text-stone-600">
                        <td class="px-3 py-2">{{ x.row + 1 }}-{{ x.col + 1 }}</td>
                        <td class="px-3 py-2">{{ x.user_id }}</td>
                        <td class="px-3 py-2">{{ formatDate(x.locked_at) }}</td>
                        <td class="px-3 py-2">{{ formatDate(x.unlocks_at) }}</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
              <!-- ที่นั่งจองแล้ว -->
              <div v-if="details[s.id].booked.length">
                <h4 class="mb-2 text-sm font-semibold text-green-700">ที่นั่งจองแล้ว</h4>
                <div class="overflow-x-auto rounded-lg border border-stone-200">
                  <table class="w-full min-w-[280px] text-left text-sm">
                    <thead class="bg-stone-100">
                      <tr>
                        <th class="px-3 py-2 font-medium text-stone-700">แถว-คอลัมน์</th>
                        <th class="px-3 py-2 font-medium text-stone-700">ผู้จอง (User ID)</th>
                        <th class="px-3 py-2 font-medium text-stone-700">จองเมื่อ</th>
                      </tr>
                    </thead>
                    <tbody class="divide-y divide-stone-100">
                      <tr v-for="x in details[s.id].booked" :key="`book-${x.row}-${x.col}`" class="text-stone-600">
                        <td class="px-3 py-2">{{ x.row + 1 }}-{{ x.col + 1 }}</td>
                        <td class="px-3 py-2">{{ x.user_id }}</td>
                        <td class="px-3 py-2">{{ formatDate(x.booked_at) }}</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
            <p v-else class="py-2 text-sm text-stone-500">ยังไม่มีที่นั่งถูกล็อกหรือจอง</p>
          </template>
        </div>
      </li>
    </ul>

    <p v-if="!loading && list.length === 0" class="mt-8 rounded-lg border border-stone-200 bg-stone-50 p-6 text-center text-stone-500">
      No screenings. Admin can create one from Admin → Create screening.
    </p>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getScreenings, getSeatDetails } from '../api'

const list = ref([])
const loading = ref(true)
const expandedId = ref(null)
const details = ref({})
const detailsLoading = ref(null)

onMounted(async () => {
  try {
    list.value = await getScreenings()
  } catch {
    list.value = []
  } finally {
    loading.value = false
  }
})

async function toggleDetails(s) {
  if (expandedId.value === s.id) {
    expandedId.value = null
    return
  }
  expandedId.value = s.id
  if (details.value[s.id]) return
  detailsLoading.value = s.id
  try {
    const data = await getSeatDetails(s.id)
    details.value[s.id] = {
      locked: data.locked || [],
      booked: data.booked || [],
    }
  } catch {
    details.value[s.id] = { locked: [], booked: [] }
  } finally {
    detailsLoading.value = null
  }
}

function formatDate(d) {
  if (!d) return ''
  const dt = new Date(d)
  return dt.toLocaleString()
}
</script>
