<script setup lang="ts">
import { computed, onMounted } from "vue";
import {
  Activity,
  AppWindow,
  CalendarDays,
  Clock,
  Code,
  Folder,
  Loader2,
  Monitor,
  Timer,
} from "lucide-vue-next";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useStatsStore } from "@/stores/stats";
import { formatDate, formatDuration } from "@/lib/formatters";
import BarChart from "@/components/charts/BarChart.vue";
import DoughnutChart from "@/components/charts/DoughnutChart.vue";
import type {
  HourlyActivity,
  SummaryItem,
  SummaryRequestParams,
  TimelineDay,
} from "@/types/api";

const statsStore = useStatsStore();

function startOfCurrentWeek() {
  const date = new Date();
  const day = date.getDay();
  const daysSinceMonday = (day + 6) % 7;
  date.setHours(0, 0, 0, 0);
  date.setDate(date.getDate() - daysSinceMonday);
  return date;
}

function toDateInputValue(date: Date) {
  const year = date.getFullYear();
  const month = `${date.getMonth() + 1}`.padStart(2, "0");
  const day = `${date.getDate()}`.padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function formatWeekday(dateStr: string) {
  return new Date(dateStr).toLocaleDateString("en-US", {
    weekday: "short",
  });
}

function formatHour(hour: number) {
  return new Intl.DateTimeFormat("en-US", {
    hour: "numeric",
    hour12: true,
  }).format(new Date(2020, 0, 1, hour));
}

const weekStart = startOfCurrentWeek();
const weekEndExclusive = new Date(weekStart);
weekEndExclusive.setDate(weekStart.getDate() + 7);
const weekEndDisplay = new Date(weekEndExclusive);
weekEndDisplay.setDate(weekEndExclusive.getDate() - 1);

const weeklyParams: SummaryRequestParams = {
  from: toDateInputValue(weekStart),
  to: toDateInputValue(weekEndExclusive),
};

const details = computed(() => statsStore.summaryDetails);
const summary = computed(() => details.value?.summary || statsStore.summary);
const timeline = computed(() => details.value?.timeline || []);
const hourlyActivity = computed(() => details.value?.hourly_activity || []);

const totalSeconds = computed(
  () => summary.value?.projects.reduce((acc, item) => acc + item.total, 0) || 0,
);

function topItem(items: SummaryItem[] | undefined) {
  return items?.[0];
}

function dayTotal(day: TimelineDay) {
  return day.projects.reduce((acc, project) => acc + project.duration, 0);
}

function activeDaysCount(days: TimelineDay[]) {
  return days.filter((day) => dayTotal(day) > 0).length;
}

const mostActiveDay = computed(() => {
  return [...timeline.value].sort((a, b) => dayTotal(b) - dayTotal(a))[0];
});

const mostActiveHour = computed(() => {
  return [...hourlyActivity.value].sort((a, b) => b.duration - a.duration)[0];
});

const weekLabel = computed(
  () => `${formatDate(weekStart.toISOString())} - ${formatDate(weekEndDisplay.toISOString())}`,
);

const reportCards = computed(() => {
  const project = topItem(summary.value?.projects);
  const language = topItem(summary.value?.languages);
  const editor = topItem(summary.value?.editors);
  const os = topItem(summary.value?.operating_systems);
  const day = mostActiveDay.value;
  const hour = mostActiveHour.value;

  return [
    {
      label: "Total Coding Time",
      value: formatDuration(totalSeconds.value),
      description: "Monday to Sunday",
      icon: Timer,
    },
    {
      label: "Most Active Project",
      value: project?.key || "-",
      description: project ? formatDuration(project.total) : "No project time",
      icon: Folder,
    },
    {
      label: "Most Active Day",
      value: day ? formatWeekday(day.date) : "-",
      description: day ? formatDuration(dayTotal(day)) : "No daily activity",
      icon: CalendarDays,
    },
    {
      label: "Most Active Hour",
      value: hour && hour.duration > 0 ? formatHour(hour.hour) : "-",
      description:
        hour && hour.duration > 0
          ? formatDuration(hour.duration)
          : "No hourly activity",
      icon: Clock,
    },
    {
      label: "Top Language",
      value: language?.key || "-",
      description: language ? formatDuration(language.total) : "No language time",
      icon: Code,
    },
    {
      label: "Top Editor",
      value: editor?.key || "-",
      description: editor ? formatDuration(editor.total) : "No editor time",
      icon: AppWindow,
    },
    {
      label: "Top OS",
      value: os?.key || "-",
      description: os ? formatDuration(os.total) : "No OS time",
      icon: Monitor,
    },
    {
      label: "Active Days",
      value: `${activeDaysCount(timeline.value)} / 7`,
      description: "Days with tracked activity",
      icon: Activity,
    },
  ];
});

function top(items: SummaryItem[] | undefined, count: number) {
  return items?.slice(0, count) || [];
}

const projectChartData = computed(() => ({
  labels: top(summary.value?.projects, 8).map((item) => item.key),
  datasets: [
    {
      label: "Hours",
      data: top(summary.value?.projects, 8).map(
        (item) => Math.round((item.total / 3600) * 10) / 10,
      ),
      backgroundColor: "#166534",
      borderRadius: 4,
    },
  ],
}));

const weekdayChartData = computed(() => ({
  labels: timeline.value.map((day) => formatWeekday(day.date)),
  datasets: [
    {
      label: "Hours",
      data: timeline.value.map(
        (day) => Math.round((dayTotal(day) / 3600) * 10) / 10,
      ),
      backgroundColor: "#0f766e",
      borderRadius: 4,
    },
  ],
}));

const hourlyChartData = computed(() => ({
  labels: hourlyActivity.value.map((item: HourlyActivity) =>
    formatHour(item.hour),
  ),
  datasets: [
    {
      label: "Minutes",
      data: hourlyActivity.value.map(
        (item) => Math.round((item.duration / 60) * 10) / 10,
      ),
      backgroundColor: "#4f46e5",
      borderRadius: 4,
    },
  ],
}));

const languageChartData = computed(() => ({
  labels: top(summary.value?.languages, 6).map((item) => item.key),
  datasets: [
    {
      data: top(summary.value?.languages, 6).map((item) => item.total),
      backgroundColor: [
        "#166534",
        "#0f766e",
        "#4f46e5",
        "#ca8a04",
        "#be123c",
        "#7c3aed",
      ],
      borderWidth: 0,
    },
  ],
}));

const weeklyTables = computed(() => [
  { title: "Projects", items: top(summary.value?.projects, 6) },
  { title: "Languages", items: top(summary.value?.languages, 6) },
  { title: "Editors", items: top(summary.value?.editors, 6) },
  { title: "Operating Systems", items: top(summary.value?.operating_systems, 6) },
  { title: "Machines", items: top(summary.value?.machines, 6) },
]);

onMounted(() => {
  statsStore.fetchSummaryDetails(weeklyParams).catch(() => undefined);
});
</script>

<template>
  <div class="p-4 md:p-8 space-y-8">
    <div
      v-if="statsStore.loading && !summary"
      class="flex min-h-80 items-center justify-center"
    >
      <Loader2 class="w-8 h-8 animate-spin text-primary" />
    </div>

    <div v-else class="space-y-8">
      <div class="flex flex-col gap-2">
        <h1 class="text-3xl font-bold tracking-tight text-primary">
          Weekly Report
        </h1>
        <p class="text-muted-foreground text-sm">
          {{ weekLabel }}. Counts reset every Monday and include activity through
          Sunday.
        </p>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4">
        <Card v-for="card in reportCards" :key="card.label">
          <CardHeader class="pb-2">
            <div class="flex items-start justify-between gap-3">
              <CardDescription
                class="text-xs uppercase tracking-wider font-semibold"
              >
                {{ card.label }}
              </CardDescription>
              <component :is="card.icon" class="w-4 h-4 text-primary" />
            </div>
            <CardTitle class="text-2xl font-bold text-primary truncate">
              {{ card.value }}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p class="text-xs text-muted-foreground">
              {{ card.description }}
            </p>
          </CardContent>
        </Card>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <Card class="lg:col-span-2">
          <CardHeader>
            <CardTitle class="text-primary">Top Projects</CardTitle>
            <CardDescription>Weekly coding time by project.</CardDescription>
          </CardHeader>
          <CardContent class="h-80">
            <BarChart v-if="summary?.projects.length" :data="projectChartData" />
            <div
              v-else
              class="h-full flex items-center justify-center text-muted-foreground italic text-sm"
            >
              No project data recorded this week
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle class="text-primary">Language Split</CardTitle>
            <CardDescription>Weekly distribution by language.</CardDescription>
          </CardHeader>
          <CardContent class="h-80">
            <DoughnutChart
              v-if="summary?.languages.length"
              :data="languageChartData"
            />
            <div
              v-else
              class="h-full flex items-center justify-center text-muted-foreground italic text-sm"
            >
              No language data recorded this week
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle class="text-primary">Weekdays</CardTitle>
            <CardDescription>Daily totals from Monday to Sunday.</CardDescription>
          </CardHeader>
          <CardContent class="h-72">
            <BarChart v-if="timeline.length" :data="weekdayChartData" />
            <div
              v-else
              class="h-full flex items-center justify-center text-muted-foreground italic text-sm"
            >
              No daily activity recorded
            </div>
          </CardContent>
        </Card>

        <Card class="lg:col-span-2">
          <CardHeader>
            <CardTitle class="text-primary">Active Hours</CardTitle>
            <CardDescription>Totals grouped by hour across the week.</CardDescription>
          </CardHeader>
          <CardContent class="h-72">
            <BarChart v-if="hourlyActivity.length" :data="hourlyChartData" />
            <div
              v-else
              class="h-full flex items-center justify-center text-muted-foreground italic text-sm"
            >
              No hourly activity recorded
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle class="text-primary">Weekly Breakdown</CardTitle>
          <CardDescription>Same categories used in the weekly report.</CardDescription>
        </CardHeader>
        <CardContent>
          <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-5 gap-8">
            <div
              v-for="section in weeklyTables"
              :key="section.title"
              class="space-y-3"
            >
              <h3 class="text-xs font-bold uppercase text-muted-foreground">
                {{ section.title }}
              </h3>
              <div
                v-for="item in section.items"
                :key="item.key"
                class="flex justify-between gap-4 border-b border-border/50 py-2"
              >
                <span class="text-sm truncate">{{ item.key }}</span>
                <span class="text-sm font-mono text-muted-foreground">
                  {{ formatDuration(item.total) }}
                </span>
              </div>
              <p
                v-if="!section.items.length"
                class="text-sm italic text-muted-foreground"
              >
                No data
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
