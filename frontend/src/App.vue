<template>
  <div class="min-h-screen bg-white font-sans text-stone-800">
    <nav
      v-if="token"
      class="sticky top-0 z-50 border-b border-stone-200 bg-white/95 backdrop-blur"
    >
      <div
        class="mx-auto flex max-w-6xl items-center justify-between px-4 py-3 sm:px-6"
      >
        <div class="flex items-center gap-6">
          <router-link
            to="/screenings"
            class="text-stone-600 transition hover:text-amber-600"
            active-class="text-amber-600 font-medium"
          >
            หน้าหลัก (Screenings)
          </router-link>
          <router-link
            v-if="role === 'ADMIN'"
            to="/admin"
            class="text-stone-600 transition hover:text-amber-600"
            active-class="text-amber-600 font-medium"
          >
            Admin
          </router-link>
        </div>
        <button
          type="button"
          @click="logout"
          class="rounded-lg border border-stone-300 bg-stone-100 px-4 py-2 text-sm font-medium text-stone-700 transition hover:bg-stone-200"
        >
          ออกระบบ
        </button>
      </div>
    </nav>
    <main class="mx-auto max-w-6xl px-4 py-6 sm:px-6 sm:py-8">
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from "vue";
import { useRoute } from "vue-router";

const token = ref("");
const role = ref("");
const route = useRoute();

function loadAuth() {
  token.value = localStorage.getItem("token");
  role.value = localStorage.getItem("role");
}

onMounted(loadAuth);
// อัปเดต token/role ตอนเปลี่ยน route (เช่น หลัง login/register ด้วย router.push) เพื่อให้ navbar ขึ้น
watch(() => route.path, loadAuth);
const logout = () => {
  localStorage.removeItem("token");
  localStorage.removeItem("user_id");
  localStorage.removeItem("role");
  token.value = "";
  role.value = "";
  window.location.href = "/login";
};
</script>
