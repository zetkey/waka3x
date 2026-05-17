<script setup lang="ts">
import { computed, onMounted } from "vue";
import { useMetaStore } from "@/stores/meta";
import { useAuthStore } from "@/stores/auth";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Code, Download, KeyRound, Rocket, Wrench } from "lucide-vue-next";

const metaStore = useMetaStore();
const authStore = useAuthStore();

onMounted(() => {
  metaStore.fetchSetup().catch(() => undefined);
});

const apiKey = computed(
  () => metaStore.setup?.api_key || authStore.user?.api_key || "<your-api-key>",
);
const apiUrl = computed(
  () => metaStore.setup?.base_url || `${window.location.origin}/api`,
);
const configSnippet = computed(
  () => `[settings]
api_url = ${apiUrl.value}
api_key = ${apiKey.value}`,
);
const configIntegrationSnippet = computed(
  () => `[settings]
api_key = defaults-to-this-api-key-when-not-defined-below
[api_urls]
.* = ${apiUrl.value}|${apiKey.value}
.* = https://api.wakatime.com/api/v1|waka-api-key`,
);
</script>

<template>
  <div class="min-h-screen bg-background text-foreground">
    <div class="max-w-4xl mx-auto px-4 py-8 space-y-8">
      <header class="flex items-center justify-between">
        <RouterLink to="/" class="flex items-center gap-2">
          <Rocket class="w-8 h-8 text-primary" />
          <span class="text-2xl font-bold tracking-tight">Waka3x</span>
        </RouterLink>
        <div class="flex gap-2">
          <RouterLink v-if="authStore.isAuthenticated" to="/dashboard">
            <Button variant="outline"> Dashboard </Button>
          </RouterLink>
          <RouterLink v-else to="/login">
            <Button>Log in</Button>
          </RouterLink>
        </div>
      </header>

      <section class="space-y-2">
        <h1 class="text-3xl font-bold">Setup</h1>
        <p class="text-muted-foreground">
          Configure WakaTime-compatible clients to send heartbeats to this
          server.
        </p>
      </section>

      <Alert v-if="!authStore.isAuthenticated">
        <AlertDescription
          >Log in to show your API key directly in the setup
          snippets.</AlertDescription
        >
      </Alert>

      <Card>
        <CardHeader>
          <CardTitle class="flex items-center gap-2">
            <Download class="w-5 h-5" /> Install the WakaTime plugin
          </CardTitle>
          <CardDescription
            >First step is to download and install the WakaTime plugin into your
            editor or IDE. Available for Visual Studio Code, PyCharm, IntelliJ,
            Neovim and many more – even Chrome or Firefox are
            supported.</CardDescription
          >
        </CardHeader>
        <CardContent class="space-y-3">
          <a
            href="https://wakatime.com/plugins"
            target="_blank"
            rel="noreferrer"
          >
            <Button variant="outline" class="w-full"
              >Open editor plugin list</Button
            >
          </a>
        </CardContent>
      </Card>

      <div class="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle class="flex items-center gap-2">
              <KeyRound class="w-5 h-5" /> Update your WakaTime config
            </CardTitle>
            <CardDescription>
              On Linux / macOS: <code>~/.wakatime.cfg</code><br />
              On Windows: <code>%USERPROFILE%\.wakatime.cfg</code>
            </CardDescription>
          </CardHeader>
          <CardContent>
            <pre class="rounded-md bg-muted p-3 text-sm overflow-x-auto">{{
              configSnippet
            }}</pre>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle class="flex items-center gap-2">
              <Code class="w-5 h-5" /> API endpoint
            </CardTitle>
            <CardDescription
              >Use this endpoint for WakaTime-compatible
              integrations.</CardDescription
            >
          </CardHeader>
          <CardContent class="space-y-2">
            <div class="font-mono text-sm rounded-md bg-muted p-3">
              {{ apiUrl }}
            </div>
            <div class="font-mono text-sm rounded-md bg-muted p-3 break-all">
              {{ apiKey }}
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle class="flex items-center gap-2">
            <Wrench class="w-5 h-5" />WakaTime integration
          </CardTitle>
          <CardDescription
            >You can use WakaTime and Waka3x in parallel, that is, have your
            coding activity tracked in both systems. This can be configured
            either on the client-side (preferred) on a system-wide- or
            per-project basis or using Waka3x's relay functionality (Settings →
            Integrations) to forward heartbeats to WakaTime.</CardDescription
          >
        </CardHeader>
        <CardContent>
          <p>Example:</p>
          <pre class="rounded-md bg-muted p-3 text-sm overflow-x-auto">{{
            configIntegrationSnippet
          }}</pre>
          <br />
          See
          <a
            href="https://github.com/wakatime/wakatime-cli/blob/develop/USAGE.md#api-urls-section"
            target="_blank"
            rel="noreferrer"
            class="text-blue-600 hover:underline"
            >wakatime-cli</a
          >
          API URLs section for details.
        </CardContent>
      </Card>
    </div>
  </div>
</template>
