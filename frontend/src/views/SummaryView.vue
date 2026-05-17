<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  AppWindow,
  Calendar,
  ChevronDown,
  Clock,
  Code,
  Bot,
  Download,
  Filter,
  Folder,
  Loader2,
} from "lucide-vue-next";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Alert, AlertDescription } from "@/components/ui/alert";
import BarChart from "@/components/charts/BarChart.vue";
import DoughnutChart from "@/components/charts/DoughnutChart.vue";
import { useStatsStore } from "@/stores/stats";
import { formatDuration, formatDate } from "@/lib/formatters";
import type {
  SummaryInterval,
  SummaryItem,
  SummaryRequestParams,
} from "@/types/api";

const statsStore = useStatsStore();
const route = useRoute();
const router = useRouter();

const intervals: Array<{ label: string; value: SummaryInterval }> = [
  { label: "Today", value: "today" },
  { label: "Yesterday", value: "yesterday" },
  { label: "This Week", value: "week" },
  { label: "Last 7 Days", value: "last_7_days" },
  { label: "Last 14 Days", value: "last_14_days" },
  { label: "Last 30 Days", value: "last_30_days" },
  { label: "This Month", value: "month" },
  { label: "This Year", value: "year" },
  { label: "All Time", value: "all_time" },
];

const currentInterval = ref<SummaryInterval>(
  (route.query.interval as SummaryInterval) || "last_7_days",
);
const customFrom = ref(route.query.from?.toString() || "");
const customTo = ref(route.query.to?.toString() || "");
const topN = ref(10);
const filters = reactive({
  project: route.query.project?.toString() || "",
  language: route.query.language?.toString() || "",
  machine: route.query.machine?.toString() || "",
  label: route.query.label?.toString() || "",
  category: route.query.category?.toString() || "",
});

const details = computed(() => statsStore.summaryDetails);
const summary = computed(() => details.value?.summary || statsStore.summary);
const totalSeconds = computed(
  () => summary.value?.projects.reduce((acc, p) => acc + p.total, 0) || 0,
);

function normalizeCategoryKey(key: string) {
  return key.toLowerCase().trim().replace(/\s+/g, " ");
}

function categoryTotal(key: string) {
  const wanted = normalizeCategoryKey(key);
  return (
    summary.value?.categories.reduce(
      (acc, category) =>
        normalizeCategoryKey(category.key) === wanted
          ? acc + category.total
          : acc,
      0,
    ) || 0
  );
}

const categoryAiCodingRatio = computed(() => {
  const aiCoding = categoryTotal("ai coding");
  const coding = categoryTotal("coding");
  const total = aiCoding + coding;
  if (!total) return 0;

  return Math.round((aiCoding / total) * 10000) / 10000;
});

const aiCodingRatio = computed(() => {
  const apiRatio = details.value?.ai_coding_ratio;
  if (typeof apiRatio === "number" && apiRatio > 0) return apiRatio;

  return categoryAiCodingRatio.value;
});

const aiCodingPercentage = computed(() => {
  const percentage = aiCodingRatio.value * 100;
  if (percentage > 0 && percentage < 1) return "< 1 %";
  return `${Math.round(percentage)} %`;
});

const kpis = computed(() => {
  const s = summary.value;
  if (!s) return [];
  return [
    {
      label: "Total Time",
      value: formatDuration(totalSeconds.value),
      icon: Clock,
      sub: `from ${formatDate(s.from)}`,
    },
    {
      label: "Active Projects",
      value: String(s.projects.length),
      icon: Folder,
    },
    { label: "Top Project", value: s.projects[0]?.key || "-", icon: Folder },
    { label: "Top Language", value: s.languages[0]?.key || "-", icon: Code },
    { label: "Top Editor", value: s.editors[0]?.key || "-", icon: AppWindow },
    { label: "AI coding ratio", value: aiCodingPercentage.value, icon: Bot },
  ];
});

