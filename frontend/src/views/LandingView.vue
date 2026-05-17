<script setup lang="ts">
import { computed, onMounted } from "vue";
import { Button } from "@/components/ui/button";
import {
  CheckCircle2,
  Heart,
  LayoutDashboard,
  LogOut,
  Rocket,
  Server,
} from "lucide-vue-next";
import { useAuthStore } from "@/stores/auth";
import { useMetaStore } from "@/stores/meta";
import { GitHubIcon } from "vue3-simple-icons";
import heroImage from "@/assets/hero.png";

const authStore = useAuthStore();
const metaStore = useMetaStore();

const features = [
  "Self-hosted WakaTime-compatible tracking",
  "Project, language, editor, OS and machine summaries",
  "Public leaderboards and profile sharing",
  "Readme badges and API integrations",
  "Weekly e-mail reports",
  "Prometheus metrics and diagnostics",
  "OIDC, passkeys and local authentication",
  "WakaTime relay and import support",
];

const home = computed(() => metaStore.home);
const config = computed(() => metaStore.config);

onMounted(() => {
  metaStore.fetchConfig().catch(() => undefined);
  metaStore.fetchHome().catch(() => undefined);
});
</script>

<template>
  <div
    class="relative bg-background text-foreground min-h-screen overflow-x-hidden"
  >
    <header
      class="max-w-7xl mx-auto px-4 py-6 flex justify-between items-center"
    >
      <RouterLink to="/" class="flex items-center gap-2">
        <img src="@/assets/icon.svg" alt="Waka3x-logo" class="h-10" />
        <span class="text-2xl font-bold tracking-tight italic text-green-800"
          >Waka<sup>3x</sup></span
        >
      </RouterLink>
      <div class="flex gap-2">
        <template v-if="!authStore.isAuthenticated">
          <RouterLink to="/login">
            <Button variant="ghost"> Login </Button>
          </RouterLink>
          <RouterLink
            v-if="config?.allow_signup || config?.invite_codes_enabled"
            to="/signup"
          >
            <Button>Sign up</Button>
          </RouterLink>
        </template>
        <template v-else>
          <RouterLink to="/dashboard">
            <Button variant="ghost" class="gap-2">
              <LayoutDashboard class="w-4 h-4" /> Dashboard
            </Button>
          </RouterLink>
          <Button variant="outline" class="gap-2" @click="authStore.logout">
            <LogOut class="w-4 h-4" /> Logout
          </Button>
        </template>
      </div>
    </header>

    <main class="max-w-7xl mx-auto px-4 pb-12">
      <section
        class="grid lg:grid-cols-[1fr_520px] gap-10 items-center min-h-[calc(100vh-180px)]"
      >
        <div class="space-y-8">
          <div class="space-y-5">
            <h1 class="text-5xl md:text-7xl font-semibold leading-tight">
              Track your coding time.
            </h1>
            <p class="text-muted-foreground text-xl max-w-2xl">
              Waka3x is a self-hosted, WakaTime-compatible backend for coding
              statistics across projects, languages, editors, machines and more.
            </p>
          </div>

          <div class="flex flex-wrap gap-2">
            <RouterLink
              :to="authStore.isAuthenticated ? '/dashboard' : '/login'"
            >
              <Button size="lg" class="gap-2">
                <Rocket class="w-4 h-4" /> Get started
              </Button>
            </RouterLink>
            <RouterLink to="/setup">
              <Button variant="outline" size="lg" class="gap-2">
                <Server class="w-4 h-4" /> Setup
              </Button>
            </RouterLink>
            <a
              href="https://github.com/zetkey/waka3x"
              target="_blank"
              rel="noreferrer"
            >
              <Button variant="outline" size="lg" class="gap-2"
                ><GitHubIcon class="w-5 h-5" /> GitHub</Button
              >
            </a>
          </div>

          <div class="flex flex-wrap gap-6 text-sm text-muted-foreground">
            <div>
              <span class="font-mono text-foreground text-lg">{{
                home?.total_hours ?? 0
              }}</span>
              tracked hours
            </div>
            <div>
              <span class="font-mono text-foreground text-lg">{{
                home?.total_users ?? 0
              }}</span>
              users
            </div>
            <div class="flex items-center gap-2">
              <span class="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
              <span
                ><span class="font-mono text-foreground text-lg">{{
                  home?.currently_online ?? 0
                }}</span>
                active now</span
              >
            </div>
          </div>
        </div>

        <div class="relative">
          <img
            :src="heroImage"
            alt="Waka3x dashboard preview"
            class="w-full rounded-lg border shadow-2xl bg-card object-cover"
          />
        </div>
      </section>

      <section class="py-12 grid md:grid-cols-2 gap-4">
        <div
          v-for="feature in features"
          :key="feature"
          class="flex items-center gap-3 text-muted-foreground"
        >
          <CheckCircle2 class="w-5 h-5 text-primary shrink-0" />
          <span>{{ feature }}</span>
        </div>
      </section>
    </main>

    <footer class="border-t py-6">
      <div
        class="max-w-7xl mx-auto px-4 flex flex-col md:flex-row justify-between items-center gap-4 text-muted-foreground text-sm"
      >
        <div class="font-mono">
          v{{ config?.version || "dev" }} @ {{ config?.db_type || "database" }}
        </div>
        <div class="flex items-center gap-2">
          <span>Made with</span
          ><Heart class="w-4 h-4 text-red-500 fill-red-500" /><span
            >for developers</span
          >
        </div>
        <RouterLink to="/imprint" class="hover:text-foreground">
          Imprint, cookies and privacy
        </RouterLink>
      </div>
    </footer>
  </div>
</template>
