import { useEffect, useState, useCallback, useMemo } from "react";
import { useDebounce } from './useDebounce';

export interface FilterCondition {
  field: string;
  operator: string;
  value: any;
  values?: any[];
}

export interface FilterGroup {
  logic: string; // "AND" or "OR"
  conditions: FilterCondition[];
  groups?: FilterGroup[];
}

export interface SortCondition {
  field: string;
  direction: "asc" | "desc";
}

export interface FilterParams {
  filters?: FilterGroup;
  sort?: SortCondition[];
  pagination?: {
    page: number;
    page_size: number;
  };
}

export interface FetchParams {
  page: number;
  pageSize: number;
  filters?: FilterParams;
}

export interface QueryResult<T> {
  data: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export type Fetcher<T> = (
  params: FetchParams,
) => Promise<QueryResult<T> | { data: T[]; total: number }>;

export interface UseOptimizedServerPaginationOptions {
  defaultSize?: number;
  debounceDelay?: number;
  staleTime?: number;
  enableDebouncing?: boolean;
  retryAttempts?: number;
}

/**
 * Optimized server pagination hook with debouncing and better performance characteristics
 */
export default function useOptimizedServerPagination<T>(
  fetcher: Fetcher<T>,
  options: UseOptimizedServerPaginationOptions = {}
) {
  const {
    defaultSize = 20,
    debounceDelay = 500,
    enableDebouncing = true,
    retryAttempts = 2,
  } = options;

  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(defaultSize);
  const [data, setData] = useState<T[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [filters, setFilters] = useState<FilterParams | undefined>();
  const [sortConditions, setSortConditions] = useState<SortCondition[]>([]);

  // Debounce filters to prevent excessive API calls
  const debouncedFilters = useDebounce(filters, enableDebouncing ? debounceDelay : 0);

  // Memoize fetch parameters to prevent unnecessary re-renders
  const fetchParams = useMemo((): FetchParams => ({
    page,
    pageSize,
    filters: debouncedFilters ? {
      ...debouncedFilters,
      sort: sortConditions.length > 0 ? sortConditions : debouncedFilters.sort,
      pagination: { page, page_size: pageSize }
    } : {
      sort: sortConditions.length > 0 ? sortConditions : undefined,
      pagination: { page, page_size: pageSize }
    }
  }), [page, pageSize, debouncedFilters, sortConditions]);

  const load = useCallback(async (showLoading = true) => {
    if (showLoading) {
      setLoading(true);
    }
    
    let retries = 0;
    
    while (retries <= retryAttempts) {
      try {
        const res = await fetcher(fetchParams);
        
        // Handle both old and new response formats
        if ('page' in res && 'page_size' in res) {
          // New format with QueryResult
          setData(res.data);
          setTotal(res.total);
          const pages = Math.max(1, Math.ceil(res.total / pageSize));
          if (page > pages) {
            setPage(pages);
          }
        } else {
          // Legacy format
          setData(res.data);
          setTotal(res.total);
          const pages = Math.max(1, Math.ceil(res.total / pageSize));
          if (page > pages) {
            setPage(pages);
          }
        }
        setError(null);
        break; // Success, exit retry loop
      } catch (e: any) {
        retries++;
        if (retries > retryAttempts) {
          setError(e.message || 'Failed to load data');
          setData([]);
          setTotal(0);
        } else {
          // Wait before retrying (exponential backoff)
          await new Promise(resolve => setTimeout(resolve, 1000 * retries));
        }
      }
    }
    
    if (showLoading) {
      setLoading(false);
    }
  }, [fetchParams, fetcher, page, pageSize, retryAttempts]);

  // Only reload when debounced parameters actually change
  useEffect(() => {
    load();
  }, [load]);

  const applyFilters = useCallback((newFilters: FilterParams | undefined) => {
    setFilters(newFilters);
    setPage(1); // Reset to first page when filters change
  }, []);

  const applySort = useCallback((sort: SortCondition[]) => {
    setSortConditions(sort);
    setPage(1); // Reset to first page when sort changes
  }, []);

  const clearFilters = useCallback(() => {
    setFilters(undefined);
    setSortConditions([]);
    setPage(1);
  }, []);

  const reload = useCallback((showLoading = true) => {
    return load(showLoading);
  }, [load]);

  // Optimized pagination controls with memoization - returns props for component
  const paginationProps = useMemo(() => ({
    total,
    loading,
    pageSize,
    page,
    totalPages: Math.max(1, Math.ceil(total / pageSize)),
    onPageSizeChange: (newSize: number) => {
      setPageSize(newSize);
      setPage(1);
    },
    onPageChange: (newPage: number) => setPage(newPage),
  }), [total, loading, pageSize, page]);

  return { 
    data, 
    loading, 
    error, 
    page, 
    setPage, 
    pageSize, 
    setPageSize, 
    reload,
    filters,
    applyFilters,
    applySort,
    clearFilters,
    sortConditions,
    // Additional performance utilities
    isFiltersDebouncing: enableDebouncing && filters !== debouncedFilters,
    totalPages: Math.max(1, Math.ceil(total / pageSize)),
    paginationProps,
  };
}