<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { authApi, getApiErrorMessage } from "@/lib/api";
import { getPasskeyAssertion } from "@/lib/webauthn";
import { useAuthStore } from "@/stores/auth";
import { useMetaStore } from "@/stores/meta";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Key, Lock, Rocket, User } from "lucide-vue-next";
import { Separator } from "@/components/ui/separator";
import { toast } from "vue-sonner";

const username = ref("");
const password = ref("");
const passkeyLoading = ref(false);
const authStore = useAuthStore();
const metaStore = useMetaStore();
const router = useRouter();
const route = useRoute();

const config = computed(() => metaStore.config);
const oidcError = computed(() =>
  route.query.error?.toString().replaceAll("_", " "),
);

onMounted(() => {
  metaStore.fetchConfig().catch(() => undefined);
});

const handleLogin = async () => {
  try {
    await authStore.login(username.value, password.value);
    router.push("/dashboard");
  } catch {
    toast({
      title: "Login failed",
      description: authStore.error || "Please check your credentials.",
      variant: "destructive",
    });
  }
};

const handlePasskeyLogin = async () => {
  passkeyLoading.value = true;
  try {
    const options = await authApi.webAuthnOptions();
    const assertion = await getPasskeyAssertion(options);
    const response = await authApi.webAuthnLogin({ assertion_json: assertion });
    authStore.user = response.user;
    authStore.isAuthenticated = true;
    router.push("/dashboard");
  } catch (err) {
    toast({
      title: "Passkey login failed",
      description: getApiErrorMessage(
        err,
        err instanceof Error
          ? err.message
          : "Could not authenticate with passkey.",
      ),
      variant: "destructive",
    });
  } finally {
    passkeyLoading.value = false;
  }
};
</script>

<template>
  <div
    class="min-h-screen flex flex-col items-center justify-center p-4 bg-background"
  >
    <RouterLink to="/" class="mb-8 flex items-center gap-2">
      <Rocket class="w-10 h-10 text-primary" />
      <span class="text-3xl font-bold tracking-tighter text-foreground"
        >Waka3x</span
      >
    </RouterLink>

    <Card class="w-full max-w-md">
      <CardHeader>
        <CardTitle class="text-2xl text-foreground"> Welcome back </CardTitle>
        <CardDescription>
          Log in to continue tracking your coding activity.
        </CardDescription>
      </CardHeader>
      <CardContent class="space-y-6">
        <Alert v-if="oidcError" variant="destructive">
          <AlertDescription>{{ oidcError }}</AlertDescription>
        </Alert>

        <form
          v-if="!config?.disable_local_auth"
          class="space-y-4"
          @submit.prevent="handleLogin"
        >
          <div class="space-y-2">
            <Label for="username" class="text-foreground"
              >Username or email</Label
            >
            <div class="relative">
              <User
                class="absolute left-3 top-3 w-4 h-4 text-muted-foreground"
              />
              <Input
                id="username"
                v-model="username"
                placeholder="Username or email"
                class="pl-10"
                required
              />
            </div>
          </div>
          <div class="space-y-2">
            <div class="flex items-center justify-between">
              <Label for="password" class="text-foreground">Password</Label>
              <RouterLink
                v-if="config?.mail_enabled"
                to="/reset-password"
                class="text-xs text-primary hover:underline"
              >
                Forgot password?
              </RouterLink>
            </div>
            <div class="relative">
              <Lock
                class="absolute left-3 top-3 w-4 h-4 text-muted-foreground"
              />
              <Input
                id="password"
                v-model="password"
                type="password"
                class="pl-10"
                required
              />
            </div>
          </div>
          <Button class="w-full" :disabled="authStore.loading">
            {{ authStore.loading ? "Logging in..." : "Log in" }}
          </Button>
        </form>

        <Alert v-else>
          <AlertDescription
            >Local username and password login is disabled on this
            server.</AlertDescription
          >
        </Alert>

        <template
          v-if="!config?.disable_webauthn || config?.oidc_providers?.length"
        >
          <div class="relative">
            <div class="absolute inset-0 flex items-center">
              <Separator />
            </div>
            <div class="relative flex justify-center text-xs uppercase">
              <span class="bg-card px-2 text-muted-foreground font-semibold"
                >Other sign-in methods</span
              >
            </div>
          </div>

          <div class="grid gap-3">
            <Button
              v-if="!config?.disable_webauthn"
              variant="outline"
              class="gap-2"
              :disabled="passkeyLoading"
              @click="handlePasskeyLogin"
            >
              <Key class="w-4 h-4" />
              {{
                passkeyLoading
                  ? "Waiting for passkey..."
                  : "Continue with passkey"
              }}
            </Button>
            <a
              v-for="provider in config?.oidc_providers || []"
              :key="provider.name"
              :href="provider.login_url"
            >
              <Button variant="outline" class="w-full">{{
                provider.display_name
              }}</Button>
            </a>
          </div>
        </template>

        <div
          class="flex items-center justify-between text-sm text-muted-foreground"
        >
          <RouterLink to="/setup" class="text-primary hover:underline">
            Setup instructions
          </RouterLink>
          <RouterLink
            v-if="config?.allow_signup || config?.invite_codes_enabled"
            to="/signup"
            class="text-primary hover:underline"
          >
            Create account
          </RouterLink>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
