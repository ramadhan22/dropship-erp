import { useEffect, useState } from 'react';
import { loadingEmitter } from './loadingEmitter';

export default function useLoading(): boolean {
  const [count, setCount] = useState(0);

  useEffect(() => {
    const handler = (e: Event) => {
      setCount((e as CustomEvent<number>).detail);
    };
    loadingEmitter.addEventListener('change', handler);
    return () => loadingEmitter.removeEventListener('change', handler);
  }, []);

  return count > 0;
}
