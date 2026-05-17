import { createRouter, createWebHistory } from "vue-router";
import LandingView from "../views/LandingView.vue";

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
      meta: { layout: "Landing" },
    },
    {
      path: "/signup",
      name: "signup",
      component: () => import("../views/SignupView.vue"),
      meta: { layout: "Landing" },
    },
    {
      path: "/setup",
      name: "setup",
      component: () => import("../views/SetupView.vue"),
      meta: { layout: "Landing" },
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
      meta: { layout: "Dashboard" },
    },
    {
      path: "/projects",
      name: "projects",
      component: () => import("../views/ProjectsView.vue"),
      meta: { layout: "Dashboard" },
    },
    {
      path: "/settings",
      name: "settings",
      component: () => import("../views/SettingsView.vue"),
      meta: { layout: "Dashboard" },
    },
    {
      path: "/leaderboard",
      name: "leaderboard",
      component: () => import("../views/LeaderboardView.vue"),
      meta: { layout: "Dashboard" },
    },
    {
      path: "/summary",
      name: "summary",
      component: () => import("../views/SummaryView.vue"),
      meta: { layout: "Dashboard" },
    },
  ],
});

export default router;
