<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import {
  AlertTriangle,
  Copy,
  Database,
  ExternalLink,
  Image,
  Key,
  Loader2,
  Plug,
  Plus,
  ShieldCheck,
  Trash2,
  User,
} from "lucide-vue-next";
import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useAuthStore } from "@/stores/auth";
import { useMetaStore } from "@/stores/meta";
import { useSettingsStore } from "@/stores/settings";
import { toast } from "vue-sonner";
import { getApiErrorMessage } from "@/lib/api";
import { createPasskeyCredential } from "@/lib/webauthn";
import { settingsApi } from "@/lib/api";

const authStore = useAuthStore();
const metaStore = useMetaStore();
const settingsStore = useSettingsStore();
const activeTab = ref("account");
const browserTimezone =
  Intl.DateTimeFormat().resolvedOptions().timeZone || "Local";
const settings = computed(() => settingsStore.settings);
const githubReadmeStatsDefaultUrl = "https://github-readme-stats.vercel.app";
const supportedTimezones =
  (
    Intl as typeof Intl & {
      supportedValuesOf?: (key: "timeZone") => string[];
    }
  ).supportedValuesOf?.("timeZone") || [];

const weekDays = [
  { label: "Sunday", value: 0 },
  { label: "Monday", value: 1 },
  { label: "Tuesday", value: 2 },
  { label: "Wednesday", value: 3 },
  { label: "Thursday", value: 4 },
  { label: "Friday", value: 5 },
  { label: "Saturday", value: 6 },
];

type SharingToggleKey =
  | "share_projects"
  | "share_languages"
  | "share_editors"
  | "share_oss"
  | "share_machines"
  | "share_labels"
  | "share_activity_chart";

const sharingOptions: Array<{ key: SharingToggleKey; label: string }> = [
  { key: "share_projects", label: "Share projects" },
  { key: "share_languages", label: "Share languages" },
  { key: "share_editors", label: "Share editors" },
  { key: "share_oss", label: "Share operating systems" },
  { key: "share_machines", label: "Share machines" },
  { key: "share_labels", label: "Share project labels" },
  { key: "share_activity_chart", label: "Share activity chart" },
];

const profileForm = reactive({
  email: "",
  location: browserTimezone,
  reports_weekly: false,
  start_of_week: 1,
});
const timezoneOptions = computed(() => {
  return Array.from(
    new Set([
      browserTimezone,
      "Local",
      profileForm.location,
      ...supportedTimezones,
    ]),
  )
    .filter(Boolean)
    .sort((a, b) => {
      if (a === browserTimezone) return -1;
      if (b === browserTimezone) return 1;
      if (a === "Local") return -1;
      if (b === "Local") return 1;
      return a.localeCompare(b);
    });
});
const passwordForm = reactive({
  password_old: "",
  password_new: "",
  password_repeat: "",
});
const usernameForm = reactive({ new_userid: "" });
const dataForm = reactive({
  exclude_unknown_projects: false,
  heartbeats_timeout: 2,
  readme_stats_base_url: "",
});
const sharingForm = reactive({
  enable_leaderboard: false,
  max_days: 0,
  share_projects: false,
  share_languages: false,
  share_editors: false,
  share_oss: false,
  share_machines: false,
  share_labels: false,
  share_activity_chart: false,
});
const wakatimeForm = reactive({
  api_url: "",
  api_key: "",
  use_legacy_importer: false,
});
const newApiKey = reactive({ api_name: "", api_readonly: false });
const newAlias = reactive({ type: 0, key: "", value: "" });
const newLabel = reactive({ key: "", value: "" });
const newMapping = reactive({ extension: "", language: "" });
const newPasskeyName = ref("");
const inviteLink = ref("");
const vibrantColorsEnabled = ref(false);

function trimTrailingSlashes(value: string) {
  return value.replace(/\/+$/, "");
}

const basePath = computed(() => {
  const configuredPath = metaStore.config?.base_path?.trim();
  if (!configuredPath || configuredPath === "/") return "";
  return `/${configuredPath.replace(/^\/+|\/+$/g, "")}`;
});

function publicPath(path: string) {
  return `${basePath.value}${path}`;
}

