<script setup lang="ts">
import { computed, watch } from "vue";
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
import { useAuthStore } from "@/stores/auth";
import { useStatsStore } from "@/stores/stats";
import { formatDate, formatDuration, formatWeekday } from "@/lib/formatters";
import BarChart from "@/components/charts/BarChart.vue";
import DoughnutChart from "@/components/charts/DoughnutChart.vue";
import type {
  HourlyActivity,
  SummaryItem,
  SummaryRequestParams,
  TimelineDay,
} from "@/types/api";

const authStore = useAuthStore();
const statsStore = useStatsStore();
const browserTimezone =
  Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";

function browserSafeTimezone(timeZone: string) {
  if (!timeZone || timeZone === "Local") return browserTimezone;

  try {
    new Intl.DateTimeFormat("en-US", { timeZone });
    return timeZone;
  } catch {
    return browserTimezone;
  }
}

function datePartsInTimezone(date: Date, timeZone: string) {
  const safeTimezone = browserSafeTimezone(timeZone);
  const parts = new Intl.DateTimeFormat("en-US", {
    timeZone: safeTimezone,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    weekday: "short",
  }).formatToParts(date);
  const get = (type: Intl.DateTimeFormatPartTypes) =>
    parts.find((part) => part.type === type)?.value || "";

  return {
    year: Number(get("year")),
    month: Number(get("month")),
    day: Number(get("day")),
    weekday: get("weekday"),
  };
}

function dateInputValueInTimezone(date: Date, timeZone: string) {
  const { year, month, day } = datePartsInTimezone(date, timeZone);
  return `${year}-${`${month}`.padStart(2, "0")}-${`${day}`.padStart(2, "0")}`;
}

function startOfCurrentWeek(timeZone: string) {
  const now = new Date();
  const parts = datePartsInTimezone(now, timeZone);
  const weekdayIndex =
    ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"].indexOf(parts.weekday);
  const day = weekdayIndex >= 0 ? weekdayIndex : now.getDay();
  const daysSinceMonday = (day + 6) % 7;
  const start = new Date(Date.UTC(parts.year, parts.month - 1, parts.day));
  start.setUTCDate(start.getUTCDate() - daysSinceMonday);
  return start;
}

function formatHour(hour: number) {
  return new Intl.DateTimeFormat("en-US", {
    hour: "numeric",
    hour12: true,
    timeZone: "UTC",
  }).format(new Date(Date.UTC(2020, 0, 1, hour)));
}

const userTimezone = computed(() => authStore.user?.location || "Local");
const weekStart = computed(() => startOfCurrentWeek(userTimezone.value));
const weekEndExclusive = computed(() => {
  const date = new Date(weekStart.value);
  date.setUTCDate(weekStart.value.getUTCDate() + 7);
  return date;
});
const weekEndDisplay = computed(() => {
  const date = new Date(weekEndExclusive.value);
  date.setUTCDate(weekEndExclusive.value.getUTCDate() - 1);
  return date;
});

const weeklyParams = computed<SummaryRequestParams>(() => ({
  from: dateInputValueInTimezone(weekStart.value, userTimezone.value),
  to: dateInputValueInTimezone(weekEndExclusive.value, userTimezone.value),
}));

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
  () =>
    `${formatDate(weekStart.value, userTimezone.value)} - ${formatDate(
      weekEndDisplay.value,
      userTimezone.value,
    )}`,
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
      value: day ? formatWeekday(day.date, userTimezone.value) : "-",
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
  labels: timeline.value.map((day) =>
    formatWeekday(day.date, userTimezone.value),
  ),
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

watch(
  weeklyParams,
  (params) => {
    statsStore.fetchSummaryDetails(params).catch(() => undefined);
  },
  { immediate: true },
);
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
