<script setup lang="ts">
import AppSidebar from "@/components/layout/AppSidebar.vue";
import { computed } from "vue";
import { useRoute } from "vue-router";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Separator } from "@/components/ui/separator";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import AppHeader from "@/components/layout/AppHeader.vue";

const route = useRoute();

const pageTitle = computed(() => {
  if (typeof route.name === "string") {
    return route.name
      .split("-")
      .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
      .join(" ");
  }
  return "Dashboard";
});
</script>

<template>
  <SidebarProvider class="h-screen min-h-0 overflow-hidden">
    <AppSidebar />
    <SidebarInset class="min-h-0 overflow-hidden">
      <header
        class="flex h-12 shrink-0 items-center justify-between gap-2 border-b bg-background/80 px-4 backdrop-blur-sm"
      >
        <div class="flex min-w-0 items-center gap-2">
          <SidebarTrigger class="-ml-1" />
          <Separator
            orientation="vertical"
            class="mr-2 data-[orientation=vertical]:h-8"
          />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem class="hidden md:block">
                <BreadcrumbLink href="/dashboard"> Waka3x </BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbSeparator class="hidden md:block" />
              <BreadcrumbItem>
                <BreadcrumbPage>{{ pageTitle }}</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </div>
        <AppHeader />
      </header>
      <main class="min-h-0 flex-1 overflow-y-auto">
        <slot />
      </main>
    </SidebarInset>
  </SidebarProvider>
</template>
