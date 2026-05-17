<script setup lang="ts">
import { Line } from "vue-chartjs";
import {
  Chart as ChartJS,
  Title,
  Tooltip,
  Legend,
  LineElement,
  PointElement,
  CategoryScale,
  LinearScale,
  type ChartOptions,
  type ChartData,
} from "chart.js";

ChartJS.register(
  Title,
  Tooltip,
  Legend,
  LineElement,
  PointElement,
  CategoryScale,
  LinearScale,
);

defineProps<{
  data: ChartData<"line">;
  options?: ChartOptions<"line">;
}>();

const defaultOptions: ChartOptions<"line"> = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: false,
    },
  },
  scales: {
    y: {
      beginAtZero: true,
      grid: {
        color: "rgba(255, 255, 255, 0.1)",
      },
      ticks: {
        color: "#94a3b8",
      },
    },
    x: {
      grid: {
        display: false,
      },
      ticks: {
        color: "#94a3b8",
      },
    },
  },
  elements: {
    line: {
      tension: 0.3,
    },
  },
};
</script>

<template>
  <Line :data="data" :options="{ ...defaultOptions, ...options }" />
</template>
