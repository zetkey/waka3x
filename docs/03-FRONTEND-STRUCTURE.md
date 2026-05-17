# Waka3x Frontend Structure

**AI Agent Reference Document**  
Last Updated: 2026-05-17

## Overview

The Waka3x frontend is a Vue 3 Single Page Application (SPA) built with TypeScript, Vite, and Tailwind CSS. It provides a modern, responsive interface for viewing coding statistics and managing user settings.

## Technology Stack

- **Framework**: Vue 3.5 (Composition API)
- **Build Tool**: Vite 8
- **Language**: TypeScript 6
- **Styling**: Tailwind CSS 4
- **UI Components**: shadcn-vue (based on Reka UI)
- **State Management**: Pinia stores (via @vueuse/core)
- **Routing**: Vue Router
- **HTTP Client**: Axios
- **Charts**: Chart.js + vue-chartjs
- **Form Validation**: vee-validate + zod
- **Icons**: lucide-vue-next, vue3-simple-icons

## Directory Structure

```
frontend/
├── src/
│   ├── main.ts                 # Application entry point
│   ├── App.vue                 # Root component
│   ├── router/
│   │   └── index.ts           # Route definitions
│   ├── stores/                # Pinia state stores
│   │   ├── auth.ts            # Authentication state
│   │   ├── meta.ts            # App metadata
│   │   └── settings.ts        # User settings
│   ├── views/                 # Page components (13 views)
│   │   ├── LandingView.vue    # Landing page
│   │   ├── LoginView.vue      # Login page
│   │   ├── SignupView.vue     # Signup page
│   │   ├── HomeView.vue       # Dashboard (main)
│   │   ├── ProjectsView.vue   # Projects page
│   │   ├── SettingsView.vue   # Settings page
│   │   ├── LeaderboardView.vue # Leaderboard
│   │   └── SummaryView.vue    # Summary statistics
│   ├── components/            # Reusable components
│   │   ├── layout/           # Layout components
│   │   ├── charts/           # Chart components
│   │   └── ui/               # shadcn-vue UI components
│   ├── layouts/              # Layout wrappers
│   │   ├── LandingLayout.vue # Public pages layout
│   │   └── DashboardLayout.vue # Authenticated pages layout
│   ├── lib/                  # Utilities and helpers
│   │   ├── api.ts            # Axios instance and API calls
│   │   ├── utils.ts          # Helper functions
│   │   └── constants.ts      # Constants
│   ├── types/                # TypeScript type definitions
│   │   ├── api.ts            # API response types
│   │   └── models.ts         # Data model types
│   └── assets/               # Static assets (images, etc.)
├── public/                   # Public static files
├── dist/                     # Built output (embedded in Go binary)
├── index.html                # HTML entry point
├── vite.config.ts            # Vite configuration
├── tailwind.config.ts        # Tailwind configuration
├── tsconfig.json             # TypeScript configuration
└── package.json              # Dependencies and scripts
```

## Key Concepts

### Composition API

All components use Vue 3 Composition API with `<script setup>` syntax:

```vue
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const count = ref(0)
const doubled = computed(() => count.value * 2)

onMounted(() => {
  // Component mounted
})
</script>

<template>
  <div>{{ doubled }}</div>
</template>
```

### State Management (Pinia)

**Stores Location**: `src/stores/`

**Key Stores**:
- `auth.ts` - User authentication state, login/logout
- `meta.ts` - App metadata (leaderboard settings, etc.)
- `settings.ts` - User preferences

**Example Store**:
```typescript
import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const isAuthenticated = computed(() => user.value !== null)
  
  async function login(credentials: LoginCredentials) {
    const response = await api.post('/api/login', credentials)
    user.value = response.data
  }
  
  return { user, isAuthenticated, login }
})
```

### Routing

**File**: `src/router/index.ts`

**Route Meta Fields**:
- `layout`: "Landing" | "Dashboard" - Which layout to use
- `requiresAuth`: boolean - Requires authentication
- `guestOnly`: boolean - Only accessible when not authenticated
- `requiresAuthWhenLeaderboardPrivate`: boolean - Conditional auth

**Example Routes**:
```typescript
{
  path: '/dashboard',
  name: 'dashboard',
  component: () => import('../views/HomeView.vue'),
  meta: { layout: 'Dashboard', requiresAuth: true }
}
```

**Navigation Guards**: Check authentication state before route changes

### API Communication

**File**: `src/lib/api.ts`

**Axios Instance**: Pre-configured with base URL and interceptors

```typescript
import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  withCredentials: true, // Send cookies
})

// Request interceptor
api.interceptors.request.use((config) => {
  // Add auth headers if needed
  return config
})

// Response interceptor
api.interceptors.response.use(
  (response) => response,
  (error) => {
    // Handle errors globally
    return Promise.reject(error)
  }
)

export default api
```

### UI Components (shadcn-vue)

**Location**: `src/components/ui/`

**Components**: Button, Card, Input, Table, Dialog, Dropdown, etc.

