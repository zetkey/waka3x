<script setup lang="ts">
import {
  Search,
  Bell,
  ChevronDown,
  LogOut,
  User as UserIcon,
  Settings as SettingsIcon,
  CreditCard,
} from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useAuthStore } from "@/stores/auth";
import { useRouter } from "vue-router";

defineProps<{
  title?: string;
}>();

const authStore = useAuthStore();
const router = useRouter();

const handleLogout = async () => {
  await authStore.logout();
  router.push("/");
};
</script>

<template>
  <div class="flex">
    <div class="flex-1 max-w-md relative hidden sm:block">
      <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
      <Input
        placeholder="Search..."
        class="pl-9 bg-muted/50 border-none h-9 focus-visible:ring-1"
      />
    </div>

    <div class="flex items-center gap-4">
      <Button variant="ghost" size="icon">
        <Bell class="w-5 h-5" />
      </Button>

      <DropdownMenu v-if="authStore.user">
        <DropdownMenuTrigger as-child>
          <Button variant="ghost" class="pl-0 pr-2 flex items-center gap-2 h-9">
            <Avatar class="h-8 w-8">
              <AvatarImage
                :src="
                  authStore.user.avatar_url || 'https://github.com/nutlope.png'
                "
              />
              <AvatarFallback>{{
                authStore.user.id?.substring(0, 2).toUpperCase()
              }}</AvatarFallback>
            </Avatar>
            <div class="hidden lg:flex flex-col items-start text-xs text-left">
              <span class="font-bold text-primary leading-tight"
                >@{{ authStore.user.id }}</span
              >
              <span class="text-muted-foreground">{{
                authStore.user.has_active_subscription
                  ? "Pro Plan"
                  : "Free Plan"
              }}</span>
            </div>
            <ChevronDown class="h-4 w-4 text-muted-foreground" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" class="w-56">
          <DropdownMenuLabel>My Account</DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuItem @click="router.push('/settings')">
            <UserIcon class="mr-2 h-4 w-4" />
            <span>Profile</span>
          </DropdownMenuItem>
          <DropdownMenuItem @click="router.push('/settings')">
            <SettingsIcon class="mr-2 h-4 w-4" />
            <span>Settings</span>
          </DropdownMenuItem>
          <DropdownMenuItem @click="router.push('/settings')">
            <CreditCard class="mr-2 h-4 w-4" />
            <span>Billing</span>
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem class="text-destructive" @click="handleLogout">
            <LogOut class="mr-2 h-4 w-4" />
            <span>Log out</span>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  </div>
</template>
