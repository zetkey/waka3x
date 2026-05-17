<script setup lang="ts">
import { ref, onMounted, watch } from "vue";
import { Search, Loader2 } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { useStatsStore } from "@/stores/stats";
import { formatDate } from "@/lib/formatters";

const statsStore = useStatsStore();
const searchQuery = ref("");
const page = ref(1);

const fetchProjects = () => {
  statsStore.fetchProjects(searchQuery.value, page.value);
};

// Debounce search
let timeout: any;
watch(searchQuery, () => {
  clearTimeout(timeout);
  page.value = 1;
  timeout = setTimeout(fetchProjects, 300);
});

watch(page, fetchProjects);

onMounted(fetchProjects);
</script>

<template>
  <div class="p-4 md:p-8 space-y-8">
    <div
      class="flex flex-col md:flex-row md:items-center justify-between gap-4"
    >
      <div>
        <h1 class="text-3xl font-bold tracking-tight text-primary">
          Your Projects
        </h1>
        <p class="text-muted-foreground">
          Overview of all your projects, ordered by recent activity.
        </p>
      </div>
      <div class="flex gap-2">
        <div class="relative w-64">
          <Search
            class="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground"
          />
          <Input
            v-model="searchQuery"
            placeholder="Filter projects..."
            class="pl-9 bg-muted/50 border-none h-9 focus-visible:ring-1"
          />
        </div>
      </div>
    </div>

    <div
      v-if="statsStore.loading && !statsStore.projects.length"
      class="flex justify-center py-20"
    >
      <Loader2 class="w-8 h-8 animate-spin text-primary" />
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <RouterLink
        v-for="project in statsStore.projects"
        :key="project.project"
        :to="{
          path: '/summary',
          query: { interval: 'all_time', project: project.project },
        }"
      >
        <Card
          class="group relative overflow-hidden transition-all hover:shadow-lg hover:-translate-y-1"
        >
          <CardHeader>
            <div class="flex justify-between items-start">
              <CardTitle
                class="text-xl text-primary group-hover:text-primary transition-colors"
              >
                {{ project.project }}
              </CardTitle>
              <span
                class="px-2 py-0.5 rounded-full bg-muted text-[10px] font-bold uppercase tracking-wider text-muted-foreground"
              >
                {{ project.top_language || "Unknown" }}
              </span>
            </div>
            <CardDescription
              >{{ project.total_heartbeats }} heartbeats
              recorded</CardDescription
            >
          </CardHeader>
          <CardContent>
            <div
              class="flex items-center gap-2 text-xs text-muted-foreground font-mono"
            >
              <span>{{ formatDate(project.first_heartbeat) }}</span>
              <span>→</span>
              <span>{{ formatDate(project.last_heartbeat) }}</span>
            </div>
          </CardContent>
          <div
            class="absolute bottom-0 left-0 h-1 bg-primary transition-all duration-500 w-0 group-hover:w-full"
          ></div>
        </Card>
      </RouterLink>

      <div
        v-if="!statsStore.projects.length"
        class="col-span-full text-center py-20 border border-dashed rounded-lg bg-muted/5"
      >
        <p class="text-muted-foreground italic">
          No projects found matching your search
        </p>
      </div>
    </div>

    <!-- Simple pagination for now -->
    <div class="flex justify-center items-center gap-4 mt-12">
      <Button
        variant="outline"
        :disabled="page <= 1 || statsStore.loading"
        @click="page--"
        >Previous</Button
      >
      <span class="text-sm text-muted-foreground"
        >Page {{ page }} · Showing
        {{ statsStore.projects.length }} projects</span
      >
      <Button
        variant="outline"
        :disabled="statsStore.projects.length < 24 || statsStore.loading"
        @click="page++"
        >Next</Button
      >
    </div>
  </div>
</template>
