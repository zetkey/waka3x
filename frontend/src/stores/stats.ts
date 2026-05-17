import { defineStore } from "pinia";
import { getApiErrorMessage, statsApi } from "@/lib/api";
import type {
  LeaderboardItem,
  LeaderboardResponse,
  LeaderboardRequestParams,
  ProjectStat,
  Summary,
  SummaryDetails,
  SummaryRequestParams,
} from "@/types/api";

export const useStatsStore = defineStore("stats", {
  state: () => ({
    summary: null as Summary | null,
    summaryDetails: null as SummaryDetails | null,
    projects: [] as ProjectStat[],
    leaderboard: [] as LeaderboardItem[],
    leaderboardDetails: null as LeaderboardResponse | null,
    loading: false,
    error: null as string | null,
  }),
  actions: {
    async fetchSummary(params: SummaryRequestParams = { interval: "today" }) {
      this.loading = true;
      this.error = null;
      try {
        const response = await statsApi.summary(params);
        this.summary = response;
      } catch (err: any) {
        this.error = getApiErrorMessage(err, "Failed to fetch summary");
        throw err;
      } finally {
        this.loading = false;
      }
    },
    async fetchSummaryDetails(
      params: SummaryRequestParams = { interval: "today" },
    ) {
      this.loading = true;
      this.error = null;
      try {
        const response = await statsApi.summaryDetails(params);
        this.summaryDetails = response;
        this.summary = response.summary;
      } catch (err: any) {
        this.error = getApiErrorMessage(err, "Failed to fetch summary");
        throw err;
      } finally {
        this.loading = false;
      }
    },
    async fetchProjects(query: string = "", page: number = 1) {
      this.loading = true;
      try {
        const response = await statsApi.projects({
          q: query,
          page,
          page_size: 24,
        });
        this.projects = response;
      } catch (err: any) {
        this.error = getApiErrorMessage(err, "Failed to fetch projects");
        throw err;
      } finally {
        this.loading = false;
      }
    },
    async fetchLeaderboard(params: LeaderboardRequestParams = {}) {
      this.loading = true;
      try {
        const response = await statsApi.leaderboard(params);
        this.leaderboardDetails = response;
        this.leaderboard = response.items;
      } catch (err: any) {
        this.error = getApiErrorMessage(err, "Failed to fetch leaderboard");
        throw err;
      } finally {
        this.loading = false;
      }
    },
  },
});