function chartItems(items: SummaryItem[] = []) {
  return items.slice(0, topN.value);
}

const projectChartData = computed(() => ({
  labels: chartItems(summary.value?.projects).map((p) => p.key),
  datasets: [
    {
      label: "Hours",
      data: chartItems(summary.value?.projects).map(
        (p) => Math.round((p.total / 3600) * 10) / 10,
      ),
      backgroundColor: "#3b82f6",
      borderRadius: 4,
    },
  ],
}));

const languageChartData = computed(() => ({
  labels: chartItems(summary.value?.languages).map((l) => l.key),
  datasets: [
    {
      data: chartItems(summary.value?.languages).map((l) => l.total),
      backgroundColor: chartItems(summary.value?.languages).map(
        (l, index) =>
          details.value?.language_colors?.[l.key.toLowerCase()] ||
          ["#3b82f6", "#10b981", "#f59e0b", "#ef4444", "#8b5cf6"][index % 5],
      ),
      borderWidth: 0,
    },
  ],
}));

const editorChartData = computed(() => ({
  labels: chartItems(summary.value?.editors).map((e) => e.key),
  datasets: [
    {
      data: chartItems(summary.value?.editors).map((e) => e.total),
      backgroundColor: chartItems(summary.value?.editors).map(
        (e, index) =>
          details.value?.editor_colors?.[e.key.toLowerCase()] ||
          ["#6366f1", "#a855f7", "#ec4899", "#f43f5e"][index % 4],
      ),
      borderWidth: 0,
    },
  ],
}));

const timelineChartData = computed(() => ({
  labels: details.value?.timeline?.map((day) => formatDate(day.date)) || [],
  datasets: [
    {
      label: "Hours",
      data:
        details.value?.timeline?.map(
          (day) =>
            Math.round(
              (day.projects.reduce((acc, p) => acc + p.duration, 0) / 3600) *
                10,
            ) / 10,
        ) || [],
      borderColor: "#10b981",
      backgroundColor: "#10b981",
    },
  ],
}));

const hourlyChartData = computed(() => {
  const groups = details.value?.hourly_breakdown || [];
  const labels = groups.flatMap((group) =>
    group.items.map((item) =>
      new Date(item.from_time).toLocaleTimeString([], {
        hour: "2-digit",
        minute: "2-digit",
      }),
    ),
  );
  const data = groups.flatMap((group) =>
    group.items.map((item) => Math.round((item.duration / 60) * 10) / 10),
  );
  return {
    labels,
    datasets: [
      { label: "Minutes", data, backgroundColor: "#f59e0b", borderRadius: 4 },
    ],
  };
});

function buildParams(): SummaryRequestParams {
  const params: SummaryRequestParams = {};
  if (customFrom.value && customTo.value) {
    params.from = customFrom.value;
    params.to = customTo.value;
  } else {
    params.interval = currentInterval.value;
  }
  Object.entries(filters).forEach(([key, value]) => {
    if (value) (params as Record<string, string>)[key] = value;
  });
  return params;
}

function syncRouteAndFetch() {
  const params = buildParams();
  router.replace({ path: "/summary", query: params as Record<string, string> });
  statsStore.fetchSummaryDetails(params).catch(() => undefined);
}

function exportJson() {
  const payload = JSON.stringify(details.value || summary.value, null, 2);
  const blob = new Blob([payload], { type: "application/json" });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = "wakapi-summary.json";
  link.click();
  URL.revokeObjectURL(url);
}

watch(currentInterval, () => {
  customFrom.value = "";
  customTo.value = "";
  syncRouteAndFetch();
});

onMounted(syncRouteAndFetch);
</script>

