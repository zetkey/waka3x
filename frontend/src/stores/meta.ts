import { defineStore } from "pinia";
import { getApiErrorMessage, metaApi } from "@/lib/api";
import type {
  BootstrapConfig,
  HomeStats,
  ImprintResponse,
  SetupResponse,
} from "@/types/api";

export const useMetaStore = defineStore("meta", {
  state: () => ({
    config: null as BootstrapConfig | null,
    home: null as HomeStats | null,
    setup: null as SetupResponse | null,
    imprint: null as ImprintResponse | null,
    loading: false,
    error: null as string | null,
  }),
  actions: {
    async fetchConfig() {
      if (this.config) return this.config;
      this.loading = true;
      this.error = null;
      try {
        this.config = await metaApi.config();
        return this.config;
      } catch (err) {
        this.error = getApiErrorMessage(err, "Failed to load server config");
        throw err;
      } finally {
        this.loading = false;
      }
    },

    async fetchHome() {
      this.loading = true;
      this.error = null;
      try {
        this.home = await metaApi.home();
        return this.home;
      } catch (err) {
        this.error = getApiErrorMessage(err, "Failed to load home stats");
        throw err;
      } finally {
        this.loading = false;
      }
    },

    async fetchSetup() {
      this.loading = true;
      this.error = null;
      try {
        this.setup = await metaApi.setup();
        return this.setup;
      } catch (err) {
        this.error = getApiErrorMessage(err, "Failed to load setup data");
        throw err;
      } finally {
        this.loading = false;
      }
    },

    async fetchImprint() {
      this.loading = true;
      this.error = null;
      try {
        this.imprint = await metaApi.imprint();
        return this.imprint;
      } catch (err) {
        this.error = getApiErrorMessage(err, "Failed to load imprint");
        throw err;
      } finally {
        this.loading = false;
      }
    },
  },
});
