<template>
  <div class="flex min-h-[80vh] flex-col items-center justify-center px-4 py-12">
    <div class="w-full max-w-md">
      <h1 class="mb-2 text-center text-3xl font-bold tracking-tight text-stone-800 sm:text-4xl">
        ลงทะเบียน
      </h1>
      <p class="mb-8 text-center text-stone-500">สร้างบัญชีเพื่อจองที่นั่ง</p>

      <div class="rounded-2xl border border-stone-200 bg-stone-50 p-6 shadow-xl sm:p-8">
        <h2 class="mb-1 text-xl font-semibold text-stone-800">Register</h2>
        <p class="mb-5 text-sm text-stone-500">
          รหัสผ่านอย่างน้อย 6 ตัวอักษร
        </p>
        <form @submit.prevent="submit" class="space-y-4">
          <input
            v-model="email"
            type="email"
            placeholder="Email"
            required
            autocomplete="email"
            class="w-full rounded-lg border border-stone-300 bg-white px-4 py-3 text-stone-800 placeholder-stone-400 outline-none transition focus:border-amber-500 focus:ring-1 focus:ring-amber-500"
          />
          <input
            v-model="name"
            type="text"
            placeholder="Display name (optional)"
            autocomplete="name"
            class="w-full rounded-lg border border-stone-300 bg-white px-4 py-3 text-stone-800 placeholder-stone-400 outline-none transition focus:border-amber-500 focus:ring-1 focus:ring-amber-500"
          />
          <input
            v-model="password"
            type="password"
            placeholder="Password (min 6 characters)"
            required
            minlength="6"
            autocomplete="new-password"
            class="w-full rounded-lg border border-stone-300 bg-white px-4 py-3 text-stone-800 placeholder-stone-400 outline-none transition focus:border-amber-500 focus:ring-1 focus:ring-amber-500"
          />
          <input
            v-model="passwordConfirm"
            type="password"
            placeholder="Confirm password"
            required
            autocomplete="new-password"
            class="w-full rounded-lg border border-stone-300 bg-white px-4 py-3 text-stone-800 placeholder-stone-400 outline-none transition focus:border-amber-500 focus:ring-1 focus:ring-amber-500"
          />
          <button
            type="submit"
            :disabled="loading"
            class="w-full rounded-lg bg-amber-500 px-4 py-3 font-medium text-stone-950 transition hover:bg-amber-400 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {{ loading ? 'กำลังลงทะเบียน...' : 'ลงทะเบียน' }}
          </button>
        </form>
        <div v-if="error" class="mt-4 space-y-2">
          <p class="text-sm text-red-600">{{ error }}</p>
          <p v-if="isEmailTaken" class="text-sm text-stone-600">
            อีเมลนี้ลงทะเบียนแล้ว — ใช้อีเมลและรหัสผ่านเดิม
            <router-link to="/login" class="ml-1 text-amber-600 hover:underline">เข้าสู่ระบบ</router-link>
          </p>
        </div>
        <p class="mt-4 text-center text-sm text-stone-500">
          มีบัญชีอยู่แล้ว?
          <router-link to="/login" class="text-amber-600 hover:underline">เข้าสู่ระบบ</router-link>
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { register } from '../api'

const router = useRouter()
const email = ref('')
const name = ref('')
const password = ref('')
const passwordConfirm = ref('')
const loading = ref(false)
const error = ref('')
const isEmailTaken = ref(false)

async function submit() {
  error.value = ''
  isEmailTaken.value = false
  if (password.value !== passwordConfirm.value) {
    error.value = 'รหัสผ่านไม่ตรงกัน'
    return
  }
  if (password.value.length < 6) {
    error.value = 'รหัสผ่านต้องมีอย่างน้อย 6 ตัวอักษร'
    return
  }
  loading.value = true
  try {
    const res = await register({
      email: email.value,
      name: name.value || undefined,
      password: password.value,
    })
    localStorage.setItem('token', res.token)
    localStorage.setItem('user_id', res.user_id)
    localStorage.setItem('role', res.role || 'USER')
    router.push('/screenings')
  } catch (e) {
    error.value = e.message
    isEmailTaken.value = /already registered|email already/i.test(e.message)
  } finally {
    loading.value = false
  }
}
</script>
