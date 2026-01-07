// Test file: React/TypeScript detection patterns
// Should detect: React, TypeScript, possibly Next.js patterns

import React, { useState, useEffect, useCallback, useMemo } from 'react';
import type { FC, ReactNode } from 'react';
import { createRoot } from 'react-dom/client';

interface Props {
  title: string;
  children: ReactNode;
}

const MyComponent: FC<Props> = ({ title, children }) => {
  const [count, setCount] = useState(0);
  const [data, setData] = useState<string[]>([]);

  useEffect(() => {
    // Fetch data on mount
    fetch('/api/data')
      .then(res => res.json())
      .then(setData);
  }, []);

  const handleClick = useCallback(() => {
    setCount(prev => prev + 1);
  }, []);

  const memoizedValue = useMemo(() => {
    return data.filter(item => item.length > 3);
  }, [data]);

  return (
    <div className="container">
      <h1>{title}</h1>
      <button onClick={handleClick}>
        Count: {count}
      </button>
      <ul>
        {memoizedValue.map((item, i) => (
          <li key={i}>{item}</li>
        ))}
      </ul>
      {children}
    </div>
  );
};

// React 18 root API
const container = document.getElementById('root');
const root = createRoot(container!);
root.render(<MyComponent title="Test App">Hello</MyComponent>);

export default MyComponent;
