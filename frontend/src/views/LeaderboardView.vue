<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { Crown, Loader2, User } from "lucide-vue-next";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useStatsStore } from "@/stores/stats";
import { formatDateTime, formatDuration } from "@/lib/formatters";
import { useAuthStore } from "@/stores/auth";

const statsStore = useStatsStore();
const activeBy = ref<"total" | "language">("total");
const selectedKey = ref("");
const page = ref(1);

const authStore = useAuthStore();

const details = computed(() => statsStore.leaderboardDetails);
const topKeys = computed(() => details.value?.top_keys || []);
const userTimezone = computed(() => authStore.user?.location || "Local");

function fetchLeaderboard() {
  statsStore.fetchLeaderboard({
    by: activeBy.value === "language" ? "language" : undefined,
    key:
      activeBy.value === "language" && selectedKey.value
        ? selectedKey.value
        : undefined,
    page: page.value,
    page_size: 100,
  });
}

watch(activeBy, () => {
  page.value = 1;
  selectedKey.value = "";
  fetchLeaderboard();
});

watch(selectedKey, () => {
  if (activeBy.value === "language") fetchLeaderboard();
});

watch(page, fetchLeaderboard);

watch(topKeys, (keys) => {
  if (activeBy.value === "language" && keys.length && !selectedKey.value) {
    selectedKey.value = keys[0];
  }
});

onMounted(fetchLeaderboard);
</script>

<template>
  <div :class="authStore.isAuthenticated ? '' :'min-h-screen bg-background text-foreground'">
    <div :class="authStore.isAuthenticated ? 'p-4 md:p-8' : 'max-w-4xl mx-auto px-4 py-8'" class="space-y-8">
      <header v-if="!authStore.isAuthenticated" class="flex items-center justify-between">
        <RouterLink to="/" class="flex items-center gap-2">
          <Rocket class="w-8 h-8 text-primary" />
          <span class="text-2xl font-bold tracking-tight">Waka3x</span>
        </RouterLink>
        <div class="flex gap-2">
          <RouterLink to="/login">
            <Button>Log in</Button>
          </RouterLink>
        </div>
      </header>
    </div>
    <div class="p-4 md:p-8 space-y-8">
      <div class="space-y-4">
        <div class="flex items-end gap-2">
          <h1 class="text-3xl font-bold tracking-tight text-primary">
            Leaderboard
          </h1>
          <span class="text-muted-foreground text-sm pb-1"
            >({{ details?.interval_label || "current scope" }})</span
          >
        </div>
        <p class="text-muted-foreground max-w-2xl">
          Ranking of active users on this server. Aggregated language mode matches
          the old template leaderboard.
        </p>
      </div>
  
      <div class="flex flex-wrap items-center gap-3 border-b pb-3">
        <Button
          :variant="activeBy === 'total' ? 'default' : 'outline'"
          @click="activeBy = 'total'"
        >
          Total time
        </Button>
        <Button
          :variant="activeBy === 'language' ? 'default' : 'outline'"
          @click="activeBy = 'language'"
        >
          By language
        </Button>
        <Select v-if="activeBy === 'language'" v-model="selectedKey">
          <SelectTrigger class="w-56">
            <SelectValue placeholder="Language" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem v-for="key in topKeys" :key="key" :value="key">
              {{ key }}
            </SelectItem>
          </SelectContent>
        </Select>
      </div>
  
      <div
        v-if="statsStore.loading && !statsStore.leaderboard.length"
        class="flex justify-center py-20"
      >
        <Loader2 class="w-8 h-8 animate-spin text-primary" />
      </div>
  
      <div v-else class="bg-card border rounded-lg overflow-hidden">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead class="w-16"> Rank </TableHead>
              <TableHead>User</TableHead>
              <TableHead v-if="activeBy === 'language'">
                Top Languages
              </TableHead>
              <TableHead class="text-right"> Total Time </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow
              v-for="item in statsStore.leaderboard"
              :key="`${item.user_id}-${item.key || 'total'}`"
              class="hover:bg-accent/50 group"
            >
              <TableCell class="font-bold text-lg text-primary">
                <span class="inline-flex items-center gap-1">
                  #{{ item.rank }}
                  <Crown v-if="item.rank === 1" class="w-4 h-4 text-yellow-400" />
                </span>
              </TableCell>
              <TableCell>
                <div class="flex items-center gap-3">
                  <Avatar class="h-8 w-8">
                    <AvatarImage :src="item.user?.avatar_url || ''" />
                    <AvatarFallback><User class="w-4 h-4" /></AvatarFallback>
                  </Avatar>
                  <div class="flex flex-col text-left">
                    <span class="font-bold text-primary tracking-tight"
                      >@{{ item.user_id }}</span
                    >
                    <span
                      v-if="item.user?.has_active_subscription"
                      class="text-xs text-primary"
                      >supporter</span
                    >
                  </div>
                </div>
              </TableCell>
              <TableCell v-if="activeBy === 'language'">
                <div class="flex flex-wrap gap-1">
                  <span
                    v-for="lang in details?.user_languages?.[item.user_id] || []"
                    :key="lang"
                    class="rounded bg-muted px-2 py-0.5 text-xs"
                  >
                    {{ lang }}
                  </span>
                </div>
              </TableCell>
              <TableCell class="text-right font-mono text-primary">
                {{ formatDuration(item.total) }}
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
  
        <div
          v-if="!statsStore.leaderboard.length"
          class="text-center py-20 italic text-muted-foreground"
        >
          The leaderboard is currently empty
        </div>
      </div>
  
      <div
        class="flex items-center justify-between text-xs text-muted-foreground"
      >
        <span
          >Last updated:
          {{ formatDateTime(details?.last_updated, userTimezone) }}</span
        >
        <div class="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            :disabled="page <= 1"
            @click="page--"
          >
            Previous
          </Button>
          <span>Page {{ page }}</span>
          <Button
            variant="outline"
            size="sm"
            :disabled="statsStore.leaderboard.length < 100"
            @click="page++"
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
