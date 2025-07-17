import { useEffect, useState } from 'react';

/**
 * Custom hook that debounces a value, delaying updates until the specified delay has passed
 * since the last change. This is useful for preventing excessive API calls on rapid filter changes.
 * 
 * @param value - The value to debounce
 * @param delay - The debounce delay in milliseconds
 * @returns The debounced value
 */
export function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState<T>(value);

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);

    return () => {
      clearTimeout(handler);
    };
  }, [value, delay]);

  return debouncedValue;
}

/**
 * Custom hook that provides debounced input handling for form controls.
 * Returns both the immediate value (for UI) and debounced value (for API calls).
 * 
 * @param initialValue - Initial input value
 * @param delay - Debounce delay in milliseconds (default: 500ms)
 * @returns Object with immediate value, debounced value, and setter function
 */
export function useDebouncedInput(initialValue: string, delay: number = 500) {
  const [value, setValue] = useState(initialValue);
  const debouncedValue = useDebounce(value, delay);

  return {
    value,           // Immediate value for UI input
    debouncedValue,  // Debounced value for API calls
    setValue,        // Function to update the value
  };
}