import axios, { AxiosError, type AxiosRequestConfig } from "axios";
import type {
  AddAliasRequest,
  AddApiKeyRequest,
  AddLabelRequest,
  AddLanguageMappingRequest,
  ApiErrorResponse,
  AuthUserEnvelope,
  BootstrapConfig,
  CaptchaResponse,
  ChangePasswordRequest,
  ChangeUserIdRequest,
  CurrentUser,
  DeleteAliasRequest,
  DeleteApiKeyRequest,
  DeleteLabelRequest,
  DeleteLanguageMappingRequest,
  GenerateInviteResponse,
  HomeStats,
  ImprintResponse,
  ImportWakatimeRequest,
  LeaderboardRequestParams,
  LeaderboardResponse,
  LoginRequest,
  MessageResponse,
  ProjectListRequestParams,
  ProjectStat,
  ResetPasswordRequest,
  SetPasswordRequest,
  SettingsResponse,
  SetupResponse,
  SignupRequest,
  Summary,
  SummaryDetails,
  SummaryRequestParams,
  UnsubscribeRequest,
  UpdateCurrentUserRequest,
  UpdateHeartbeatsTimeoutRequest,
  UpdateLeaderboardRequest,
  UpdateReadmeStatsBaseUrlRequest,
  UpdateSharingRequest,
  UpdateUnknownProjectsRequest,
  UpdateWakatimeCredentialsRequest,
  WebAuthnAddRequest,
  WebAuthnDeleteRequest,
  WebAuthnLoginRequest,
} from "@/types/api";

export const http = axios.create({
  withCredentials: true,
});

async function request<T>(config: AxiosRequestConfig): Promise<T> {
  const response = await http.request<T>(config);
  return response.data;
}

export function getApiErrorMessage(error: unknown, fallback: string): string {
  if (error instanceof AxiosError) {
    return (
      (error.response?.data as ApiErrorResponse | undefined)?.error || fallback
    );
  }
  return fallback;
}

export const authApi = {
  login(payload: LoginRequest) {
    return request<AuthUserEnvelope>({
      method: "POST",
      url: "/api/login",
      data: payload,
    });
  },

  signup(payload: SignupRequest) {
    return request<AuthUserEnvelope>({
      method: "POST",
      url: "/api/signup",
      data: payload,
    });
  },

  logout() {
    return request<void>({
      method: "POST",
      url: "/api/logout",
    });
  },

  resetPassword(payload: ResetPasswordRequest) {
    return request<MessageResponse>({
      method: "POST",
      url: "/api/password/reset",
      data: payload,
    });
  },

  setPassword(payload: SetPasswordRequest) {
    return request<MessageResponse>({
      method: "POST",
      url: "/api/password/set",
      data: payload,
    });
  },

  webAuthnOptions() {
    return request<unknown>({
      method: "GET",
      url: "/api/webauthn/options",
    });
  },

  webAuthnLogin(payload: WebAuthnLoginRequest) {
    return request<AuthUserEnvelope>({
      method: "POST",
      url: "/api/webauthn/login",
      data: payload,
    });
  },

  captcha() {
    return request<CaptchaResponse>({
      method: "GET",
      url: "/api/captcha/new",
    });
  },

  currentUser() {
    return request<CurrentUser>({
      method: "GET",
      url: "/api/users/current",
    });
  },

  updateCurrentUser(payload: UpdateCurrentUserRequest) {
    return request<CurrentUser>({
      method: "PUT",
      url: "/api/users/current",
      data: payload,
    });
  },
};

export const statsApi = {
  summary(params: SummaryRequestParams = { interval: "today" }) {
    return request<Summary>({
      method: "GET",
      url: "/api/summary",
      params,
    });
  },

  summaryDetails(params: SummaryRequestParams = { interval: "today" }) {
    return request<SummaryDetails>({
      method: "GET",
      url: "/api/summary/details",
      params,
    });
  },

  projects(params: ProjectListRequestParams = {}) {
    return request<ProjectStat[]>({
      method: "GET",
      url: "/api/projects",
      params,
    });
  },

  leaderboard(params: LeaderboardRequestParams = {}) {
    return request<LeaderboardResponse>({
      method: "GET",
      url: "/api/leaderboard",
      params,
    });
  },
};

export const metaApi = {
  config() {
    return request<BootstrapConfig>({
      method: "GET",
      url: "/api/config",
    });
  },

  home() {
    return request<HomeStats>({
      method: "GET",
      url: "/api/home",
    });
  },

  setup() {
    return request<SetupResponse>({
      method: "GET",
      url: "/api/setup",
    });
  },

  imprint() {
    return request<ImprintResponse>({
      method: "GET",
      url: "/api/imprint",
    });
  },

  unsubscribe(payload: UnsubscribeRequest) {
    return request<MessageResponse>({
      method: "POST",
      url: "/api/unsubscribe",
      data: payload,
    });
  },
};

