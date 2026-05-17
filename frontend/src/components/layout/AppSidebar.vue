<script setup lang="ts">
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  type SidebarProps,
} from "@/components/ui/sidebar";
import {
  BarChart3,
  BookOpen,
  Folders,
  LayoutDashboard,
  Settings,
} from "lucide-vue-next";
import { useRoute } from "vue-router";

const props = withDefaults(defineProps<SidebarProps>(), {
  variant: "inset",
});

const route = useRoute();

const sidebarItems = [
  { name: "Dashboard", icon: LayoutDashboard, href: "/dashboard" },
  { name: "Leaderboard", icon: BarChart3, href: "/leaderboard" },
  { name: "Projects", icon: Folders, href: "/projects" },
  { name: "Summary", icon: BarChart3, href: "/summary" },
  { name: "Setup", icon: BookOpen, href: "/setup" },
  { name: "Settings", icon: Settings, href: "/settings" },
];

const isActive = (path: string) => route.path === path;
</script>

<template>
  <Sidebar v-bind="props">
    <SidebarHeader>
      <div class="p-2 flex items-center gap-2">
        <img src="@/assets/icon.svg" alt="Waka3x-logo" class="h-8" />
        <span class="text-2xl font-bold tracking-tight italic text-green-800"
          >Waka<sup>3x</sup></span
        >
      </div>
    </SidebarHeader>
    <SidebarContent class="overflow-hidden">
      <SidebarGroup>
        <SidebarGroupLabel>Navigation</SidebarGroupLabel>
        <SidebarGroupContent>
          <SidebarMenu>
            <SidebarMenuItem v-for="item in sidebarItems" :key="item.name">
              <SidebarMenuButton
                as-child
                :is-active="isActive(item.href)"
                :tooltip="item.name"
              >
                <RouterLink :to="item.href">
                  <component :is="item.icon" class="w-4 h-4" />
                  <span>{{ item.name }}</span>
                </RouterLink>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarGroupContent>
      </SidebarGroup>
    </SidebarContent>
  </Sidebar>
</template>
