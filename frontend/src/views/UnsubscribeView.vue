<script setup lang="ts">
import { computed, ref } from "vue";
import { useRoute } from "vue-router";
import { getApiErrorMessage, metaApi } from "@/lib/api";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { MailMinus, Rocket } from "lucide-vue-next";

const route = useRoute();
const token = computed(() => route.query.token?.toString() || "");
const loading = ref(false);
const message = ref("");
const error = ref("");

async function unsubscribe() {
  loading.value = true;
  message.value = "";
  error.value = "";
  try {
    const response = await metaApi.unsubscribe({ token: token.value });
    message.value = response.message;
  } catch (err) {
    error.value = getApiErrorMessage(err, "Could not unsubscribe.");
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
        <CardTitle class="flex items-center gap-2">
          <MailMinus class="w-5 h-5" /> Unsubscribe
        </CardTitle>
        <CardDescription
          >Stop receiving weekly coding summary reports.</CardDescription
        >
      </CardHeader>
      <CardContent class="space-y-4">
        <Alert v-if="message">
          <AlertDescription>{{ message }}</AlertDescription>
        </Alert>
        <Alert v-if="error || !token" variant="destructive">
          <AlertDescription>{{
            error || "Missing unsubscribe token."
          }}</AlertDescription>
        </Alert>
        <Button
          class="w-full"
          :disabled="loading || !token || Boolean(message)"
          @click="unsubscribe"
        >
          {{ loading ? "Unsubscribing..." : "Unsubscribe" }}
        </Button>
        <RouterLink
          to="/"
          class="block text-center text-sm text-primary hover:underline"
        >
          Back home
        </RouterLink>
      </CardContent>
    </Card>
  </div>
</template>
