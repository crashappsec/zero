# Component Architecture Guide

## Principles

### 1. Single Responsibility
Each component should do one thing well. If a component is doing multiple things, split it.

**Signs of violation:**
- Component file > 300 lines
- Multiple unrelated state variables
- Component name needs "And" (e.g., `UserListAndFilter`)

### 2. Composition Over Configuration
Build complex UIs from simple, composable pieces rather than prop-heavy mega-components.

```tsx
// ❌ Configuration approach
<DataTable
  data={users}
  sortable
  filterable
  pagination
  selectable
  onSort={...}
  onFilter={...}
  onPageChange={...}
  onSelect={...}
  columns={...}
  // 20+ props
/>

// ✅ Composition approach
<DataTable data={users}>
  <DataTable.Toolbar>
    <DataTable.Search />
    <DataTable.Filter />
  </DataTable.Toolbar>
  <DataTable.Columns>
    <DataTable.Column field="name" sortable />
    <DataTable.Column field="email" />
  </DataTable.Columns>
  <DataTable.Pagination />
</DataTable>
```

### 3. Colocation
Keep related code together. Tests, styles, and types should live near their components.

```
components/
└── Button/
    ├── Button.tsx
    ├── Button.test.tsx
    ├── Button.styles.ts
    ├── Button.types.ts
    └── index.ts
```

## Component Categories

### Presentational Components
Pure UI components with no business logic.

```tsx
// Good presentational component
interface CardProps {
  title: string;
  children: React.ReactNode;
  variant?: 'default' | 'outlined';
}

export const Card = ({ title, children, variant = 'default' }: CardProps) => (
  <div className={`card card--${variant}`}>
    <h2 className="card__title">{title}</h2>
    <div className="card__content">{children}</div>
  </div>
);
```

### Container Components
Handle data fetching, state management, and business logic.

```tsx
// Container component
export const UserListContainer = () => {
  const { data: users, isLoading, error } = useUsers();

  if (isLoading) return <LoadingSpinner />;
  if (error) return <ErrorMessage error={error} />;

  return <UserList users={users} />;
};
```

### Layout Components
Structure and arrange other components.

```tsx
// Layout component
interface PageLayoutProps {
  sidebar?: React.ReactNode;
  children: React.ReactNode;
}

export const PageLayout = ({ sidebar, children }: PageLayoutProps) => (
  <div className="page-layout">
    {sidebar && <aside className="page-layout__sidebar">{sidebar}</aside>}
    <main className="page-layout__main">{children}</main>
  </div>
);
```

## Props Design

### Keep Props Minimal
Each prop should have a clear purpose. Avoid boolean props when possible.

```tsx
// ❌ Boolean props
<Button primary small disabled loading />

// ✅ Variant props
<Button variant="primary" size="small" state="loading" />
```

### Use Children for Content
Prefer `children` over content props for flexibility.

```tsx
// ❌ Content props
<Modal title="Confirm" body="Are you sure?" footer={<Button>OK</Button>} />

// ✅ Children composition
<Modal>
  <Modal.Header>Confirm</Modal.Header>
  <Modal.Body>Are you sure?</Modal.Body>
  <Modal.Footer>
    <Button>OK</Button>
  </Modal.Footer>
</Modal>
```

### Consistent Naming
- Event handlers: `onAction` (onClick, onSubmit, onChange)
- Boolean props: `isState` or `hasFeature` (isLoading, isOpen, hasError)
- Render props: `renderItem`, `renderHeader`

## State Management Decision Tree

```
Is state used by single component?
├── Yes → useState
└── No → Is state used by nearby children?
    ├── Yes → Lift state up + props
    └── No → Is state used by many components?
        ├── Yes (simple) → React Context
        ├── Yes (complex) → State library (Redux, Zustand)
        └── Is it server data? → React Query / SWR
```

## File Organization

### Feature-Based Structure
Group by feature, not by type.

```
src/
├── features/
│   ├── auth/
│   │   ├── components/
│   │   ├── hooks/
│   │   ├── api/
│   │   └── index.ts
│   ├── users/
│   └── products/
├── shared/
│   ├── components/
│   ├── hooks/
│   └── utils/
└── app/
    ├── routes/
    └── providers/
```

### Index Files for Public API
Export only what other features need.

```tsx
// features/auth/index.ts
export { LoginForm } from './components/LoginForm';
export { useAuth } from './hooks/useAuth';
export type { User } from './types';

// Keep internal components private
// Don't export: AuthContext, validateCredentials, etc.
```

## Common Patterns

### Render Props
For sharing logic with flexible rendering.

```tsx
<MousePosition>
  {({ x, y }) => <p>Position: {x}, {y}</p>}
</MousePosition>
```

### Higher-Order Components (HOCs)
For cross-cutting concerns (use sparingly, prefer hooks).

```tsx
const withAuth = (Component) => (props) => {
  const { user } = useAuth();
  if (!user) return <Redirect to="/login" />;
  return <Component {...props} user={user} />;
};
```

### Compound Components
For components that work together.

```tsx
<Tabs defaultValue="tab1">
  <Tabs.List>
    <Tabs.Tab value="tab1">Tab 1</Tabs.Tab>
    <Tabs.Tab value="tab2">Tab 2</Tabs.Tab>
  </Tabs.List>
  <Tabs.Panel value="tab1">Content 1</Tabs.Panel>
  <Tabs.Panel value="tab2">Content 2</Tabs.Panel>
</Tabs>
```

### Controlled vs Uncontrolled
Support both patterns when appropriate.

```tsx
interface InputProps {
  // Controlled
  value?: string;
  onChange?: (value: string) => void;
  // Uncontrolled
  defaultValue?: string;
}
```

## Anti-Patterns to Avoid

1. **Prop drilling** - Pass through 3+ levels → Use Context
2. **God components** - > 500 lines → Split into smaller pieces
3. **Smart/dumb naming** - Container/Presentational naming is outdated
4. **Over-abstraction** - Creating abstractions before needed
5. **Premature optimization** - Memoizing everything "just in case"
