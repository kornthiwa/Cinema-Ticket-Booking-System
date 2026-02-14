<template>
  <div
    class="flex min-h-[80vh] flex-col items-center justify-center px-4 py-12"
  >
    <div class="w-full max-w-md">
      <h1
        class="mb-2 text-center text-3xl font-bold tracking-tight text-stone-800 sm:text-4xl"
      >
        Cinema Booking
      </h1>
      <p class="mb-8 text-center text-stone-500">จองที่นั่งภาพยนตร์</p>

      <div
        class="rounded-2xl border border-stone-200 bg-stone-50 p-6 shadow-xl sm:p-8"
      >
        <h2 class="mb-1 text-xl font-semibold text-stone-800">Login</h2>
        <p class="mb-5 text-sm text-stone-500">
          ใช้ email + รหัสผ่าน (seed: user@cinema.local / 123456)
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
            v-model="password"
            type="password"
            placeholder="Password"
            required
            autocomplete="current-password"
            class="w-full rounded-lg border border-stone-300 bg-white px-4 py-3 text-stone-800 placeholder-stone-400 outline-none transition focus:border-amber-500 focus:ring-1 focus:ring-amber-500"
          />
          <button
            type="submit"
            :disabled="loading"
            class="w-full rounded-lg bg-amber-500 px-4 py-3 font-medium text-stone-950 transition hover:bg-amber-400 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {{ loading ? "กำลังเข้าสู่ระบบ..." : "Login" }}
          </button>
        </form>
        <p v-if="error" class="mt-4 text-sm text-red-600">{{ error }}</p>
        <p class="mt-4 text-center text-sm text-stone-500">
          ยังไม่มีบัญชี?
          <router-link to="/register" class="text-amber-600 hover:underline"
            >ลงทะเบียน</router-link
          >
        </p>
      </div>

      <p class="mt-6 text-center text-sm text-stone-500">
        <button
          type="button"
          @click="adminLogin"
          class="rounded-lg border border-stone-300 bg-stone-100 px-4 py-2 text-stone-600 transition hover:bg-stone-200"
        >
          Admin login
        </button>
        <span class="ml-2">(admin@cinema.local / 123456)</span>
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { login, adminLogin as adminLoginApi } from "../api";

const router = useRouter();
const email = ref("");
const password = ref("");
const loading = ref(false);
const error = ref("");

async function submit() {
  error.value = "";
  loading.value = true;
  try {
    const res = await login({ email: email.value, password: password.value });
    localStorage.setItem("token", res.token);
    localStorage.setItem("user_id", res.user_id);
    localStorage.setItem("role", res.role || "USER");
    router.push("/screenings");
  } catch (e) {
    error.value = e.message;
  } finally {
    loading.value = false;
  }
}
async function adminLogin() {
  error.value = "";
  email.value = "admin@cinema.local";
  password.value = "123456";
  loading.value = true;
  try {
    const res = await adminLoginApi({
      email: email.value,
      password: password.value,
    });
    localStorage.setItem("token", res.token);
    localStorage.setItem("user_id", res.user_id);
    localStorage.setItem("role", res.role || "ADMIN");
    router.push("/admin");
  } catch (e) {
    error.value = e.message;
  } finally {
    loading.value = false;
  }
}
</script>