**Usage**: Import and use directly in components
```vue
<script setup lang="ts">
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
</script>

<template>
  <Card>
    <Button @click="handleClick">Click me</Button>
  </Card>
</template>
```

### Charts

**Library**: Chart.js + vue-chartjs

**Location**: `src/components/charts/`

**Example**:
```vue
<script setup lang="ts">
import { Line } from 'vue-chartjs'
import { Chart, registerables } from 'chart.js'

Chart.register(...registerables)

const chartData = {
  labels: ['Mon', 'Tue', 'Wed'],
  datasets: [{
    label: 'Coding Time',
    data: [2, 4, 3]
  }]
}
</script>

<template>
  <Line :data="chartData" />
</template>
```

## Key Views

### LandingView.vue
- Public landing page
- Shows features and signup CTA
- Layout: LandingLayout

### LoginView.vue
- Login form (username/password)
- OIDC provider buttons
- WebAuthn support
- Layout: LandingLayout

### HomeView.vue (Dashboard)
- Main dashboard after login
- Shows today's coding statistics
- Charts: activity heatmap, language breakdown
- Layout: DashboardLayout

### ProjectsView.vue
- List of all projects
- Project statistics
- Layout: DashboardLayout

### SettingsView.vue
- User settings management
- API key display
- Integrations (WakaTime relay)
- Aliases configuration
- Layout: DashboardLayout

### LeaderboardView.vue
- Public/private leaderboard
- Rankings by coding time
- Layout: DashboardLayout

## Layouts

### LandingLayout.vue
- Used for public pages (landing, login, signup)
- Simple header with logo
- No sidebar

### DashboardLayout.vue
- Used for authenticated pages
- Sidebar navigation
- User menu
- Responsive (mobile-friendly)

## Type Definitions

**Location**: `src/types/`

**Key Types**:
```typescript
// src/types/models.ts
export interface User {
  id: string
  email: string
  api_key: string
  created_at: string
  // ...
}

export interface Summary {
  user_id: string
  from: string
  to: string
  projects: ProjectStats[]
  languages: LanguageStats[]
  editors: EditorStats[]
  // ...
}

// src/types/api.ts
export interface ApiResponse<T> {
  data: T
  error?: string
}
```

## Styling Conventions

### Tailwind CSS

**Configuration**: `tailwind.config.ts`

**Usage**: Utility classes in templates
```vue
<template>
  <div class="flex items-center justify-between p-4 bg-white rounded-lg shadow">
    <h1 class="text-2xl font-bold text-gray-900">Title</h1>
  </div>
</template>
```

### Component Styling

- Prefer Tailwind utility classes over custom CSS
- Use `@apply` directive for repeated patterns
- Scoped styles when needed: `<style scoped>`

## Build and Development

### Development Server

```bash
cd frontend
bun install
bun dev
```

- Runs on port 5173
- Proxies API requests to backend (port 3000)
- Hot module replacement (HMR)

### Production Build

```bash
bun build
```

- Outputs to `frontend/dist/`
- Minified and optimized
- Embedded in Go binary via `//go:embed`

### Vite Configuration

**File**: `vite.config.ts`

```typescript
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    proxy: {
      '/api': 'http://localhost:3000'
    }
  }
})
```

## Common Patterns

### Fetching Data

```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/lib/api'

const data = ref(null)
const loading = ref(false)
const error = ref(null)

async function fetchData() {
  loading.value = true
  try {
    const response = await api.get('/summary')
    data.value = response.data
  } catch (e) {
    error.value = e
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchData()
})
</script>
```

### Form Handling

```vue
<script setup lang="ts">
import { useForm } from 'vee-validate'
import { z } from 'zod'
import { toTypedSchema } from '@vee-validate/zod'

const schema = z.object({
  email: z.string().email(),
  password: z.string().min(8)
})

const { handleSubmit, errors } = useForm({
  validationSchema: toTypedSchema(schema)
})

const onSubmit = handleSubmit(async (values) => {
  await api.post('/login', values)
})
</script>
```

### Conditional Rendering

```vue
<template>
  <div v-if="loading">Loading...</div>
  <div v-else-if="error">Error: {{ error }}</div>
  <div v-else>{{ data }}</div>
</template>
```

## Key Entry Points for AI Agents

**To understand the frontend**:
1. Start with `src/main.ts` - application initialization
2. Check `src/router/index.ts` - available routes
3. Look at `src/stores/` - state management
4. Examine `src/views/` - page components

**To add a new page**:
1. Create view component in `src/views/`
2. Add route in `src/router/index.ts`
3. Add navigation link in layout component
4. Create API calls in `src/lib/api.ts` if needed

**To add a new feature**:
1. Define types in `src/types/`
2. Create API functions in `src/lib/api.ts`
3. Create/update store in `src/stores/` if state needed
4. Create/update components in `src/components/`
5. Update views to use new components

**To modify styling**:
1. Update Tailwind classes in component templates
2. Modify `tailwind.config.ts` for theme changes
3. Add custom CSS in component `<style>` blocks if needed
