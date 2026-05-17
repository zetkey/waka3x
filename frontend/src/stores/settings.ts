import { defineStore } from "pinia";
import { getApiErrorMessage, settingsApi } from "@/lib/api";
import type {
  AddAliasRequest,
  AddApiKeyRequest,
  AddLabelRequest,
  AddLanguageMappingRequest,
  ChangePasswordRequest,
  ChangeUserIdRequest,
  DeleteAliasRequest,
  DeleteApiKeyRequest,
  DeleteLabelRequest,
  DeleteLanguageMappingRequest,
  ImportWakatimeRequest,
  MessageResponse,
  SettingsResponse,
  UpdateHeartbeatsTimeoutRequest,
  UpdateLeaderboardRequest,
  UpdateReadmeStatsBaseUrlRequest,
  UpdateSharingRequest,
  UpdateUnknownProjectsRequest,
  UpdateWakatimeCredentialsRequest,
  WebAuthnDeleteRequest,
} from "@/types/api";

export const useSettingsStore = defineStore("settings", {
  state: () => ({
    settings: null as SettingsResponse | null,
    loading: false,
    saving: false,
    error: null as string | null,
  }),
  actions: {
    async fetchSettings() {
      this.loading = true;
      this.error = null;
      try {
        this.settings = await settingsApi.get();
        return this.settings;
      } catch (err) {
        this.error = getApiErrorMessage(err, "Failed to load settings");
        throw err;
      } finally {
        this.loading = false;
      }
    },

    async runAction<T = MessageResponse>(
      action: () => Promise<T>,
      refresh = true,
    ) {
      this.saving = true;
      this.error = null;
      try {
        const response = await action();
        if (refresh) await this.fetchSettings();
        return response;
      } catch (err) {
        this.error = getApiErrorMessage(err, "Action failed");
        throw err;
      } finally {
        this.saving = false;
      }
    },

    changePassword(payload: ChangePasswordRequest) {
      return this.runAction(() => settingsApi.changePassword(payload), false);
    },

    changeUserId(payload: ChangeUserIdRequest) {
      return this.runAction(() => settingsApi.changeUserId(payload), false);
    },

    resetApiKey() {
      return this.runAction(() => settingsApi.resetApiKey());
    },

    addApiKey(payload: AddApiKeyRequest) {
      return this.runAction(() => settingsApi.addApiKey(payload));
    },

    deleteApiKey(payload: DeleteApiKeyRequest) {
      return this.runAction(() => settingsApi.deleteApiKey(payload));
    },

    generateInvite() {
      return this.runAction(() => settingsApi.generateInvite(), false);
    },

    updateUnknownProjects(payload: UpdateUnknownProjectsRequest) {
      return this.runAction(() => settingsApi.updateUnknownProjects(payload));
    },

    updateHeartbeatsTimeout(payload: UpdateHeartbeatsTimeoutRequest) {
      return this.runAction(() => settingsApi.updateHeartbeatsTimeout(payload));
    },

    updateReadmeStatsBaseUrl(payload: UpdateReadmeStatsBaseUrlRequest) {
      return this.runAction(() =>
        settingsApi.updateReadmeStatsBaseUrl(payload),
      );
    },

    updateLeaderboard(payload: UpdateLeaderboardRequest) {
      return this.runAction(() => settingsApi.updateLeaderboard(payload));
    },

    updateSharing(payload: UpdateSharingRequest) {
      return this.runAction(() => settingsApi.updateSharing(payload));
    },

    addAlias(payload: AddAliasRequest) {
      return this.runAction(() => settingsApi.addAlias(payload));
    },

    deleteAlias(payload: DeleteAliasRequest) {
      return this.runAction(() => settingsApi.deleteAlias(payload));
    },

    addLabel(payload: AddLabelRequest) {
      return this.runAction(() => settingsApi.addLabel(payload));
    },

    deleteLabel(payload: DeleteLabelRequest) {
      return this.runAction(() => settingsApi.deleteLabel(payload));
    },

    addLanguageMapping(payload: AddLanguageMappingRequest) {
      return this.runAction(() => settingsApi.addLanguageMapping(payload));
    },

    deleteLanguageMapping(payload: DeleteLanguageMappingRequest) {
      return this.runAction(() => settingsApi.deleteLanguageMapping(payload));
    },

    updateWakatime(payload: UpdateWakatimeCredentialsRequest) {
      return this.runAction(() => settingsApi.updateWakatime(payload));
    },

    importWakatime(payload: ImportWakatimeRequest) {
      return this.runAction(() => settingsApi.importWakatime(payload), false);
    },

    regenerateSummaries() {
      return this.runAction(() => settingsApi.regenerateSummaries(), false);
    },

    clearData() {
      return this.runAction(() => settingsApi.clearData(), false);
    },

    deleteAccount() {
      return this.runAction(() => settingsApi.deleteAccount(), false);
    },

    deleteWebAuthn(payload: WebAuthnDeleteRequest) {
      return this.runAction(() => settingsApi.webAuthnDelete(payload));
    },
  },
});
