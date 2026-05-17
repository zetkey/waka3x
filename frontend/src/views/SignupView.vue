<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { authApi, getApiErrorMessage } from "@/lib/api";
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
import { Lock, Mail, RefreshCcw, Rocket, Ticket, User } from "lucide-vue-next";
import { toast } from "vue-sonner";

const username = ref("");
const email = ref("");
const password = ref("");
const passwordRepeat = ref("");
const inviteCode = ref("");
const captchaId = ref("");
const captchaImage = ref("");
const captcha = ref("");
const authStore = useAuthStore();
const metaStore = useMetaStore();
const router = useRouter();
const route = useRoute();
const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone || "Local";

const config = computed(() => metaStore.config);
const signupAllowed = computed(() =>
  Boolean(config.value?.allow_signup || config.value?.invite_codes_enabled),
);

async function refreshCaptcha() {
  if (!config.value?.signup_captcha) return;
  const response = await authApi.captcha();
  captchaId.value = response.id;
  captchaImage.value = response.image_url;
  captcha.value = "";
}

onMounted(async () => {
  inviteCode.value = route.query.invite?.toString() || "";
  await metaStore.fetchConfig().catch(() => undefined);
  await refreshCaptcha().catch(() => undefined);
});

const handleSignup = async () => {
  try {
    await authStore.signup({
      username: username.value,
      email: email.value || undefined,
      password: password.value,
      password_repeat: passwordRepeat.value,
      invite_code: inviteCode.value || undefined,
      location: timezone,
      captcha_id: captchaId.value || undefined,
      captcha: captcha.value || undefined,
    });
    router.push("/dashboard");
  } catch (err) {
    toast({
      title: "Signup failed",
      description:
        authStore.error || getApiErrorMessage(err, "Could not create account."),
      variant: "destructive",
    });
    refreshCaptcha().catch(() => undefined);
  }
};
</script>

<template>
  <div
    class="min-h-screen flex flex-col items-center justify-center p-4 bg-background text-foreground"
  >
    <RouterLink to="/" class="mb-8 flex items-center gap-2">
      <Rocket class="w-10 h-10 text-primary" />
      <span class="text-3xl font-bold tracking-tighter text-foreground"
        >Waka3x</span
      >
    </RouterLink>

    <Card class="w-full max-w-md">
      <CardHeader>
        <CardTitle class="text-2xl text-foreground">
          Create an account
        </CardTitle>
        <CardDescription
          >Start tracking your coding activity on this server.</CardDescription
        >
      </CardHeader>
      <CardContent>
        <Alert
          v-if="config && !signupAllowed"
          variant="destructive"
          class="mb-4"
        >
          <AlertDescription
            >Registration is disabled on this server.</AlertDescription
          >
        </Alert>

        <form
          v-if="signupAllowed && !config?.disable_local_auth"
          class="space-y-4"
          @submit.prevent="handleSignup"
        >
          <div class="space-y-2">
            <Label for="username" class="text-foreground">Username</Label>
            <div class="relative">
              <User
                class="absolute left-3 top-3 w-4 h-4 text-muted-foreground"
              />
              <Input
                id="username"
                v-model="username"
                placeholder="johndoe"
                class="pl-10"
                required
              />
            </div>
          </div>
          <div class="space-y-2">
            <Label for="email" class="text-foreground">Email</Label>
            <div class="relative">
              <Mail
                class="absolute left-3 top-3 w-4 h-4 text-muted-foreground"
              />
              <Input
                id="email"
                v-model="email"
                type="email"
                placeholder="john@example.com"
                class="pl-10"
              />
            </div>
          </div>
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div class="space-y-2">
              <Label for="password" class="text-foreground">Password</Label>
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
            <div class="space-y-2">
              <Label for="passwordRepeat" class="text-foreground">Repeat</Label>
              <Input
                id="passwordRepeat"
                v-model="passwordRepeat"
                type="password"
                required
              />
            </div>
          </div>
          <div v-if="config?.invite_codes_enabled" class="space-y-2">
            <Label for="invite" class="text-foreground">Invite code</Label>
            <div class="relative">
              <Ticket
                class="absolute left-3 top-3 w-4 h-4 text-muted-foreground"
              />
              <Input
                id="invite"
                v-model="inviteCode"
                placeholder="Invite code"
                class="pl-10"
              />
            </div>
          </div>

          <div v-if="config?.signup_captcha" class="space-y-2">
            <Label for="captcha" class="text-foreground">Captcha</Label>
            <div class="flex items-center gap-3">
              <img
                v-if="captchaImage"
                :src="captchaImage"
                alt="Signup captcha"
                class="h-10 rounded border bg-white"
              />
              <Button
                type="button"
                variant="outline"
                size="sm"
                class="gap-2"
                @click="refreshCaptcha"
              >
                <RefreshCcw class="w-4 h-4" /> Reload
              </Button>
            </div>
            <Input id="captcha" v-model="captcha" required />
          </div>

          <Button class="w-full mt-6" :disabled="authStore.loading">
            {{ authStore.loading ? "Creating account..." : "Sign up" }}
          </Button>

          <div class="text-center mt-4 text-sm text-muted-foreground">
            Already have an account?
            <RouterLink
              to="/login"
              class="text-primary hover:underline font-semibold"
            >
              Log in
            </RouterLink>
          </div>
        </form>

        <div v-if="config?.oidc_providers?.length" class="mt-5 grid gap-2">
          <a
            v-for="provider in config.oidc_providers"
            :key="provider.name"
            :href="provider.login_url"
          >
            <Button variant="outline" class="w-full"
              >Continue with {{ provider.display_name }}</Button
            >
          </a>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