<template>
  <div class="p-4 md:p-8 space-y-8">
    <div
      v-if="statsStore.loading && !summary"
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
            {{ details?.project_details ? details.project : "Summary" }}
          </h1>
          <p class="text-muted-foreground text-sm">
            Activity from {{ formatDate(summary?.from || "") }} to
            {{ formatDate(summary?.to || "") }}
          </p>
        </div>
        <div class="flex items-center gap-2 flex-wrap">
          <DropdownMenu>
            <DropdownMenuTrigger as-child>
              <Button variant="outline" size="sm" class="gap-2">
                <Calendar class="w-4 h-4" />
                {{
                  intervals.find((i) => i.value === currentInterval)?.label ||
                  "Custom"
                }}
                <ChevronDown class="w-3 h-3 opacity-50" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem
                v-for="interval in intervals"
                :key="interval.value"
                @click="currentInterval = interval.value"
              >
                {{ interval.label }}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
          <Button size="sm" class="gap-2" @click="syncRouteAndFetch"
            ><Filter class="w-4 h-4" /> Apply</Button
          >
          <Button variant="outline" size="sm" class="gap-2" @click="exportJson"
            ><Download class="w-4 h-4" /> Export</Button
          >
        </div>
      </div>

      <Alert v-if="details?.user_data_expiring" variant="destructive">
        <AlertDescription>
          Your oldest data is outside this server's
          {{ details.data_retention_months }} month retention window.
        </AlertDescription>
      </Alert>

      <Card>
        <CardHeader
          ><CardTitle class="text-primary text-sm"
            >Filters</CardTitle
          ></CardHeader
        >
        <CardContent
          class="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-6 gap-3"
        >
          <div class="space-y-2">
            <Label>Project</Label>
            <Select v-model="filters.project"
              ><SelectTrigger><SelectValue placeholder="Any" /></SelectTrigger
              ><SelectContent
                ><SelectItem value="">Any</SelectItem
                ><SelectItem
                  v-for="p in details?.available_filters.projects || []"
                  :key="p"
                  :value="p"
                  >{{ p }}</SelectItem
                ></SelectContent
              ></Select
            >
          </div>
          <div class="space-y-2">
            <Label>Language</Label>
            <Select v-model="filters.language"
              ><SelectTrigger><SelectValue placeholder="Any" /></SelectTrigger
              ><SelectContent
                ><SelectItem value="">Any</SelectItem
                ><SelectItem
                  v-for="p in details?.available_filters.languages || []"
                  :key="p"
                  :value="p"
                  >{{ p }}</SelectItem
                ></SelectContent
              ></Select
            >
          </div>
          <div class="space-y-2">
            <Label>Machine</Label>
            <Select v-model="filters.machine"
              ><SelectTrigger><SelectValue placeholder="Any" /></SelectTrigger
              ><SelectContent
                ><SelectItem value="">Any</SelectItem
                ><SelectItem
                  v-for="p in details?.available_filters.machines || []"
                  :key="p"
                  :value="p"
                  >{{ p }}</SelectItem
                ></SelectContent
              ></Select
            >
          </div>
          <div class="space-y-2">
            <Label>Label</Label>
            <Select v-model="filters.label"
              ><SelectTrigger><SelectValue placeholder="Any" /></SelectTrigger
              ><SelectContent
                ><SelectItem value="">Any</SelectItem
                ><SelectItem
                  v-for="p in details?.available_filters.labels || []"
                  :key="p"
                  :value="p"
                  >{{ p }}</SelectItem
                ></SelectContent
              ></Select
            >
          </div>
          <div class="space-y-2">
            <Label>Category</Label>
            <Select v-model="filters.category"
              ><SelectTrigger><SelectValue placeholder="Any" /></SelectTrigger
              ><SelectContent
                ><SelectItem value="">Any</SelectItem
                ><SelectItem
                  v-for="p in details?.available_filters.categories || []"
                  :key="p"
                  :value="p"
                  >{{ p }}</SelectItem
                ></SelectContent
              ></Select
            >
          </div>
          <div class="space-y-2">
            <Label>Top N</Label>
            <Input v-model.number="topN" type="number" min="1" max="50" />
          </div>
        </CardContent>
      </Card>

      <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
        <Card
          v-for="kpi in kpis"
          :key="kpi.label"
          class="border-accent/50 bg-card/50"
        >
          <CardHeader class="p-4 pb-1">
            <div class="flex items-center justify-between">
              <span
                class="text-[10px] font-bold uppercase tracking-widest text-muted-foreground"
                >{{ kpi.label }}</span
              >
              <component :is="kpi.icon" class="w-3 h-3 text-primary" />
            </div>
          </CardHeader>
          <CardContent class="p-4 pt-0">
            <div class="text-xl font-bold text-primary truncate">
              {{ kpi.value }}
            </div>
            <p v-if="kpi.sub" class="text-[10px] text-muted-foreground mt-1">
              {{ kpi.sub }}
            </p>
          </CardContent>
        </Card>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card class="md:col-span-2 md:row-span-2">
          <CardHeader
            ><CardTitle class="text-primary">Projects</CardTitle></CardHeader
          >
          <CardContent class="h-[400px]">
            <BarChart
              v-if="summary?.projects.length"
              :data="projectChartData"
            />
            <div
              v-else
              class="h-full flex items-center justify-center text-muted-foreground italic text-sm"
            >
              No project data available
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader
            ><CardTitle class="text-primary text-sm"
              >Languages</CardTitle
            ></CardHeader
          >
          <CardContent class="h-[250px]">
            <DoughnutChart
              v-if="summary?.languages.length"
              :data="languageChartData"
            />
            <div
              v-else
              class="h-full flex items-center justify-center text-muted-foreground italic text-xs"
            >
              No language data
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader
            ><CardTitle class="text-primary text-sm"
              >Editors</CardTitle
            ></CardHeader
          >
          <CardContent class="h-[250px]">
            <DoughnutChart
              v-if="summary?.editors.length"
              :data="editorChartData"
            />
            <div
              v-else
              class="h-full flex items-center justify-center text-muted-foreground italic text-xs"
            >
              No editor data
            </div>
          </CardContent>
        </Card>

        <Card v-if="details?.timeline?.length" class="md:col-span-3">
          <CardHeader
            ><CardTitle class="text-primary">Timeline</CardTitle></CardHeader
          >
          <CardContent class="h-[280px]"
            ><BarChart :data="timelineChartData"
          /></CardContent>
        </Card>

        <Card v-if="details?.hourly_breakdown?.length" class="md:col-span-3">
          <CardHeader
            ><CardTitle class="text-primary"
              >Hourly Breakdown</CardTitle
            ></CardHeader
          >
          <CardContent class="h-[280px]"
            ><BarChart :data="hourlyChartData"
          /></CardContent>
        </Card>

        <Card class="md:col-span-3">
          <CardHeader
            ><CardTitle class="text-primary"
              >Environment Breakdown</CardTitle
            ></CardHeader
          >
          <CardContent>
            <div class="grid grid-cols-1 md:grid-cols-4 gap-8 m-4">
              <div
                v-for="group in [
                  { title: 'Machines', items: summary?.machines || [] },
                  {
                    title: 'Operating Systems',
                    items: summary?.operating_systems || [],
                  },
                  { title: 'Branches', items: summary?.branches || [] },
                  { title: 'Categories', items: summary?.categories || [] },
                ]"
                :key="group.title"
                class="space-y-2"
              >
                <h4
                  class="text-xs font-bold uppercase text-muted-foreground mb-4"
                >
                  {{ group.title }}
                </h4>
                <div
                  v-for="item in group.items.slice(0, topN)"
                  :key="item.key"
                  class="flex justify-between border-b border-white/5 py-1 gap-3"
                >
                  <span class="text-sm truncate">{{ item.key }}</span>
                  <span class="text-sm text-muted-foreground font-mono">{{
                    formatDuration(item.total)
                  }}</span>
                </div>
                <div
                  v-if="!group.items.length"
                  class="text-muted-foreground italic text-sm"
                >
                  No data
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  </div>
</template>
