import { useInfiniteQuery, useQuery, useQueryClient } from '@tanstack/react-query';
import { listDropshipPurchases, sumDropshipPurchases } from '../api';
import type { DropshipPurchase } from '../types';

// Query keys factory for consistent caching
export const dropshipQueryKeys = {
  all: ['dropship'] as const,
  purchases: () => [...dropshipQueryKeys.all, 'purchases'] as const,
  purchaseList: (filters: PurchaseFilters) => [...dropshipQueryKeys.purchases(), 'list', filters] as const,
  summary: (filters: PurchaseFilters) => [...dropshipQueryKeys.all, 'summary', filters] as const,
  details: (id: string) => [...dropshipQueryKeys.purchases(), 'details', id] as const,
};

export interface PurchaseFilters {
  channel?: string;
  store?: string;
  from?: string;
  to?: string;
  orderNo?: string;
  sortBy?: string;
  sortDirection?: 'asc' | 'desc';
}

// Optimized hook for infinite scrolling dropship purchases
export function useInfiniteDropshipPurchases(
  filters: PurchaseFilters,
  pageSize: number = 50,
) {
  return useInfiniteQuery({
    queryKey: dropshipQueryKeys.purchaseList(filters),
    queryFn: async ({ pageParam = 0 }) => {
      const response = await listDropshipPurchases({
        ...filters,
        page: Math.floor((pageParam as number) / pageSize),
        page_size: pageSize,
      });
      
      return {
        data: response.data.data,
        total: response.data.total,
        hasNextPage: response.data.data.length === pageSize,
        nextOffset: (pageParam as number) + pageSize,
      };
    },
    getNextPageParam: (lastPage: any) => {
      return lastPage.hasNextPage ? lastPage.nextOffset : undefined;
    },
    initialPageParam: 0,
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 30 * 60 * 1000, // 30 minutes (was cacheTime)
    retry: (failureCount, error: any) => {
      // Don't retry on 4xx errors
      if (error?.response?.status >= 400 && error?.response?.status < 500) {
        return false;
      }
      return failureCount < 3;
    },
  });
}

// Optimized hook for dropship purchase summary with intelligent caching
export function useDropshipPurchaseSummary(filters: PurchaseFilters) {
  return useQuery({
    queryKey: dropshipQueryKeys.summary(filters),
    queryFn: async () => {
      const response = await sumDropshipPurchases(filters);
      return response;
    },
    staleTime: 10 * 60 * 1000, // 10 minutes for summary data
    gcTime: 60 * 60 * 1000, // 1 hour
    retry: 2,
  });
}

// Hook for prefetching next page of purchases
export function usePrefetchNextPurchases() {
  const queryClient = useQueryClient();
  
  return (filters: PurchaseFilters, currentOffset: number, pageSize: number = 50) => {
    const nextOffset = currentOffset + pageSize;
    
    queryClient.prefetchInfiniteQuery({
      queryKey: dropshipQueryKeys.purchaseList(filters),
      queryFn: async ({ pageParam = nextOffset }) => {
        const response = await listDropshipPurchases({
          ...filters,
          page: Math.floor((pageParam as number) / pageSize),
          page_size: pageSize,
        });
        
        return {
          data: response.data.data,
          total: response.data.total,
          hasNextPage: response.data.data.length === pageSize,
          nextOffset: (pageParam as number) + pageSize,
        };
      },
      initialPageParam: nextOffset,
      staleTime: 5 * 60 * 1000,
    });
  };
}

// Hook for optimistic updates and cache invalidation
export function useDropshipPurchaseMutations() {
  const queryClient = useQueryClient();
  
  const invalidatePurchases = (filters?: Partial<PurchaseFilters>) => {
    if (filters) {
      queryClient.invalidateQueries({
        queryKey: dropshipQueryKeys.purchaseList(filters as PurchaseFilters),
      });
    } else {
      queryClient.invalidateQueries({
        queryKey: dropshipQueryKeys.purchases(),
      });
    }
  };
  
  const invalidateSummary = (filters?: Partial<PurchaseFilters>) => {
    if (filters) {
      queryClient.invalidateQueries({
        queryKey: dropshipQueryKeys.summary(filters as PurchaseFilters),
      });
    } else {
      queryClient.invalidateQueries({
        queryKey: dropshipQueryKeys.all,
        predicate: (query) => query.queryKey.includes('summary'),
      });
    }
  };
  
  const optimisticUpdatePurchase = (purchase: DropshipPurchase) => {
    // Update the cache optimistically for better UX
    queryClient.setQueriesData(
      { queryKey: dropshipQueryKeys.purchases() },
      (oldData: any) => {
        if (!oldData) return oldData;
        
        // Update the purchase in infinite query data
        const newPages = oldData.pages.map((page: any) => ({
          ...page,
          data: page.data.map((item: DropshipPurchase) =>
            item.kode_pesanan === purchase.kode_pesanan ? purchase : item
          ),
        }));
        
        return { ...oldData, pages: newPages };
      }
    );
  };
  
  return {
    invalidatePurchases,
    invalidateSummary,
    optimisticUpdatePurchase,
  };
}

// Memoized selector hooks for better performance
export function useDropshipPurchaseStats(
  purchases: DropshipPurchase[] | undefined
) {
  return {
    totalCount: purchases?.length ?? 0,
    totalValue: purchases?.reduce((sum, p) => sum + (p.total_transaksi || 0), 0) ?? 0,
    averageValue: purchases?.length 
      ? (purchases.reduce((sum, p) => sum + (p.total_transaksi || 0), 0) / purchases.length)
      : 0,
    uniqueStores: new Set(purchases?.map(p => p.nama_toko) ?? []).size,
  };
}