export const settingsApi = {
  get() {
    return request<SettingsResponse>({
      method: "GET",
      url: "/api/settings",
    });
  },

  changePassword(payload: ChangePasswordRequest) {
    return request<MessageResponse>({
      method: "POST",
      url: "/api/settings/password",
      data: payload,
    });
  },

  changeUserId(payload: ChangeUserIdRequest) {
    return request<MessageResponse>({
      method: "POST",
      url: "/api/settings/username",
      data: payload,
    });
  },

  resetApiKey() {
    return request<MessageResponse>({
      method: "POST",
      url: "/api/settings/api-key/reset",
    });
  },

  addApiKey(payload: AddApiKeyRequest) {
    return request<{ message: string; api_key: string }>({
      method: "POST",
      url: "/api/settings/api-keys",
      data: payload,
    });
  },

  deleteApiKey(payload: DeleteApiKeyRequest) {
    return request<MessageResponse>({
      method: "DELETE",
      url: "/api/settings/api-keys",
      data: payload,
    });
  },

  generateInvite() {
    return request<GenerateInviteResponse>({
      method: "POST",
      url: "/api/settings/invite",
    });
  },

  updateUnknownProjects(payload: UpdateUnknownProjectsRequest) {
    return request<MessageResponse>({
      method: "PUT",
      url: "/api/settings/unknown-projects",
      data: payload,
    });
  },

  updateHeartbeatsTimeout(payload: UpdateHeartbeatsTimeoutRequest) {
    return request<MessageResponse>({
      method: "PUT",
      url: "/api/settings/heartbeats-timeout",
      data: payload,
    });
  },

  updateReadmeStatsBaseUrl(payload: UpdateReadmeStatsBaseUrlRequest) {
    return request<MessageResponse>({
      method: "PUT",
      url: "/api/settings/readme-stats-base-url",
      data: payload,
    });
  },

  updateLeaderboard(payload: UpdateLeaderboardRequest) {
    return request<MessageResponse>({
      method: "PUT",
      url: "/api/settings/leaderboard",
      data: payload,
    });
  },

  updateSharing(payload: UpdateSharingRequest) {
    return request<MessageResponse>({
      method: "PUT",
      url: "/api/settings/sharing",
      data: payload,
    });
  },

  addAlias(payload: AddAliasRequest) {
    return request<MessageResponse>({
      method: "POST",
      url: "/api/settings/aliases",
      data: payload,
    });
  },

  deleteAlias(payload: DeleteAliasRequest) {
    return request<MessageResponse>({
      method: "DELETE",
      url: "/api/settings/aliases",
      data: payload,
    });
  },

  addLabel(payload: AddLabelRequest) {
    return request<MessageResponse>({
      method: "POST",
      url: "/api/settings/labels",
      data: payload,
    });
  },

  deleteLabel(payload: DeleteLabelRequest) {
    return request<MessageResponse>({
      method: "DELETE",
      url: "/api/settings/labels",
      data: payload,
    });
  },

  addLanguageMapping(payload: AddLanguageMappingRequest) {
    return request<MessageResponse>({
      method: "POST",
      url: "/api/settings/language-mappings",
      data: payload,
    });
  },

  deleteLanguageMapping(payload: DeleteLanguageMappingRequest) {
    return request<MessageResponse>({
      method: "DELETE",
      url: "/api/settings/language-mappings",
      data: payload,
    });
  },

  updateWakatime(payload: UpdateWakatimeCredentialsRequest) {
    return request<MessageResponse>({
      method: "PUT",
      url: "/api/settings/wakatime",
      data: payload,
    });
  },

  importWakatime(payload: ImportWakatimeRequest) {
    return request<MessageResponse>({
      method: "POST",
      url: "/api/settings/wakatime/import",
      data: payload,
    });
  },

  regenerateSummaries() {
    return request<MessageResponse>({
      method: "POST",
      url: "/api/settings/summaries/regenerate",
    });
  },

  clearData() {
    return request<MessageResponse>({
      method: "POST",
      url: "/api/settings/data/clear",
    });
  },

  deleteAccount() {
    return request<MessageResponse>({
      method: "DELETE",
      url: "/api/settings/account",
    });
  },

  webAuthnOptions() {
    return request<unknown>({
      method: "GET",
      url: "/api/settings/webauthn/options",
    });
  },

  webAuthnAdd(payload: WebAuthnAddRequest) {
    return request<MessageResponse>({
      method: "POST",
      url: "/api/settings/webauthn",
      data: payload,
    });
  },

  webAuthnDelete(payload: WebAuthnDeleteRequest) {
    return request<MessageResponse>({
      method: "DELETE",
      url: "/api/settings/webauthn",
      data: payload,
    });
  },
};
