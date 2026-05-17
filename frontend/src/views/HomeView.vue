<script setup lang="ts">
import { onMounted, computed } from "vue";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Calendar, Filter, Loader2 } from "lucide-vue-next";
import { useStatsStore } from "@/stores/stats";
import { formatDuration } from "@/lib/formatters";
import BarChart from "@/components/charts/BarChart.vue";
import DoughnutChart from "@/components/charts/DoughnutChart.vue";

const statsStore = useStatsStore();

const stats = computed(() => {
  if (!statsStore.summary) return [];

  const s = statsStore.summary;
  const totalSeconds = s.projects.reduce((acc, p) => acc + p.total, 0);
  const topLang = s.languages[0]?.key || "-";
  const topProject = s.projects[0]?.key || "-";

  return [
    {
      label: "Total Coding Time",
      value: formatDuration(totalSeconds),
      description: "For selected interval",
    },
    {
      label: "Top Project",
      value: topProject,
      description: "Most active project",
    },
    {
      label: "Projects Active",
      value: String(s.projects.length),
      description: "Unique projects",
    },
    { label: "Top Language", value: topLang, description: "Primary language" },
  ];
});

const projectChartData = computed(() => ({
  labels: statsStore.summary?.projects.slice(0, 7).map((p) => p.key) || [],
  datasets: [
    {
      label: "Hours",
      data:
        statsStore.summary?.projects
          .slice(0, 7)
          .map((p) => Math.round((p.total / 3600) * 10) / 10) || [],
      backgroundColor: "#3b82f6",
      borderRadius: 4,
    },
  ],
}));

const languageChartData = computed(() => ({
  labels: statsStore.summary?.languages.slice(0, 5).map((l) => l.key) || [],
  datasets: [
    {
      data: statsStore.summary?.languages.slice(0, 5).map((l) => l.total) || [],
      backgroundColor: ["#3b82f6", "#10b981", "#f59e0b", "#ef4444", "#8b5cf6"],
      borderWidth: 0,
    },
  ],
}));

onMounted(() => {
  statsStore.fetchSummary({ interval: "last_7_days" });
});
</script>

<template>
  <div class="p-4 md:p-8 space-y-8">
    <div
      v-if="statsStore.loading && !statsStore.summary"
      class="flex-1 flex items-center justify-center"
    >
      <Loader2 class="w-8 h-8 animate-spin text-primary" />
    </div>

    <div v-else class="space-y-8">
      <div
        class="flex flex-col md:flex-row md:items-center justify-between gap-4"
      >
        <div>
          <h1 class="text-3xl font-bold tracking-tight text-primary">
            Dashboard
          </h1>
          <p class="text-muted-foreground text-sm">
            Welcome back! Here's your activity for the last 7 days.
          </p>
        </div>
        <div class="flex items-center gap-2">
          <Button variant="outline" size="sm" class="gap-2">
            <Calendar class="w-4 h-4" /> Last 7 Days
          </Button>
          <Button variant="outline" size="sm" class="gap-2">
            <Filter class="w-4 h-4" /> Filter
          </Button>
          <Button size="sm"> Export Report </Button>
        </div>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card v-for="stat in stats" :key="stat.label">
          <CardHeader class="pb-2">
            <CardDescription
              class="text-xs uppercase tracking-wider font-semibold"
            >
              {{ stat.label }}
            </CardDescription>
            <CardTitle class="text-2xl font-bold text-primary">
              {{ stat.value }}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p class="text-xs text-muted-foreground">
              {{ stat.description }}
            </p>
          </CardContent>
        </Card>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <Card class="lg:col-span-2">
          <CardHeader>
            <CardTitle class="text-primary"> Top Projects </CardTitle>
            <CardDescription>Hours spent coding by project.</CardDescription>
          </CardHeader>
          <CardContent class="h-80">
            <BarChart
              v-if="statsStore.summary?.projects.length"
              :data="projectChartData"
            />
            <div
              v-else
              class="h-full flex items-center justify-center text-muted-foreground italic text-sm"
            >
              No project data recorded
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle class="text-primary"> Language Split </CardTitle>
            <CardDescription>Distribution of time by language.</CardDescription>
          </CardHeader>
          <CardContent class="h-80">
            <DoughnutChart
              v-if="statsStore.summary?.languages.length"
              :data="languageChartData"
            />
            <div
              v-else
              class="h-full flex items-center justify-center text-muted-foreground italic text-sm"
            >
              No language data recorded
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  </div>
</template>
