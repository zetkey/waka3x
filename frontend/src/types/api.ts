export type DurationSeconds = number;
export type ISODateTime = string;

export interface ApiErrorResponse {
  error: string;
}

export interface MessageResponse {
  message: string;
  user?: CurrentUser;
}

export interface OidcProvider {
  name: string;
  display_name: string;
  login_url: string;
}

export interface BootstrapConfig {
  version: string;
  base_path: string;
  public_url: string;
  db_type: string;
  leaderboard_enabled: boolean;
  leaderboard_require_auth: boolean;
  allow_signup: boolean;
  invite_codes_enabled: boolean;
  signup_captcha: boolean;
  disable_local_auth: boolean;
  disable_webauthn: boolean;
  subscriptions_enabled: boolean;
  subscription_price: string;
  stripe_api_key: string;
  support_contact: string;
  data_retention_months: number;
  default_wakatime_url: string;
  avatar_url_template: string;
  oidc_providers: OidcProvider[];
  mail_enabled: boolean;
  import_enabled: boolean;
  import_backoff_min: number;
  import_max_rate_hours: number;
}

export interface Newsbox {
  enabled?: boolean;
  title?: string;
  text?: string;
  html?: string;
  link?: string;
  [key: string]: unknown;
}

export interface HomeStats {
  total_hours: number;
  total_users: number;
  currently_online: number;
  newsbox?: Newsbox | null;
}

export interface ImprintResponse {
  html: string;
}

export interface SetupResponse {
  api_key: string;
  base_url: string;
  public_url: string;
  username?: string;
  authenticated: boolean;
}

export interface CaptchaResponse {
  id: string;
  image_url: string;
}

