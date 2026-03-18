# Architecture - Frontend

**Generated:** 2026-03-06
**Part:** frontend
**Type:** React SPA (Single Page Application)

---

## Architecture Pattern

**Component-based SPA** with centralized state and declarative data fetching:

```
Browser
  |
  v
[React Router] --> Route matching, auth guard (PrivateRoute)
  |
  v
[Pages]        --> Route-level components, layout composition
  |
  v
[Components]   --> Reusable UI (common/, transaction/, charts/)
  |
  v
[Hooks]        --> useAuth, useTransactions (TanStack Query wrappers)
  |
  v
[API Layer]    --> Axios client with JWT interceptor
  |
  v
[Backend REST API]
```

## Technology Stack

| Component | Technology | Version |
|-----------|-----------|---------|
| Language | TypeScript | 5.3 |
| Framework | React | 18.2 |
| Build Tool | Vite | 5.0 |
| Routing | React Router | 6.21 |
| State Management | Zustand | 4.4.7 |
| Server State | TanStack Query | 5.17.9 |
| HTTP Client | Axios | 1.6.5 |
| Styling | Tailwind CSS | 3.4.1 |
| Charts | Recharts | 2.10.3 |
| Forms | React Hook Form | 7.49.3 |
| Date Utils | date-fns | 3.2.0 |
| Unit Testing | Vitest | 1.2.1 |
| Component Testing | React Testing Library | 14.1.2 |
| E2E Testing | Playwright | 1.41.0 |

## Entry Point

**File:** `src/main.tsx` -> `src/App.tsx`

Bootstrap sequence:
1. Mount React to `#root`
2. Initialize `QueryClientProvider` (TanStack Query)
3. Set up `BrowserRouter` with routes
4. `useAuthStore().initAuth()` restores auth from localStorage
5. `PrivateRoute` guards authenticated pages

## Routing Structure

| Path | Page | Auth Required |
|------|------|---------------|
| `/login` | Login | No |
| `/register` | Register | No |
| `/dashboard` | Dashboard | Yes |
| `/transactions` | Transactions | Yes |
| `/upload` | Upload | Yes |
| `/charts` | Charts | Yes |
| `/settings` | Settings | Yes |
| `/` | Redirect to `/dashboard` | Yes |

## State Management

### Client State (Zustand)

**`authStore.ts`** - Auth state:
- `user`, `token`, `isAuthenticated`, `isLoading`
- `setAuth()` - Store JWT + user in localStorage and state
- `clearAuth()` / `logout()` - Clear localStorage and state
- `initAuth()` - Restore from localStorage on app load

### Server State (TanStack Query)

Used via custom hooks for all API data:
- Auto caching and refetching
- `staleTime: 0` (always refetch)
- `retry: 1`
- `refetchOnWindowFocus: false`

## API Client Layer

**`api/client.ts`** - Axios instance:
- Base URL from environment (`VITE_API_URL`)
- Request interceptor: Injects `Authorization: Bearer <token>` from Zustand store
- Response interceptor: Auto-logout on 401

**API modules:**
- `api/auth.ts` - login, register, me
- `api/transactions.ts` - CRUD, stats
- `api/categories.ts` - list categories

## Component Architecture

### Common Components
- `Button` - Styled button with variants (primary, secondary, danger)
- `Input` - Form input with label, error display
- `Modal` - Dialog overlay with close handling

### Domain Components
- `TransactionList` - Table with pagination controls
- `TransactionModal` - Add/edit form dialog
- `PieChart` - Category distribution (Recharts)
- `BarChart` - Monthly income/expense comparison (Recharts)

### Layout
- `Layout` - App shell with sidebar navigation and header

## Styling Approach

- **Tailwind CSS** utility classes for all styling
- `clsx` for conditional class merging
- No CSS modules or styled-components
- Global styles in `index.css` (Tailwind base/components/utilities)

## Testing Strategy

- **Unit tests:** `*.test.ts` / `*.test.tsx` files co-located with source
- **Component tests:** React Testing Library with jsdom
- **E2E tests:** Playwright in `tests/e2e/`
- **Test setup:** `src/tests/setup.ts` configures testing-library/jest-dom
- **Coverage target:** >= 75%
