<script setup lang="ts">
import { Bar } from "vue-chartjs";
import {
  Chart as ChartJS,
  Title,
  Tooltip,
  Legend,
  BarElement,
  CategoryScale,
  LinearScale,
  type ChartOptions,
  type ChartData,
} from "chart.js";

ChartJS.register(
  Title,
  Tooltip,
  Legend,
  BarElement,
  CategoryScale,
  LinearScale,
);

defineProps<{
  data: ChartData<"bar">;
  options?: ChartOptions<"bar">;
}>();

const defaultOptions: ChartOptions<"bar"> = {
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
};
</script>

<template>
  <Bar :data="data" :options="{ ...defaultOptions, ...options }" />
</template>
