<script setup lang="ts">
import { RouterView, useRoute } from "vue-router";
import { Toaster } from "@/components/ui/sonner";
import { computed, onMounted } from "vue";
import { useAuthStore } from "@/stores/auth";
import { useMetaStore } from "@/stores/meta";
import LandingLayout from "./layouts/LandingLayout.vue";
import DashboardLayout from "./layouts/DashboardLayout.vue";

const authStore = useAuthStore();
const metaStore = useMetaStore();
const route = useRoute();

onMounted(() => {
  authStore.fetchUser();
  metaStore.fetchConfig().catch(() => undefined);
});

const layouts = {
  Landing: LandingLayout,
  Dashboard: DashboardLayout,
};

const currentLayout = computed(() => {
  return (
    layouts[route.meta.layout as keyof typeof layouts] || layouts.Dashboard
  );
});
</script>

<template>
  <component :is="currentLayout">
    <Toaster />
    <RouterView />
  </component>
</template>
