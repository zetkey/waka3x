<script setup lang="ts">
import { computed, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { authApi, getApiErrorMessage } from "@/lib/api";
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
import { Lock, Rocket } from "lucide-vue-next";

const route = useRoute();
const router = useRouter();
const password = ref("");
const passwordRepeat = ref("");
const loading = ref(false);
const message = ref("");
const error = ref("");
const token = computed(() => route.query.token?.toString() || "");

async function submit() {
  loading.value = true;
  message.value = "";
  error.value = "";
  try {
    const response = await authApi.setPassword({
      token: token.value,
      password: password.value,
      password_repeat: passwordRepeat.value,
    });
    message.value = response.message;
    setTimeout(() => router.push("/login"), 900);
  } catch (err) {
    error.value = getApiErrorMessage(err, "Could not set password.");
  } finally {
    loading.value = false;
  }
}
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
        <CardTitle>Set password</CardTitle>
        <CardDescription
          >Choose a new password for your account.</CardDescription
        >
      </CardHeader>
      <CardContent>
        <Alert v-if="!token" variant="destructive" class="mb-4">
          <AlertDescription>Missing reset token.</AlertDescription>
        </Alert>
        <form class="space-y-4" @submit.prevent="submit">
          <Alert v-if="message">
            <AlertDescription>{{ message }}</AlertDescription>
          </Alert>
          <Alert v-if="error" variant="destructive">
            <AlertDescription>{{ error }}</AlertDescription>
          </Alert>
          <div class="space-y-2">
            <Label for="password">New password</Label>
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
            <Label for="passwordRepeat">Repeat password</Label>
            <Input
              id="passwordRepeat"
              v-model="passwordRepeat"
              type="password"
              required
            />
          </div>
          <Button class="w-full" :disabled="loading || !token">
            {{ loading ? "Saving..." : "Set password" }}
          </Button>
        </form>
      </CardContent>
    </Card>
  </div>
</template>
