# Accessibility (a11y) Guide

## WCAG 2.1 Overview

### Conformance Levels
- **Level A** - Minimum accessibility (must have)
- **Level AA** - Standard accessibility (target for most sites)
- **Level AAA** - Enhanced accessibility (specialized needs)

Most organizations target **WCAG 2.1 Level AA**.

## Core Principles (POUR)

### 1. Perceivable
Information must be presentable in ways users can perceive.

### 2. Operable
Interface components must be operable by all users.

### 3. Understandable
Information and operation must be understandable.

### 4. Robust
Content must be robust enough for assistive technologies.

## Common Issues and Fixes

### Images

```tsx
// ❌ Missing alt text
<img src="hero.jpg" />

// ✅ Descriptive alt text
<img src="hero.jpg" alt="Team collaborating around whiteboard" />

// ✅ Decorative image (empty alt)
<img src="decoration.svg" alt="" role="presentation" />
```

### Buttons and Links

```tsx
// ❌ Non-descriptive
<button onClick={handleSubmit}>Click here</button>
<a href="/docs">Read more</a>

// ✅ Descriptive
<button onClick={handleSubmit}>Submit application</button>
<a href="/docs">Read documentation</a>

// ❌ Icon-only button without label
<button onClick={toggleMenu}>
  <MenuIcon />
</button>

// ✅ With accessible label
<button onClick={toggleMenu} aria-label="Open navigation menu">
  <MenuIcon aria-hidden="true" />
</button>
```

### Forms

```tsx
// ❌ Input without label
<input type="email" placeholder="Email" />

// ✅ Properly labeled
<label htmlFor="email">Email address</label>
<input id="email" type="email" aria-describedby="email-hint" />
<span id="email-hint">We'll never share your email</span>

// ✅ Error handling
<label htmlFor="password">Password</label>
<input
  id="password"
  type="password"
  aria-invalid={hasError}
  aria-describedby="password-error"
/>
{hasError && (
  <span id="password-error" role="alert">
    Password must be at least 8 characters
  </span>
)}
```

### Headings

```tsx
// ❌ Skipped heading levels
<h1>Page Title</h1>
<h3>Section Title</h3> {/* Skipped h2! */}

// ✅ Proper hierarchy
<h1>Page Title</h1>
<h2>Section Title</h2>
<h3>Subsection Title</h3>

// ❌ Heading for styling only
<h3 className="large-text">Not actually a heading</h3>

// ✅ Use CSS for styling
<p className="large-text">Just styled text</p>
```

### Color Contrast

**Minimum contrast ratios (WCAG AA):**
- Normal text: 4.5:1
- Large text (18px+ or 14px+ bold): 3:1
- UI components: 3:1

```css
/* ❌ Low contrast */
.text {
  color: #999;  /* Gray on white = ~2.8:1 */
}

/* ✅ Sufficient contrast */
.text {
  color: #595959;  /* = 7:1 on white */
}
```

### Keyboard Navigation

```tsx
// ❌ Not keyboard accessible
<div onClick={handleClick}>Clickable thing</div>

// ✅ Keyboard accessible
<button onClick={handleClick}>Clickable thing</button>

// ✅ Custom component with keyboard support
<div
  role="button"
  tabIndex={0}
  onClick={handleClick}
  onKeyDown={(e) => {
    if (e.key === 'Enter' || e.key === ' ') {
      handleClick();
    }
  }}
>
  Custom button
</div>
```

### Focus Management

```tsx
// ❌ Focus outline removed without replacement
button:focus {
  outline: none;
}

// ✅ Custom focus indicator
button:focus {
  outline: none;
  box-shadow: 0 0 0 3px rgba(66, 153, 225, 0.6);
}

// ✅ Focus visible for keyboard only
button:focus-visible {
  outline: 2px solid #4299e1;
  outline-offset: 2px;
}
```

### Modal Dialogs

```tsx
// ✅ Accessible modal
<dialog
  ref={dialogRef}
  aria-labelledby="modal-title"
  aria-describedby="modal-description"
  onClose={handleClose}
>
  <h2 id="modal-title">Confirm Action</h2>
  <p id="modal-description">Are you sure you want to proceed?</p>
  <button onClick={handleConfirm}>Confirm</button>
  <button onClick={handleClose}>Cancel</button>
</dialog>

// Focus trap - focus should stay within modal
// Return focus to trigger element on close
```

### Skip Links

```tsx
// Add at start of page
<a href="#main-content" className="skip-link">
  Skip to main content
</a>

// Target element
<main id="main-content" tabIndex={-1}>
  ...
</main>
```

```css
.skip-link {
  position: absolute;
  top: -40px;
  left: 0;
  padding: 8px;
  background: #000;
  color: #fff;
  z-index: 100;
}

.skip-link:focus {
  top: 0;
}
```

## ARIA Best Practices

### Use Semantic HTML First
ARIA should supplement, not replace, semantic HTML.

```tsx
// ❌ ARIA when semantic HTML exists
<div role="button">Click me</div>
<div role="navigation">...</div>

// ✅ Semantic HTML
<button>Click me</button>
<nav>...</nav>
```

### Common ARIA Patterns

```tsx
// Live regions for dynamic content
<div aria-live="polite" aria-atomic="true">
  {statusMessage}
</div>

// Loading state
<button aria-busy={isLoading} disabled={isLoading}>
  {isLoading ? 'Saving...' : 'Save'}
</button>

// Expanded/collapsed
<button
  aria-expanded={isExpanded}
  aria-controls="panel-content"
>
  Toggle Panel
</button>
<div id="panel-content" hidden={!isExpanded}>
  Panel content
</div>

// Current page in navigation
<nav>
  <a href="/" aria-current="page">Home</a>
  <a href="/about">About</a>
</nav>

// Required fields
<input aria-required="true" />
```

## Testing Checklist

### Keyboard Testing
- [ ] All interactive elements are focusable
- [ ] Focus order is logical
- [ ] Focus is visible
- [ ] No keyboard traps
- [ ] Escape closes modals/menus
- [ ] Skip link present and works

### Screen Reader Testing
- [ ] All images have appropriate alt text
- [ ] Form fields have labels
- [ ] Headings are hierarchical
- [ ] Links/buttons have descriptive text
- [ ] Dynamic content announced
- [ ] Error messages announced

### Visual Testing
- [ ] Color contrast meets requirements
- [ ] Text resizes to 200% without loss
- [ ] Content reflows at 320px width
- [ ] Focus indicators visible
- [ ] No information conveyed by color alone

## Tools

### Automated Testing
- **axe-core** - Automated accessibility testing
- **jest-axe** - Jest matcher for axe
- **eslint-plugin-jsx-a11y** - ESLint rules for JSX

```tsx
// Jest with axe
import { axe, toHaveNoViolations } from 'jest-axe';

expect.extend(toHaveNoViolations);

it('has no accessibility violations', async () => {
  const { container } = render(<MyComponent />);
  const results = await axe(container);
  expect(results).toHaveNoViolations();
});
```

### Browser Extensions
- axe DevTools
- WAVE
- Lighthouse (Chrome DevTools)

### Screen Readers
- VoiceOver (macOS) - Cmd + F5
- NVDA (Windows) - Free
- JAWS (Windows) - Enterprise
