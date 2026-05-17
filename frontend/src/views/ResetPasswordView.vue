<script setup lang="ts">
import { ref } from "vue";
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
import { Mail, Rocket } from "lucide-vue-next";

const email = ref("");
const loading = ref(false);
const message = ref("");
const error = ref("");

async function submit() {
  loading.value = true;
  message.value = "";
  error.value = "";
  try {
    const response = await authApi.resetPassword({ email: email.value });
    message.value = response.message;
  } catch (err) {
    error.value = getApiErrorMessage(
      err,
      "Could not request a password reset.",
    );
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
        <CardTitle>Reset password</CardTitle>
        <CardDescription
          >Enter the email address attached to your account.</CardDescription
        >
      </CardHeader>
      <CardContent>
        <form class="space-y-4" @submit.prevent="submit">
          <Alert v-if="message">
            <AlertDescription>{{ message }}</AlertDescription>
          </Alert>
          <Alert v-if="error" variant="destructive">
            <AlertDescription>{{ error }}</AlertDescription>
          </Alert>
          <div class="space-y-2">
            <Label for="email">Email</Label>
            <div class="relative">
              <Mail
                class="absolute left-3 top-3 w-4 h-4 text-muted-foreground"
              />
              <Input
                id="email"
                v-model="email"
                type="email"
                class="pl-10"
                required
              />
            </div>
          </div>
          <Button class="w-full" :disabled="loading">
            {{ loading ? "Sending..." : "Send reset link" }}
          </Button>
          <RouterLink
            to="/login"
            class="block text-center text-sm text-primary hover:underline"
          >
            Back to login
          </RouterLink>
        </form>
      </CardContent>
    </Card>
  </div>
</template>
