# State Management Guide

## State Categories

### 1. UI State
Local component state for UI interactions.

**Examples:** Open/closed modals, active tab, form input values, hover states

**Solution:** `useState` or `useReducer`

```tsx
const [isOpen, setIsOpen] = useState(false);
```

### 2. Server State
Data fetched from APIs that needs caching, synchronization, and background updates.

**Examples:** User profiles, product lists, search results

**Solution:** React Query, SWR, or Apollo Client

```tsx
const { data, isLoading, error } = useQuery({
  queryKey: ['users'],
  queryFn: fetchUsers
});
```

### 3. Global Application State
State needed across many components.

**Examples:** User authentication, theme preferences, feature flags

**Solution:** React Context (simple) or Zustand/Redux (complex)

```tsx
const { user, logout } = useAuth();
```

### 4. URL State
State that should be reflected in the URL for sharing/bookmarking.

**Examples:** Search filters, pagination, selected items

**Solution:** URL search params with router integration

```tsx
const [searchParams, setSearchParams] = useSearchParams();
const filter = searchParams.get('filter');
```

### 5. Form State
Complex form handling with validation.

**Examples:** Multi-step forms, dynamic fields, real-time validation

**Solution:** React Hook Form, Formik, or custom hooks

```tsx
const { register, handleSubmit, errors } = useForm();
```

## Decision Framework

```
┌─────────────────────────────────────────────────────────────────┐
│                    Where does state belong?                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Is it server data (from API)?                                  │
│  ├── Yes → React Query / SWR                                    │
│  └── No ↓                                                       │
│                                                                 │
│  Should it be in the URL?                                       │
│  ├── Yes → URL state (useSearchParams)                          │
│  └── No ↓                                                       │
│                                                                 │
│  Used by only one component?                                    │
│  ├── Yes → useState                                             │
│  └── No ↓                                                       │
│                                                                 │
│  Used by parent and 1-2 children?                               │
│  ├── Yes → Lift state up                                        │
│  └── No ↓                                                       │
│                                                                 │
│  Used by deeply nested or distant components?                   │
│  ├── Simple state → React Context                               │
│  └── Complex state → Zustand / Redux Toolkit                    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## useState Patterns

### Basic State
```tsx
const [count, setCount] = useState(0);
```

### Lazy Initialization
For expensive initial values:
```tsx
const [data, setData] = useState(() => computeExpensiveValue());
```

### Functional Updates
When new state depends on previous:
```tsx
setCount(prev => prev + 1);
```

### Object State
Spread to update:
```tsx
const [user, setUser] = useState({ name: '', email: '' });
setUser(prev => ({ ...prev, name: 'John' }));
```

## useReducer Patterns

Use when:
- State has multiple sub-values
- Next state depends on previous
- Logic is complex or testable in isolation

```tsx
type State = { count: number; step: number };
type Action =
  | { type: 'increment' }
  | { type: 'decrement' }
  | { type: 'setStep'; payload: number };

function reducer(state: State, action: Action): State {
  switch (action.type) {
    case 'increment':
      return { ...state, count: state.count + state.step };
    case 'decrement':
      return { ...state, count: state.count - state.step };
    case 'setStep':
      return { ...state, step: action.payload };
    default:
      return state;
  }
}

const [state, dispatch] = useReducer(reducer, { count: 0, step: 1 });
```

## React Context

### When to Use
- Theme/locale preferences
- Authentication state
- Feature flags
- Any state needed by many components at different levels

### When NOT to Use
- Server state (use React Query)
- Frequently updating state (causes re-renders)
- State only needed by nearby components (lift state)

### Pattern
```tsx
// 1. Create context with type safety
interface AuthContextType {
  user: User | null;
  login: (credentials: Credentials) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | null>(null);

// 2. Create provider with logic
export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);

  const login = async (credentials: Credentials) => {
    const user = await api.login(credentials);
    setUser(user);
  };

  const logout = () => setUser(null);

  // Memoize to prevent unnecessary re-renders
  const value = useMemo(() => ({ user, login, logout }), [user]);

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}

// 3. Create typed hook
export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}
```

## Server State (React Query)

### Basic Query
```tsx
const { data, isLoading, error, refetch } = useQuery({
  queryKey: ['users', filters],
  queryFn: () => fetchUsers(filters),
  staleTime: 5 * 60 * 1000, // 5 minutes
});
```

### Mutations
```tsx
const mutation = useMutation({
  mutationFn: createUser,
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['users'] });
  },
});
```

### Optimistic Updates
```tsx
const mutation = useMutation({
  mutationFn: updateUser,
  onMutate: async (newUser) => {
    await queryClient.cancelQueries({ queryKey: ['users', newUser.id] });
    const previous = queryClient.getQueryData(['users', newUser.id]);
    queryClient.setQueryData(['users', newUser.id], newUser);
    return { previous };
  },
  onError: (err, newUser, context) => {
    queryClient.setQueryData(['users', newUser.id], context?.previous);
  },
});
```

## Zustand (Simple Global State)

```tsx
import { create } from 'zustand';

interface CartStore {
  items: CartItem[];
  addItem: (item: CartItem) => void;
  removeItem: (id: string) => void;
  clearCart: () => void;
}

const useCartStore = create<CartStore>((set) => ({
  items: [],
  addItem: (item) => set((state) => ({ items: [...state.items, item] })),
  removeItem: (id) => set((state) => ({
    items: state.items.filter(i => i.id !== id)
  })),
  clearCart: () => set({ items: [] }),
}));

// Usage
const { items, addItem } = useCartStore();
```

## Common Mistakes

### 1. Storing Derived State
```tsx
// ❌ Bad - redundant state
const [items, setItems] = useState([]);
const [filteredItems, setFilteredItems] = useState([]);

useEffect(() => {
  setFilteredItems(items.filter(i => i.active));
}, [items]);

// ✅ Good - derive in render
const [items, setItems] = useState([]);
const filteredItems = items.filter(i => i.active);
// Or with useMemo if expensive:
const filteredItems = useMemo(() => items.filter(i => i.active), [items]);
```

### 2. Unnecessary Context
```tsx
// ❌ Bad - context for prop drilling 2 levels
<GrandparentContext.Provider value={{ data }}>
  <Parent>
    <Child /> {/* needs data */}
  </Parent>
</GrandparentContext.Provider>

// ✅ Good - just pass props
<Parent data={data}>
  <Child data={data} />
</Parent>
```

### 3. Single Giant Context
```tsx
// ❌ Bad - everything in one context
<AppContext.Provider value={{ user, theme, cart, notifications, ... }}>

// ✅ Good - separate contexts by domain
<AuthProvider>
  <ThemeProvider>
    <CartProvider>
      <NotificationsProvider>
        {children}
```

### 4. Forgetting Memoization in Context
```tsx
// ❌ Bad - new object every render
<Context.Provider value={{ user, login, logout }}>

// ✅ Good - memoized
const value = useMemo(() => ({ user, login, logout }), [user]);
<Context.Provider value={value}>
```