export interface UnsubscribeRequest {
  token: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface SignupRequest {
  username: string;
  email?: string;
  password: string;
  password_repeat?: string;
  location?: string;
  captcha_id?: string;
  captcha?: string;
  invite_code?: string;
}

export interface SetPasswordRequest {
  password: string;
  password_repeat: string;
  token: string;
}

export interface ResetPasswordRequest {
  email: string;
}

export interface WebAuthnLoginRequest {
  assertion_json: string;
}

export interface CurrentUser {
  id: string;
  api_key: string;
  email: string;
  location: string;
  start_of_week: number;
  created_at: ISODateTime;
  last_logged_in_at: ISODateTime;
  auth_type: string;
  has_data: boolean;
  reports_weekly: boolean;
  public_leaderboard: boolean;
  exclude_unknown_projects: boolean;
  heartbeats_timeout_min: number;
  share_data_max_days: number;
  share_editors: boolean;
  share_languages: boolean;
  share_projects: boolean;
  share_oss: boolean;
  share_machines: boolean;
  share_labels: boolean;
  share_activity_chart: boolean;
  has_active_subscription: boolean;
  avatar_url: string;
  wakatime_connected: boolean;
  wakatime_api_url: string;
  readme_stats_base_url: string;
}

export interface PublicUser {
  id: string;
  avatar_url: string;
  has_active_subscription: boolean;
}

export interface AuthUserEnvelope {
  user: CurrentUser;
}

export interface UpdateCurrentUserRequest {
  email: string;
  location: string;
  start_of_week: number;
  reports_weekly: boolean;
  public_leaderboard?: boolean;
}

export interface SummaryItem {
  key: string;
  total: DurationSeconds;
}

export interface Summary {
  user_id: string;
  from: ISODateTime;
  to: ISODateTime;
  projects: SummaryItem[];
  languages: SummaryItem[];
  editors: SummaryItem[];
  operating_systems: SummaryItem[];
  machines: SummaryItem[];
  labels: SummaryItem[];
  branches: SummaryItem[];
  entities: SummaryItem[];
  categories: SummaryItem[];
}

export type SummaryInterval =
  | "today"
  | "yesterday"
  | "week"
  | "last_week"
  | "month"
  | "last_month"
  | "year"
  | "7_days"
  | "last_7_days"
  | "14_days"
  | "last_14_days"
  | "30_days"
  | "last_30_days"
  | "6_months"
  | "last_6_months"
  | "12_months"
  | "last_12_months"
  | "last_year"
  | "any"
  | "all_time";

export interface SummaryRequestParams {
  interval?: SummaryInterval;
  from?: string;
  to?: string;
  recompute?: boolean;
  project?: string;
  language?: string;
  editor?: string;
  operating_system?: string;
  machine?: string;
  label?: string;
  branch?: string;
  entity?: string;
  category?: string;
}

export interface AvailableFilters {
  projects: string[];
  languages: string[];
  machines: string[];
  labels: string[];
  categories: string[];
}

export interface TimelineItem {
  name: string;
  duration: DurationSeconds;
}

export interface TimelineDay {
  date: ISODateTime;
  projects: TimelineItem[];
}

export interface HourlyBreakdownItem {
  from_time: ISODateTime;
  duration: DurationSeconds;
  entity: string;
}

export interface HourlyBreakdownProject {
  project: string;
  items: HourlyBreakdownItem[];
}

export interface HourlyActivity {
  hour: number;
  duration: DurationSeconds;
}

export interface SummaryDetails {
  summary: Summary;
  available_filters: AvailableFilters;
  editor_colors: Record<string, string>;
  language_colors: Record<string, string>;
  os_colors: Record<string, string>;
  ai_coding_ratio: number;
  timeline: TimelineDay[] | null;
  hourly_breakdown: HourlyBreakdownProject[];
  hourly_breakdown_from: ISODateTime;
  hourly_activity: HourlyActivity[];
  user_first_data: ISODateTime;
  data_retention_months: number;
  user_data_expiring: boolean;
  project_details: boolean;
  project: string;
}

export interface ProjectListRequestParams {
  q?: string;
  page?: number;
  page_size?: number;
}

export interface ProjectStat {
  project: string;
  total_heartbeats: number;
  top_language: string;
  first_heartbeat: ISODateTime;
  last_heartbeat: ISODateTime;
}

export interface LeaderboardRequestParams {
  by?: "language";
  key?: string;
  page?: number;
  page_size?: number;
}

export interface LeaderboardItem {
  rank: number;
  user_id: string;
  interval: string;
  aggregated_by?: number;
  key?: string;
  total: DurationSeconds;
  updated_at: ISODateTime;
  user?: PublicUser;
}

export interface LeaderboardResponse {
  items: LeaderboardItem[];
  by: string;
  key: string;
  top_keys: string[];
  user_languages: Record<string, string[]> | null;
  interval_label: string;
  last_updated: ISODateTime;
}

export interface ChangePasswordRequest {
  password_old: string;
  password_new: string;
  password_repeat: string;
}

export interface ChangeUserIdRequest {
  new_userid: string;
}

export interface UpdateUnknownProjectsRequest {
  exclude_unknown_projects: boolean;
}

export interface UpdateHeartbeatsTimeoutRequest {
  heartbeats_timeout: number;
}

export interface UpdateReadmeStatsBaseUrlRequest {
  readme_stats_base_url: string;
}

export interface UpdateLeaderboardRequest {
  enable_leaderboard: boolean;
}

export interface UpdateSharingRequest {
  max_days: number;
  share_projects: boolean;
  share_languages: boolean;
  share_editors: boolean;
  share_oss: boolean;
  share_machines: boolean;
  share_labels: boolean;
  share_activity_chart: boolean;
}

export interface AddAliasRequest {
  type: number;
  key: string;
  value: string;
}

export interface DeleteAliasRequest {
  type: number;
  key: string;
}

export interface AddLabelRequest {
  key: string[];
  value: string;
}

export interface DeleteLabelRequest {
  key: string;
  value: string;
}

export interface LanguageMapping {
  ID?: number;
  id?: number;
  user_id?: string;
  extension: string;
  language: string;
}

export interface AddLanguageMappingRequest {
  extension: string;
  language: string;
}

export interface DeleteLanguageMappingRequest {
  mapping_id: number;
}

export interface UpdateWakatimeCredentialsRequest {
  api_url?: string;
  api_key: string;
}

export interface ImportWakatimeRequest {
  use_legacy_importer: boolean;
}

export interface AddApiKeyRequest {
  api_name: string;
  api_readonly: boolean;
}

export interface DeleteApiKeyRequest {
  api_key_value: string;
}

export interface WebAuthnAddRequest {
  authenticator_name: string;
  credential_json: string;
}

export interface WebAuthnDeleteRequest {
  credential_name: string;
}

export interface GenerateInviteResponse {
  message: string;
  invite_link: string;
}

export interface CombinedAlias {
  key: string;
  type: number;
  values: string[];
}

export interface CombinedLabel {
  key: string;
  values: string[];
}

export interface SettingsApiKey {
  name: string;
  value: string;
  read_only: boolean;
  main: boolean;
}

export interface WebAuthnCredential {
  name: string;
}

export interface SettingsResponse {
  user: CurrentUser;
  aliases: CombinedAlias[];
  labels: CombinedLabel[];
  projects: string[];
  language_mappings: LanguageMapping[];
  api_keys: SettingsApiKey[];
  webauthn_credentials: WebAuthnCredential[];
  user_first_data: ISODateTime;
  subscription_price: string;
  subscriptions_enabled: boolean;
  support_contact: string;
  data_retention_months: number;
  invite_link?: string;
  readme_card_custom_title: string;
  disable_webauthn: boolean;
  default_wakatime_url: string;
}
