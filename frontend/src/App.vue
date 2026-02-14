<template>
  <div class="app">
    <nav v-if="token">
      <router-link to="/screenings">Screenings</router-link>
      <router-link v-if="role === 'ADMIN'" to="/admin">Admin</router-link>
      <button @click="logout">Logout</button>
    </nav>
    <main>
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
const token = ref('')
const role = ref('')
onMounted(() => {
  token.value = localStorage.getItem('token')
  role.value = localStorage.getItem('role')
})
const logout = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('user_id')
  localStorage.removeItem('role')
  token.value = ''
  role.value = ''
  window.location.href = '/login'
}
</script>

<style>
* { box-sizing: border-box; }
body { margin: 0; font-family: system-ui, sans-serif; background: #0f0f12; color: #e0e0e0; }
.app { min-height: 100vh; }
nav { padding: 0.75rem 1.5rem; background: #1a1a20; display: flex; gap: 1rem; align-items: center; }
nav a { color: #7dd3fc; text-decoration: none; }
nav a.router-link-active { text-decoration: underline; }
nav button { background: #333; color: #e0e0e0; border: 1px solid #555; padding: 0.35rem 0.75rem; cursor: pointer; border-radius: 4px; }
main { padding: 1.5rem; max-width: 1200px; margin: 0 auto; }
</style>
