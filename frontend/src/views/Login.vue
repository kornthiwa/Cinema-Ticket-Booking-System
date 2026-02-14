<template>
  <div class="login">
    <h1>Cinema Booking</h1>
    <form @submit.prevent="submit" class="card">
      <h2>Login</h2>
      <p class="hint">Mock auth: enter any user id / email (e.g. from Google)</p>
      <input v-model="email" type="text" placeholder="Email or User ID" required />
      <input v-model="name" type="text" placeholder="Display name" />
      <button type="submit" :disabled="loading">Login</button>
      <p v-if="error" class="error">{{ error }}</p>
    </form>
    <p class="admin-link"><button type="button" @click="adminLogin">Admin login</button> (same form, creates ADMIN user)</p>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { login, adminLogin as adminLoginApi } from '../api'

const router = useRouter()
const email = ref('')
const name = ref('')
const loading = ref(false)
const error = ref('')

async function submit() {
  error.value = ''
  loading.value = true
  try {
    const res = await login({ email: email.value, name: name.value, user_id: email.value })
    localStorage.setItem('token', res.token)
    localStorage.setItem('user_id', res.user_id)
    localStorage.setItem('role', res.role || 'USER')
    router.push('/screenings')
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
async function adminLogin() {
  error.value = ''
  if (!email.value) {
    email.value = 'admin@test.com'
    name.value = 'Admin'
  }
  loading.value = true
  try {
    const res = await adminLoginApi({ email: email.value, name: name.value, user_id: email.value })
    localStorage.setItem('token', res.token)
    localStorage.setItem('user_id', res.user_id)
    localStorage.setItem('role', res.role || 'ADMIN')
    router.push('/admin')
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login { text-align: center; padding: 2rem; }
.login h1 { margin-bottom: 1.5rem; }
.card { max-width: 360px; margin: 0 auto 1rem; padding: 1.5rem; background: #1a1a20; border-radius: 8px; text-align: left; }
.card h2 { margin-top: 0; }
.card input { width: 100%; padding: 0.5rem; margin-bottom: 0.75rem; border: 1px solid #333; border-radius: 4px; background: #0f0f12; color: #e0e0e0; }
.card button { width: 100%; padding: 0.6rem; background: #2563eb; color: #fff; border: none; border-radius: 4px; cursor: pointer; }
.card button:disabled { opacity: 0.6; cursor: not-allowed; }
.hint { font-size: 0.85rem; color: #888; margin-bottom: 1rem; }
.error { color: #f87171; font-size: 0.9rem; margin-top: 0.5rem; }
.admin-link { margin-top: 1rem; }
.admin-link button { background: #333; color: #94a3b8; border: 1px solid #555; padding: 0.35rem 0.75rem; cursor: pointer; border-radius: 4px; }
</style>