const publicBaseUrl = computed(() => {
  const configuredUrl = metaStore.config?.public_url?.trim();
  if (configuredUrl) return trimTrailingSlashes(configuredUrl);
  if (typeof window !== "undefined") return window.location.origin;
  return "";
});

const publicBasePathUrl = computed(() =>
  publicBaseUrl.value ? `${publicBaseUrl.value}${basePath.value}` : "",
);

const publicHost = computed(() =>
  publicBasePathUrl.value.replace(/^https?:\/\//, ""),
);

const publicApiBaseUrl = computed(() =>
  publicBasePathUrl.value ? `${publicBasePathUrl.value}/api` : "/api",
);

const publicSharingEnabled = computed(
  () => (settings.value?.user.share_data_max_days ?? 0) !== 0,
);

const invitesEnabled = computed(
  () => metaStore.config?.invite_codes_enabled ?? true,
);

const wakatimeConnected = computed(() =>
  Boolean(settings.value?.user.wakatime_connected),
);

const readmeStatsBaseUrl = computed(() =>
  trimTrailingSlashes(
    dataForm.readme_stats_base_url.trim() || githubReadmeStatsDefaultUrl,
  ),
);

const readmeStatsUrl = computed(() => {
  const userId = settings.value?.user.id;
  if (!userId) return "";

  const params = new URLSearchParams({
    username: userId,
    api_domain: publicHost.value,
    bg_color: "1A202C",
    title_color: "2F855A",
    icon_color: "2F855A",
    text_color: "ffffff",
    custom_title:
      settings.value?.readme_card_custom_title || "Waka3x Stats",
    layout: "compact",
  });

  return `${readmeStatsBaseUrl.value}/api/wakatime?${params.toString()}`;
});

const badgeExamples = computed(() => {
  const userId = settings.value?.user.id;
  if (!userId) return [];

  const encodedUserId = encodeURIComponent(userId);
  const todayUrl = `${publicApiBaseUrl.value}/badge/${encodedUserId}/interval:today?label=today`;
  const last30DaysUrl = `${publicApiBaseUrl.value}/badge/${encodedUserId}/interval:30_days?label=last%2030d`;
  const shieldsEndpoint = `${publicApiBaseUrl.value}/compat/shields/v1/${encodedUserId}/interval:all_time`;
  const shieldsUrl = `https://img.shields.io/endpoint?url=${encodeURIComponent(shieldsEndpoint)}&label=All%20time&color=blue`;

  return [
    { name: "Today", url: todayUrl, imageUrl: todayUrl, alt: "Today badge" },
    {
      name: "Last 30 days",
      url: last30DaysUrl,
      imageUrl: last30DaysUrl,
      alt: "Last 30 days badge",
    },
    {
      name: "Shields.io",
      url: shieldsUrl,
      imageUrl: shieldsUrl,
      alt: "Shields.io badge",
    },
  ];
});

watch(
  () => settings.value?.user,
  (user) => {
    if (!user) return;
    profileForm.email = user.email || "";
    profileForm.location = user.location || browserTimezone;
    profileForm.reports_weekly = user.reports_weekly;
    profileForm.start_of_week = user.start_of_week || 1;
    dataForm.exclude_unknown_projects = user.exclude_unknown_projects;
    dataForm.heartbeats_timeout = user.heartbeats_timeout_min;
    dataForm.readme_stats_base_url = user.readme_stats_base_url || "";
    sharingForm.enable_leaderboard = user.public_leaderboard;
    sharingForm.max_days = user.share_data_max_days;
    sharingForm.share_projects = user.share_projects;
    sharingForm.share_languages = user.share_languages;
    sharingForm.share_editors = user.share_editors;
    sharingForm.share_oss = user.share_oss;
    sharingForm.share_machines = user.share_machines;
    sharingForm.share_labels = user.share_labels;
    sharingForm.share_activity_chart = user.share_activity_chart;
    wakatimeForm.api_url =
      user.wakatime_api_url || settings.value?.default_wakatime_url || "";
  },
  { immediate: true },
);

onMounted(() => {
  settingsStore.fetchSettings().catch(() => undefined);
  metaStore.fetchConfig().catch(() => undefined);
  if (typeof window !== "undefined") {
    vibrantColorsEnabled.value =
      window.localStorage.getItem("wakapi_vibrant_colors") === "true";
  }
});

function notifySuccess(message: string) {
  toast.success(message);
}

function notifyError(err: unknown, fallback: string) {
  toast.error("Action failed", {
    description: getApiErrorMessage(err, fallback),
  });
}

async function saveProfile() {
  try {
    await authStore.updateProfile({
      email: profileForm.email,
      location: profileForm.location,
      reports_weekly: profileForm.reports_weekly,
      start_of_week: Number(profileForm.start_of_week),
      public_leaderboard: sharingForm.enable_leaderboard,
    });
    await settingsStore.fetchSettings();
    notifySuccess("Profile updated.");
  } catch (err) {
    notifyError(err, "Could not update profile.");
  }
}

async function action(label: string, fn: () => Promise<unknown>) {
  try {
    const response = (await fn()) as { message?: string };
    notifySuccess(response?.message || label);
  } catch (err) {
    notifyError(err, label);
  }
}

async function addPasskey() {
  await action("Could not add passkey.", async () => {
    const options = await settingsApi.webAuthnOptions();
    const credential = await createPasskeyCredential(options);
    const response = await settingsApi.webAuthnAdd({
      authenticator_name: newPasskeyName.value,
      credential_json: credential,
    });
    newPasskeyName.value = "";
    await settingsStore.fetchSettings();
    return response;
  });
}

function setSharingField(key: SharingToggleKey, checked: boolean | "indeterminate") {
  sharingForm[key] = Boolean(checked);
}

function setVibrantColors(checked: boolean | "indeterminate") {
  vibrantColorsEnabled.value = Boolean(checked);
  if (typeof window !== "undefined") {
    window.localStorage.setItem(
      "wakapi_vibrant_colors",
      String(vibrantColorsEnabled.value),
    );
  }
}

async function copyText(value: string) {
  try {
    await navigator.clipboard.writeText(value);
    notifySuccess("Copied to clipboard.");
  } catch {
    toast.error("Could not copy to clipboard.");
  }
}

async function importWakatimeData() {
  if (
    typeof window !== "undefined" &&
    !window.confirm("Are you sure? The import can not be undone.")
  ) {
    return;
  }

  await action("Could not import WakaTime data.", () =>
    settingsStore.importWakatime({
      use_legacy_importer: wakatimeForm.use_legacy_importer,
    }),
  );
}

async function disconnectWakatime() {
  if (
    typeof window !== "undefined" &&
    !window.confirm("Disconnect WakaTime forwarding?")
  ) {
    return;
  }

  await action("Could not disconnect WakaTime.", () =>
    settingsStore.updateWakatime({
      api_url: "",
      api_key: "",
    }),
  );
}
</script>

<template>
  <div class="p-4 md:p-8 space-y-8">
    <div>
      <h1 class="text-3xl font-bold tracking-tight text-primary">Settings</h1>
      <p class="text-muted-foreground">
        Manage account, data, integrations and API access.
      </p>
    </div>

    <div
      v-if="settingsStore.loading && !settings"
      class="flex justify-center py-20"
    >
      <Loader2 class="w-8 h-8 animate-spin text-primary" />
    </div>

    <Tabs v-else v-model="activeTab" class="w-full space-y-6">
      <TabsList class="bg-card border h-auto p-1 flex-wrap justify-start">
        <TabsTrigger value="account" class="gap-2 py-2 px-4"
          ><User class="w-4 h-4" /> Account</TabsTrigger
        >
        <TabsTrigger value="data" class="gap-2 py-2 px-4"
          ><Database class="w-4 h-4" /> Data</TabsTrigger
        >
        <TabsTrigger value="permissions" class="gap-2 py-2 px-4"
          ><ShieldCheck class="w-4 h-4" /> Permissions</TabsTrigger
        >
        <TabsTrigger value="integrations" class="gap-2 py-2 px-4"
          ><Plug class="w-4 h-4" /> Integrations</TabsTrigger
        >
        <TabsTrigger value="api_keys" class="gap-2 py-2 px-4"
          ><Key class="w-4 h-4" /> API Keys</TabsTrigger
        >
        <TabsTrigger value="danger" class="gap-2 py-2 px-4 text-destructive"
          ><AlertTriangle class="w-4 h-4" /> Danger</TabsTrigger
        >
      </TabsList>

      <TabsContent value="account" class="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle class="text-primary">Profile</CardTitle>
            <CardDescription
              >Update email, timezone and reporting
              preferences.</CardDescription
            >
          </CardHeader>
          <CardContent class="space-y-5">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div class="space-y-2">
                <Label>Email</Label>
                <Input v-model="profileForm.email" />
              </div>
              <div class="space-y-2">
                <Label>Timezone</Label>
                <Select v-model="profileForm.location">
                  <SelectTrigger class="w-full">
                    <SelectValue placeholder="Select timezone" />
                  </SelectTrigger>
                  <SelectContent class="max-h-80">
                    <SelectGroup>
                      <SelectItem
                        v-for="timezone in timezoneOptions"
                        :key="timezone"
                        :value="timezone"
                      >
                        {{
                          timezone === browserTimezone
                            ? `${timezone} (browser)`
                            : timezone
                        }}
                      </SelectItem>
                    </SelectGroup>
                  </SelectContent>
                </Select>
              </div>
              <div class="space-y-2">
                <Label>Start of week</Label>
                <Select v-model="profileForm.start_of_week">
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem
                      v-for="day in weekDays"
                      :key="day.value"
                      :value="day.value"
                    >
                      {{ day.label }}
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div
                class="flex items-center justify-between rounded-md border p-3"
              >
                <Label>Weekly email reports</Label>
                <Switch
                  :checked="profileForm.reports_weekly"
                  @update:checked="profileForm.reports_weekly = $event"
                />
              </div>
            </div>
            <Button :disabled="authStore.loading" @click="saveProfile"
              >Save profile</Button
            >
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle class="text-primary">Security</CardTitle>
          </CardHeader>
          <CardContent class="grid gap-6 md:grid-cols-2">
            <div class="space-y-3">
              <Label>Change password</Label>
              <Input
                v-model="passwordForm.password_old"
                type="password"
                placeholder="Current password"
              />
              <Input
                v-model="passwordForm.password_new"
                type="password"
                placeholder="New password"
              />
              <Input
                v-model="passwordForm.password_repeat"
                type="password"
                placeholder="Repeat new password"
              />
              <Button
                variant="outline"
                @click="
                  action('Could not update password.', () =>
                    settingsStore.changePassword(passwordForm),
                  )
                "
                >Update password</Button
              >
            </div>
            <div class="space-y-3">
              <Label>Change username</Label>
              <Input
                v-model="usernameForm.new_userid"
                :placeholder="settings?.user.id"
              />
              <Button
                variant="outline"
                @click="
                  action('Could not change username.', () =>
                    settingsStore.changeUserId(usernameForm),
                  )
                "
                >Change username</Button
              >
              <div v-if="invitesEnabled" class="space-y-2 pt-3">
                <Label>Invite link</Label>
                <div class="flex gap-2">
                  <Input
                    v-model="inviteLink"
                    readonly
                    placeholder="Generate an invite link"
                  />
                  <Button
                    variant="outline"
                    @click="
                      action('Could not generate invite.', async () => {
                        const r = await settingsStore.generateInvite();
                        inviteLink = r.invite_link;
                        return r;
                      })
                    "
                    >Generate</Button
                  >
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card v-if="!settings?.disable_webauthn">
          <CardHeader>
            <CardTitle class="text-primary">Passkeys</CardTitle>
            <CardDescription
              >Register or remove WebAuthn authenticators.</CardDescription
            >
          </CardHeader>
          <CardContent class="space-y-4">
            <div class="flex gap-2">
              <Input
                v-model="newPasskeyName"
                placeholder="Authenticator name"
              />
              <Button @click="addPasskey">Add passkey</Button>
            </div>
            <div
              v-for="credential in settings?.webauthn_credentials || []"
              :key="credential.name"
              class="flex items-center justify-between rounded-md border p-3"
            >
              <span>{{ credential.name }}</span>
              <Button
                variant="destructive"
                size="sm"
                @click="
                  action('Could not delete passkey.', () =>
                    settingsStore.deleteWebAuthn({
                      credential_name: credential.name,
                    }),
                  )
                "
              >
                <Trash2 class="w-4 h-4" />
              </Button>
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="data" class="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle class="text-primary">Data Processing</CardTitle>
          </CardHeader>
          <CardContent class="space-y-5">
            <div
              class="flex items-center justify-between rounded-md border p-3"
            >
              <div>
                <Label>Exclude unknown projects</Label>
                <p class="text-xs text-muted-foreground">
                  Triggers summary regeneration.
                </p>
              </div>
              <Switch
                :checked="dataForm.exclude_unknown_projects"
                @update:checked="dataForm.exclude_unknown_projects = $event"
              />
            </div>
            <Button
              variant="outline"
              @click="
                action('Could not update unknown-project behavior.', () =>
                  settingsStore.updateUnknownProjects({
                    exclude_unknown_projects: dataForm.exclude_unknown_projects,
                  }),
                )
              "
              >Save unknown-project behavior</Button
            >
            <div class="grid md:grid-cols-2 gap-4">
              <div class="space-y-2">
                <Label>Heartbeat timeout in minutes</Label>
                <Input
                  v-model.number="dataForm.heartbeats_timeout"
                  type="number"
                  min="1"
                  max="60"
                />
                <p class="text-xs text-muted-foreground">
                  Minimum 1 minute, maximum 60 minutes.
                </p>
                <Button
                  variant="outline"
                  @click="
                    action('Could not update heartbeat timeout.', () =>
                      settingsStore.updateHeartbeatsTimeout({
                        heartbeats_timeout: Number(dataForm.heartbeats_timeout),
                      }),
                    )
                  "
                  >Save timeout</Button
                >
              </div>
              <div
                class="flex items-center justify-between rounded-md border p-3"
              >
                <div>
                  <Label>Vibrant colors</Label>
                  <p class="text-xs text-muted-foreground">
                    Store this summary chart preference in your browser.
                  </p>
                </div>
                <Switch
                  :checked="vibrantColorsEnabled"
                  @update:checked="setVibrantColors"
                />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader
            ><CardTitle class="text-primary"
              >Aliases, Labels and Language Mappings</CardTitle
            ></CardHeader
          >
          <CardContent class="grid gap-6 lg:grid-cols-3">
            <div class="space-y-3">
              <Label>Aliases</Label>
              <Select v-model="newAlias.type">
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem :value="0">Project</SelectItem>
                  <SelectItem :value="1">Language</SelectItem>
                  <SelectItem :value="2">Editor</SelectItem>
                  <SelectItem :value="3">OS</SelectItem>
                  <SelectItem :value="4">Machine</SelectItem>
                </SelectContent>
              </Select>
              <Input v-model="newAlias.key" placeholder="Canonical key" />
              <Input v-model="newAlias.value" placeholder="Alias value" />
              <Button
                size="sm"
                class="gap-2"
                @click="
                  action('Could not add alias.', () =>
                    settingsStore.addAlias(newAlias),
                  )
                "
                ><Plus class="w-4 h-4" /> Add alias</Button
              >
              <div
                v-for="alias in settings?.aliases || []"
                :key="`${alias.type}-${alias.key}`"
                class="rounded-md border p-2 text-xs"
              >
                <div class="flex justify-between gap-2">
                  <span class="font-semibold">{{ alias.key }}</span>
                  <button
                    class="text-destructive"
                    @click="
                      action('Could not delete alias.', () =>
                        settingsStore.deleteAlias({
                          type: alias.type,
                          key: alias.key,
                        }),
                      )
                    "
                  >
                    Delete
                  </button>
                </div>
                <div class="text-muted-foreground">
                  {{ alias.values.join(", ") }}
                </div>
              </div>
            </div>
            <div class="space-y-3">
              <Label>Project labels</Label>
              <Input v-model="newLabel.key" placeholder="Project" />
              <Input v-model="newLabel.value" placeholder="Label" />
              <Button
                size="sm"
                class="gap-2"
                @click="
                  action('Could not add label.', () =>
                    settingsStore.addLabel({
                      key: [newLabel.key],
                      value: newLabel.value,
                    }),
                  )
                "
                ><Plus class="w-4 h-4" /> Add label</Button
              >
              <div
                v-for="label in settings?.labels || []"
                :key="label.key"
                class="rounded-md border p-2 text-xs"
              >
                <div class="font-semibold">{{ label.key }}</div>
                <div
                  v-for="project in label.values"
                  :key="project"
                  class="flex justify-between gap-2 text-muted-foreground"
                >
                  <span>{{ project }}</span>
                  <button
                    class="text-destructive"
                    @click="
                      action('Could not delete label.', () =>
                        settingsStore.deleteLabel({
                          key: label.key,
                          value: project,
                        }),
                      )
                    "
                  >
                    Delete
                  </button>
                </div>
              </div>
            </div>
            <div class="space-y-3">
              <Label>Language mappings</Label>
              <Input
                v-model="newMapping.extension"
                placeholder="Extension, e.g. vue"
              />
              <Input v-model="newMapping.language" placeholder="Language" />
              <Button
                size="sm"
                class="gap-2"
                @click="
                  action('Could not add mapping.', () =>
                    settingsStore.addLanguageMapping(newMapping),
                  )
                "
                ><Plus class="w-4 h-4" /> Add mapping</Button
              >
              <div
                v-for="mapping in settings?.language_mappings || []"
                :key="mapping.id"
                class="flex justify-between rounded-md border p-2 text-xs"
              >
                <span>.{{ mapping.extension }} -> {{ mapping.language }}</span>
                <button
                  class="text-destructive"
                  @click="
                    action('Could not delete mapping.', () =>
                      settingsStore.deleteLanguageMapping({
                        mapping_id: Number(mapping.id),
                      }),
                    )
                  "
                >
                  Delete
                </button>
              </div>
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="permissions" class="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle class="text-primary">Sharing and Leaderboard</CardTitle>
            <CardDescription
              >Control what is public in badges, readme cards and shared
              profiles.</CardDescription
            >
          </CardHeader>
          <CardContent class="space-y-4">
            <div
              class="flex items-center justify-between rounded-md border p-3"
            >
              <Label>Public leaderboard</Label>
              <Switch
                :checked="sharingForm.enable_leaderboard"
                @update:checked="sharingForm.enable_leaderboard = $event"
              />
            </div>
            <Button
              variant="outline"
              @click="
                action('Could not update leaderboard setting.', () =>
                  settingsStore.updateLeaderboard({
                    enable_leaderboard: sharingForm.enable_leaderboard,
                  }),
                )
              "
              >Save leaderboard setting</Button
            >
            <div class="space-y-2">
              <Label>Maximum shared range in days</Label>
              <Input
                v-model.number="sharingForm.max_days"
                type="number"
                min="-1"
              />
              <p class="text-xs text-muted-foreground">
                0 disables public data. Use -1 to share all historical data.
              </p>
            </div>
            <div class="grid md:grid-cols-2 gap-3">
              <label
                v-for="option in sharingOptions"
                :key="option.key"
                class="flex items-center gap-2 rounded-md border p-3 text-sm"
              >
                <Checkbox
                  :checked="sharingForm[option.key]"
                  @update:checked="setSharingField(option.key, $event)"
                />
                <span>{{ option.label }}</span>
              </label>
            </div>
            <Button
              @click="
                action('Could not update sharing.', () =>
                  settingsStore.updateSharing(sharingForm),
                )
              "
              >Save sharing</Button
            >
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="integrations" class="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle class="text-primary">WakaTime</CardTitle>
            <CardDescription
              >Relay credentials and import existing WakaTime
              data.</CardDescription
            >
          </CardHeader>
          <CardContent class="space-y-4">
            <div class="grid md:grid-cols-2 gap-4">
              <div class="space-y-2">
                <Label>API URL</Label>
                <Input
                  v-model="wakatimeForm.api_url"
                  :placeholder="settings?.default_wakatime_url"
                  :readonly="wakatimeConnected"
                />
              </div>
              <div class="space-y-2">
                <Label>API key</Label>
                <Input
                  v-model="wakatimeForm.api_key"
                  type="password"
                  :placeholder="wakatimeConnected ? '********' : 'WakaTime API key'"
                  :readonly="wakatimeConnected"
                />
              </div>
            </div>
            <label class="flex items-center gap-2 text-sm text-foreground">
              <Checkbox
                :checked="wakatimeForm.use_legacy_importer"
                @update:checked="
                  wakatimeForm.use_legacy_importer = Boolean($event)
                "
              />
              Use legacy importer
            </label>
            <div class="flex flex-wrap gap-2">
              <Button
                v-if="!wakatimeConnected"
                @click="
                  action('Could not update WakaTime credentials.', () =>
                    settingsStore.updateWakatime({
                      api_url: wakatimeForm.api_url,
                      api_key: wakatimeForm.api_key,
                    }),
                  )
                "
                >Save credentials</Button
              >
              <Button
                v-if="wakatimeConnected"
                variant="outline"
                @click="importWakatimeData"
                >Import WakaTime data</Button
              >
              <Button
                v-if="wakatimeConnected"
                variant="destructive"
                @click="disconnectWakatime"
                >Disconnect</Button
              >
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle class="flex items-center gap-2 text-primary">
              <Image class="w-4 h-4" />
              Badges
            </CardTitle>
            <CardDescription>
              Generate badges for README pages, forums and Shields.io.
            </CardDescription>
          </CardHeader>
          <CardContent class="space-y-4">
            <Alert v-if="!publicSharingEnabled">
              <AlertTitle>Public data is disabled</AlertTitle>
              <AlertDescription>
                Set a non-zero public data range in Permissions before badge
                endpoints can be used without authentication.
              </AlertDescription>
            </Alert>
            <div v-else class="space-y-3">
              <div
                v-for="badge in badgeExamples"
                :key="badge.name"
                class="grid gap-3 rounded-md border p-3 lg:grid-cols-[180px_1fr_auto]"
              >
                <div class="flex min-h-9 items-center">
                  <img :src="badge.imageUrl" :alt="badge.alt" />
                </div>
                <Input
                  :model-value="badge.url"
                  readonly
                  class="font-mono text-xs text-primary"
                />
                <Button
                  variant="outline"
                  size="icon"
                  :aria-label="`Copy ${badge.name} badge URL`"
                  @click="copyText(badge.url)"
                >
                  <Copy class="w-4 h-4" />
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle class="flex items-center gap-2 text-primary">
              <ExternalLink class="w-4 h-4" />
              GitHub Readme Stats
            </CardTitle>
            <CardDescription>
              Build a GitHub Readme Stats WakaTime card backed by Waka3x
              stats endpoint.
            </CardDescription>
          </CardHeader>
          <CardContent class="space-y-4">
            <Alert v-if="!publicSharingEnabled">
              <AlertTitle>Public data is disabled</AlertTitle>
              <AlertDescription>
                Share at least 7 days for a weekly card, 366 days for a yearly
                card, or -1 for an all-time card.
              </AlertDescription>
            </Alert>
            <template v-else>
              <div class="overflow-x-auto rounded-md border bg-muted/40 p-3">
                <img
                  :src="readmeStatsUrl"
                  alt="GitHub Readme Stats WakaTime card"
                  class="max-w-full"
                />
              </div>
              <div class="flex flex-col gap-2 md:flex-row">
                <Input
                  :model-value="readmeStatsUrl"
                  readonly
                  class="font-mono text-xs text-primary"
                />
                <Button
                  variant="outline"
                  size="icon"
                  aria-label="Copy GitHub Readme Stats URL"
                  @click="copyText(readmeStatsUrl)"
                >
                  <Copy class="w-4 h-4" />
                </Button>
              </div>
              <div class="space-y-2">
                <Label>Custom github-readme-stats base URL</Label>
                <div class="flex flex-col gap-2 md:flex-row">
                  <Input
                    v-model="dataForm.readme_stats_base_url"
                    type="url"
                    placeholder="https://github-readme-stats.vercel.app"
                  />
                  <Button
                    variant="outline"
                    @click="
                      action('Could not update readme stats URL.', () =>
                        settingsStore.updateReadmeStatsBaseUrl({
                          readme_stats_base_url:
                            dataForm.readme_stats_base_url,
                        }),
                      )
                    "
                    >Save URL</Button
                  >
                </div>
              </div>
            </template>
          </CardContent>
        </Card>
        <Card v-if="settings?.subscriptions_enabled">
          <CardHeader
            ><CardTitle class="text-primary"
              >Subscription</CardTitle
            ></CardHeader
          >
          <CardContent class="flex flex-wrap items-center gap-3">
            <span class="text-sm text-muted-foreground"
              >Standard plan:
              {{ settings.subscription_price || "configured in Stripe" }}</span
            >
            <form method="post" :action="publicPath('/subscription/checkout')">
              <Button>Checkout</Button>
            </form>
            <form method="post" :action="publicPath('/subscription/portal')">
              <Button variant="outline">Customer portal</Button>
            </form>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="api_keys">
        <Card>
          <CardHeader>
            <CardTitle class="text-primary">API Keys</CardTitle>
            <CardDescription
              >Use API keys for editor plugins and
              integrations.</CardDescription
            >
          </CardHeader>
          <CardContent class="space-y-4">
            <div class="grid md:grid-cols-[1fr_auto_auto] gap-2">
              <Input v-model="newApiKey.api_name" placeholder="Key name" />
              <label
                class="flex items-center gap-2 rounded-md border px-3 text-sm"
                ><Checkbox
                  :checked="newApiKey.api_readonly"
                  @update:checked="newApiKey.api_readonly = Boolean($event)"
                />
                Read only</label
              >
              <Button
                @click="
                  action('Could not add API key.', () =>
                    settingsStore.addApiKey(newApiKey),
                  )
                "
                >Add key</Button
              >
            </div>
            <div
              v-for="keyItem in settings?.api_keys || []"
              :key="keyItem.value"
              class="flex flex-col md:flex-row md:items-center justify-between gap-2 rounded-md border p-3"
            >
              <div>
                <div class="font-semibold">
                  {{ keyItem.name }}
                  <span
                    v-if="keyItem.main"
                    class="text-xs text-muted-foreground"
                    >(main)</span
                  >
                </div>
                <div class="font-mono text-xs text-muted-foreground break-all">
                  {{ keyItem.value }}
                </div>
              </div>
              <div class="flex gap-2">
                <Button
                  v-if="keyItem.main"
                  variant="outline"
                  size="sm"
                  @click="
                    action('Could not reset API key.', () =>
                      settingsStore.resetApiKey(),
                    )
                  "
                  >Reset</Button
                >
                <Button
                  v-else
                  variant="destructive"
                  size="sm"
                  @click="
                    action('Could not delete API key.', () =>
                      settingsStore.deleteApiKey({
                        api_key_value: keyItem.value,
                      }),
                    )
                  "
                  >Delete</Button
                >
              </div>
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="danger" class="space-y-6">
        <Card class="border-destructive/50">
          <CardHeader>
            <CardTitle class="text-destructive">Critical Actions</CardTitle>
            <CardDescription
              >These actions can take time and may be
              irreversible.</CardDescription
            >
          </CardHeader>
          <CardContent class="space-y-4">
            <div
              class="flex items-center justify-between gap-4 p-4 border border-destructive/20 rounded-lg bg-destructive/5"
            >
              <div>
                <h4 class="font-bold text-primary">Regenerate summaries</h4>
                <p class="text-xs text-muted-foreground">
                  Recompute all statistics from raw heartbeat data.
                </p>
              </div>
              <Button
                variant="destructive"
                size="sm"
                @click="
                  action('Could not regenerate summaries.', () =>
                    settingsStore.regenerateSummaries(),
                  )
                "
                >Regenerate</Button
              >
            </div>
            <div
              class="flex items-center justify-between gap-4 p-4 border border-destructive/20 rounded-lg bg-destructive/5"
            >
              <div>
                <h4 class="font-bold text-primary">Clear all data</h4>
                <p class="text-xs text-muted-foreground">
                  Delete all time tracking data.
                </p>
              </div>
              <Button
                variant="destructive"
                size="sm"
                @click="
                  action('Could not clear data.', () =>
                    settingsStore.clearData(),
                  )
                "
                >Clear Data</Button
              >
            </div>
            <div
              class="flex items-center justify-between gap-4 p-4 border border-destructive/20 rounded-lg bg-destructive/5"
            >
              <div>
                <h4 class="font-bold text-primary">Delete account</h4>
                <p class="text-xs text-muted-foreground">
                  Queue permanent account deletion and log out.
                </p>
              </div>
              <Button
                variant="destructive"
                size="sm"
                @click="
                  action('Could not delete account.', () =>
                    settingsStore
                      .deleteAccount()
                      .then(() => authStore.logout()),
                  )
                "
                >Delete Account</Button
              >
            </div>
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>
  </div>
</template>
