import { defineStore } from "pinia";
import { authApi, getApiErrorMessage } from "@/lib/api";
import type {
  CurrentUser,
  SignupRequest,
  UpdateCurrentUserRequest,
} from "@/types/api";

export const useAuthStore = defineStore("auth", {
  state: () => ({
    user: null as CurrentUser | null,
    isAuthenticated: false,
    loading: false,
    error: null as string | null,
  }),
  actions: {
    async login(username: string, password: string) {
      this.loading = true;
      this.error = null;
      try {
        const response = await authApi.login({ username, password });
        this.user = response.user;
        this.isAuthenticated = true;
      } catch (err: any) {
        this.error = getApiErrorMessage(err, "Login failed");
        throw err;
      } finally {
        this.loading = false;
      }
    },
    async signup(data: SignupRequest) {
      this.loading = true;
      this.error = null;
      try {
        const response = await authApi.signup(data);
        this.user = response.user;
        this.isAuthenticated = true;
      } catch (err: any) {
        this.error = getApiErrorMessage(err, "Signup failed");
        throw err;
      } finally {
        this.loading = false;
      }
    },
    async logout() {
      try {
        await authApi.logout();
      } finally {
        this.user = null;
        this.isAuthenticated = false;
      }
    },
    async fetchUser() {
      try {
        const response = await authApi.currentUser();
        this.user = response;
        this.isAuthenticated = true;
      } catch {
        this.user = null;
        this.isAuthenticated = false;
      }
    },
    async updateProfile(data: UpdateCurrentUserRequest) {
      this.loading = true;
      try {
        const response = await authApi.updateCurrentUser(data);
        this.user = response;
      } catch (err: any) {
        this.error = getApiErrorMessage(err, "Failed to update profile");
        throw err;
      } finally {
        this.loading = false;
      }
    },
  },
});
