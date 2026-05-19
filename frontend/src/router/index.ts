import {
  createRouter,
  createWebHistory,
  type RouteLocationNormalized,
} from "vue-router";
import { useAuthStore } from "@/stores/auth";
import { useMetaStore } from "@/stores/meta";
import LandingView from "../views/LandingView.vue";

declare module "vue-router" {
  interface RouteMeta {
    layout?: "Landing" | "Dashboard";
    requiresAuth?: boolean;
    guestOnly?: boolean;
    requiresAuthWhenLeaderboardPrivate?: boolean;
    conditionalLayout?: boolean;
  }
}

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      name: "landing",
      component: LandingView,
      meta: { layout: "Landing" },
    },
    {
      path: "/login",
      name: "login",
      component: () => import("../views/LoginView.vue"),
      meta: { layout: "Landing", guestOnly: true },
    },
    {
      path: "/signup",
      name: "signup",
      component: () => import("../views/SignupView.vue"),
      meta: { layout: "Landing", guestOnly: true },
    },
    {
      path: "/setup",
      name: "setup",
      component: () => import("../views/SetupView.vue"),
      meta: { layout: "Landing", conditionalLayout: true },
    },
    {
      path: "/reset-password",
      name: "reset-password",
      component: () => import("../views/ResetPasswordView.vue"),
      meta: { layout: "Landing" },
    },
    {
      path: "/set-password",
      name: "set-password",
      component: () => import("../views/SetPasswordView.vue"),
      meta: { layout: "Landing" },
    },
    {
      path: "/imprint",
      name: "imprint",
      component: () => import("../views/ImprintView.vue"),
      meta: { layout: "Landing" },
    },
    {
      path: "/unsubscribe",
      name: "unsubscribe",
      component: () => import("../views/UnsubscribeView.vue"),
      meta: { layout: "Landing" },
    },
    {
      path: "/dashboard",
      name: "dashboard",
      component: () => import("../views/HomeView.vue"),
      meta: { layout: "Dashboard", requiresAuth: true },
    },
    {
      path: "/projects",
      name: "projects",
      component: () => import("../views/ProjectsView.vue"),
      meta: { layout: "Dashboard", requiresAuth: true },
    },
    {
      path: "/settings",
      name: "settings",
      component: () => import("../views/SettingsView.vue"),
      meta: { layout: "Dashboard", requiresAuth: true },
    },
    {
      path: "/leaderboard",
      name: "leaderboard",
      component: () => import("../views/LeaderboardView.vue"),
      meta: {
        layout: "Landing",
        requiresAuthWhenLeaderboardPrivate: true,
        conditionalLayout: true,
      },
    },
    {
      path: "/summary",
      name: "summary",
      component: () => import("../views/SummaryView.vue"),
      meta: { layout: "Dashboard", requiresAuth: true },
    },
  ],
});

async function routeRequiresAuth(to: RouteLocationNormalized) {
  if (to.matched.some((record) => record.meta.requiresAuth)) {
    return true;
  }

  if (
    !to.matched.some(
      (record) => record.meta.requiresAuthWhenLeaderboardPrivate,
    )
  ) {
    return false;
  }

  const metaStore = useMetaStore();
  try {
    const config = await metaStore.fetchConfig();
    return config.leaderboard_require_auth;
  } catch {
    return true;
  }
}

function isSafeAuthenticatedRedirect(path: string | undefined): path is string {
  if (!path) return false;
  return (
    path.startsWith("/") &&
    !path.startsWith("//") &&
    !path.startsWith("/login") &&
    !path.startsWith("/signup")
  );
}

router.beforeEach(async (to) => {
  const authStore = useAuthStore();
  const requiresAuth = await routeRequiresAuth(to);

  if (to.meta.conditionalLayout) {
    await authStore.fetchUser();
    to.meta.layout = authStore.isAuthenticated ? "Dashboard" : "Landing";
  }

  if (requiresAuth) {
    await authStore.fetchUser();
    if (!authStore.isAuthenticated) {
      return {
        name: "login",
        query: { redirect: to.fullPath },
      };
    }
  }

  if (to.matched.some((record) => record.meta.guestOnly)) {
    await authStore.fetchUser();
    if (authStore.isAuthenticated) {
      const redirect = to.query.redirect?.toString();
      return isSafeAuthenticatedRedirect(redirect)
        ? redirect
        : { name: "dashboard" };
    }
  }
});

export default router;
