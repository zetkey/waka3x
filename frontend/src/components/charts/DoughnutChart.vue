<script setup lang="ts">
import { Doughnut } from "vue-chartjs";
import {
  Chart as ChartJS,
  ArcElement,
  Tooltip,
  Legend,
  type ChartOptions,
  type ChartData,
} from "chart.js";

ChartJS.register(ArcElement, Tooltip, Legend);

defineProps<{
  data: ChartData<"doughnut">;
  options?: ChartOptions<"doughnut">;
}>();

const defaultOptions: ChartOptions<"doughnut"> = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: "bottom",
      labels: {
        color: "#94a3b8",
        padding: 20,
        usePointStyle: true,
        pointStyle: "circle",
      },
    },
  },
  cutout: "70%",
};
</script>

<template>
  <Doughnut :data="data" :options="{ ...defaultOptions, ...options }" />
</template>
