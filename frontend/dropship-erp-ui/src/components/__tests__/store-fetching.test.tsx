import { renderHook, act } from '@testing-library/react';
import { useState, useCallback } from 'react';

// Simple test to verify the store fetching logic works correctly
describe('Store Fetching Logic', () => {
  function useStoresFetcher() {
    const [stores, setStores] = useState<{ store_id: number; nama_toko: string }[]>([]);
    const [storesLoading, setStoresLoading] = useState(false);
    const [storesError, setStoresError] = useState<string | null>(null);

    const fetchStores = useCallback(async () => {
      setStoresLoading(true);
      setStoresError(null);
      try {
        const response = await fetch("/api/stores/all");
        if (!response.ok) {
          throw new Error(`Failed to fetch stores: ${response.status} ${response.statusText}`);
        }
        const data = await response.json();
        setStores(data || []);
        
        if (!data || data.length === 0) {
          setStoresError("No stores found. Please create stores first.");
        }
      } catch (error) {
        console.error("Error fetching stores:", error);
        setStoresError(error instanceof Error ? error.message : "Failed to fetch stores");
      } finally {
        setStoresLoading(false);
      }
    }, []);

    return { stores, storesLoading, storesError, fetchStores };
  }

  test('should handle successful store fetching', async () => {
    // Mock successful API response
    global.fetch = jest.fn(() => Promise.resolve({
      ok: true,
      json: () => Promise.resolve([
        { store_id: 1, nama_toko: 'MR eStore Shopee' },
        { store_id: 2, nama_toko: 'MR Barista Gear' }
      ])
    })) as jest.Mock;

    const { result } = renderHook(() => useStoresFetcher());

    expect(result.current.stores).toEqual([]);
    expect(result.current.storesLoading).toBe(false);
    expect(result.current.storesError).toBe(null);

    await act(async () => {
      await result.current.fetchStores();
    });

    expect(result.current.stores).toEqual([
      { store_id: 1, nama_toko: 'MR eStore Shopee' },
      { store_id: 2, nama_toko: 'MR Barista Gear' }
    ]);
    expect(result.current.storesLoading).toBe(false);
    expect(result.current.storesError).toBe(null);
  });

  test('should handle empty stores response', async () => {
    // Mock empty API response
    global.fetch = jest.fn(() => Promise.resolve({
      ok: true,
      json: () => Promise.resolve([])
    })) as jest.Mock;

    const { result } = renderHook(() => useStoresFetcher());

    await act(async () => {
      await result.current.fetchStores();
    });

    expect(result.current.stores).toEqual([]);
    expect(result.current.storesLoading).toBe(false);
    expect(result.current.storesError).toBe("No stores found. Please create stores first.");
  });

  test('should handle API errors', async () => {
    // Mock API error
    global.fetch = jest.fn(() => Promise.resolve({
      ok: false,
      status: 500,
      statusText: 'Internal Server Error'
    })) as jest.Mock;

    const { result } = renderHook(() => useStoresFetcher());

    await act(async () => {
      await result.current.fetchStores();
    });

    expect(result.current.stores).toEqual([]);
    expect(result.current.storesLoading).toBe(false);
    expect(result.current.storesError).toBe("Failed to fetch stores: 500 Internal Server Error");
  });

  test('should handle network errors', async () => {
    // Mock network error
    global.fetch = jest.fn(() => Promise.reject(new Error('Network error'))) as jest.Mock;

    const { result } = renderHook(() => useStoresFetcher());

    await act(async () => {
      await result.current.fetchStores();
    });

    expect(result.current.stores).toEqual([]);
    expect(result.current.storesLoading).toBe(false);
    expect(result.current.storesError).toBe("Network error");
  });
});