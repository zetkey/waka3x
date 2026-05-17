<script setup lang="ts">
import { onMounted } from "vue";
import { useMetaStore } from "@/stores/meta";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Loader2, Rocket } from "lucide-vue-next";

const metaStore = useMetaStore();

onMounted(() => {
  metaStore.fetchImprint().catch(() => undefined);
});
</script>

<template>
  <div class="min-h-screen bg-background text-foreground">
    <div class="max-w-3xl mx-auto px-4 py-8 space-y-8">
      <header class="flex items-center justify-between">
        <RouterLink to="/" class="flex items-center gap-2">
          <Rocket class="w-8 h-8 text-primary" />
          <span class="text-2xl font-bold tracking-tight">Waka3x</span>
        </RouterLink>
        <RouterLink to="/login">
          <Button variant="outline"> Login </Button>
        </RouterLink>
      </header>

      <Card>
        <CardHeader>
          <CardTitle>Imprint, cookies and privacy</CardTitle>
        </CardHeader>
        <CardContent>
          <div v-if="metaStore.loading" class="flex justify-center py-12">
            <Loader2 class="w-6 h-6 animate-spin text-primary" />
          </div>
          <div v-else class="prose prose-invert max-w-none text-sm leading-7">
            {{ metaStore.imprint?.html || "No imprint configured." }}
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
