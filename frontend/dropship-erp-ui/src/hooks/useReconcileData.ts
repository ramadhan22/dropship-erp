import { useQuery, useQueryClient } from '@tanstack/react-query';
import { listCandidates } from '../api/reconcile';
import { listOrderDetails } from '../api/index';
import type { ReconcileCandidate } from '../types';

// Query keys factory for consistent caching
export const reconcileQueryKeys = {
  all: ['reconcile'] as const,
  candidates: () => [...reconcileQueryKeys.all, 'candidates'] as const,
  candidateList: (filters: ReconcileCandidateFilters) => 
    [...reconcileQueryKeys.candidates(), 'list', filters] as const,
  orderDetails: () => [...reconcileQueryKeys.all, 'orderDetails'] as const,
  orderDetailList: (filters: OrderDetailFilters) => 
    [...reconcileQueryKeys.orderDetails(), 'list', filters] as const,
};

export interface ReconcileCandidateFilters {
  shop?: string;
  order?: string;
  status?: string;
  from?: string;
  to?: string;
  page?: number;
  pageSize?: number;
}

export interface OrderDetailFilters {
  store?: string;
  order?: string;
  page?: number;
  pageSize?: number;
}

/**
 * Optimized hook for fetching reconcile candidates with caching and error handling
 */
export function useReconcileCandidates(filters: ReconcileCandidateFilters) {
  return useQuery({
    queryKey: reconcileQueryKeys.candidateList(filters),
    queryFn: async () => {
      const response = await listCandidates(
        filters.shop || '',
        filters.order || '',
        filters.status || '',
        filters.from || '',
        filters.to || '',
        filters.page || 1,
        filters.pageSize || 20,
        { headers: {  } } // Disable global loading spinner
      );
      return response.data;
    },
    staleTime: 2 * 60 * 1000, // 2 minutes - shorter for real-time data
    gcTime: 10 * 60 * 1000, // 10 minutes
    retry: (failureCount, error: any) => {
      // Don't retry on 4xx errors
      if (error?.response?.status >= 400 && error?.response?.status < 500) {
        return false;
      }
      return failureCount < 2; // Reduced retries for faster failure
    },
    enabled: !!(filters.from && filters.to), // Only fetch when date range is set
  });
}

/**
 * Optimized hook for fetching Shopee order details with caching
 */
export function useShopeeOrderDetails(filters: OrderDetailFilters) {
  return useQuery({
    queryKey: reconcileQueryKeys.orderDetailList(filters),
    queryFn: async () => {
      const response = await listOrderDetails({
        store: filters.store,
        order: filters.order,
        page: filters.page || 1,
        page_size: filters.pageSize || 20,
      });
      return response.data;
    },
    staleTime: 5 * 60 * 1000, // 5 minutes for order details
    gcTime: 30 * 60 * 1000, // 30 minutes
    retry: 2,
    enabled: true, // Always enabled, will return empty results if no filters
  });
}

/**
 * Hook for cache management and optimistic updates
 */
export function useReconcileMutations() {
  const queryClient = useQueryClient();
  
  const invalidateCandidates = (filters?: Partial<ReconcileCandidateFilters>) => {
    if (filters) {
      queryClient.invalidateQueries({
        queryKey: reconcileQueryKeys.candidateList(filters as ReconcileCandidateFilters),
      });
    } else {
      queryClient.invalidateQueries({
        queryKey: reconcileQueryKeys.candidates(),
      });
    }
  };
  
  const invalidateOrderDetails = (filters?: Partial<OrderDetailFilters>) => {
    if (filters) {
      queryClient.invalidateQueries({
        queryKey: reconcileQueryKeys.orderDetailList(filters as OrderDetailFilters),
      });
    } else {
      queryClient.invalidateQueries({
        queryKey: reconcileQueryKeys.orderDetails(),
      });
    }
  };

  const optimisticUpdateCandidate = (candidate: ReconcileCandidate) => {
    // Update the cache optimistically for better UX
    queryClient.setQueriesData(
      { queryKey: reconcileQueryKeys.candidates() },
      (oldData: any) => {
        if (!oldData?.data) return oldData;
        
        const newData = oldData.data.map((item: ReconcileCandidate) =>
          item.kode_pesanan === candidate.kode_pesanan ? candidate : item
        );
        
        return { ...oldData, data: newData };
      }
    );
  };
  
  return {
    invalidateCandidates,
    invalidateOrderDetails,
    optimisticUpdateCandidate,
  };
}

/**
 * Hook to prefetch next page of data for smoother pagination
 */
export function usePrefetchReconcileData() {
  const queryClient = useQueryClient();
  
  const prefetchCandidates = (filters: ReconcileCandidateFilters) => {
    const nextPageFilters = { ...filters, page: (filters.page || 1) + 1 };
    
    queryClient.prefetchQuery({
      queryKey: reconcileQueryKeys.candidateList(nextPageFilters),
      queryFn: async () => {
        const response = await listCandidates(
          nextPageFilters.shop || '',
          nextPageFilters.order || '',
          nextPageFilters.status || '',
          nextPageFilters.from || '',
          nextPageFilters.to || '',
          nextPageFilters.page || 1,
          nextPageFilters.pageSize || 20,
          { headers: {  } }
        );
        return response.data;
      },
      staleTime: 2 * 60 * 1000,
    });
  };
  
  return { prefetchCandidates };
}