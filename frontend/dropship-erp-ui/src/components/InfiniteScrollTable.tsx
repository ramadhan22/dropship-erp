import { useState, useCallback, useMemo } from 'react';
import {
  Box,
  CircularProgress,
  Typography,
  Button,
} from '@mui/material';
import { useInfiniteQuery } from '@tanstack/react-query';
import VirtualizedTable from './VirtualizedTable';
import type { Column } from './SortableTable';

interface InfiniteScrollTableProps<T extends Record<string, any>> {
  columns: Column<T>[];
  queryKey: string[];
  queryFn: ({ pageParam }: { pageParam: number }) => Promise<{
    data: T[];
    total: number;
    hasNextPage: boolean;
  }>;
  height?: number;
  itemHeight?: number;
  pageSize?: number;
  emptyMessage?: string;
  enableSearch?: boolean;
  searchPlaceholder?: string;
}

export default function InfiniteScrollTable<T extends Record<string, any>>({
  columns,
  queryKey,
  queryFn,
  height = 600,
  itemHeight = 53,
  pageSize = 50,
  emptyMessage = 'No data available',
  enableSearch = false,
  searchPlaceholder = 'Search...',
}: InfiniteScrollTableProps<T>) {
  const [searchTerm, setSearchTerm] = useState('');

  const {
    data,
    error,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    status,
    refetch,
  } = useInfiniteQuery({
    queryKey: [...queryKey, searchTerm],
    queryFn: ({ pageParam = 0 }) => queryFn({ pageParam: pageParam as number }),
    getNextPageParam: (lastPage: any, pages: any) => {
      if (lastPage.hasNextPage) {
        return pages.length * pageSize;
      }
      return undefined;
    },
    initialPageParam: 0,
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: 3,
  });

  // Flatten all pages into a single array
  const allItems = useMemo(() => {
    return data?.pages.flatMap(page => page.data) ?? [];
  }, [data]);

  const totalCount = data?.pages[0]?.total ?? 0;

  const handleLoadMore = useCallback(() => {
    if (hasNextPage && !isFetchingNextPage) {
      fetchNextPage();
    }
  }, [hasNextPage, isFetchingNextPage, fetchNextPage]);

  const handleSearch = useCallback((term: string) => {
    setSearchTerm(term);
  }, []);

  if (status === 'pending') {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" height={height}>
        <CircularProgress />
      </Box>
    );
  }

  if (status === 'error') {
    return (
      <Box display="flex" flexDirection="column" alignItems="center" gap={2} height={height}>
        <Typography color="error">
          Error loading data: {error?.message || 'Unknown error'}
        </Typography>
        <Button variant="outlined" onClick={() => refetch()}>
          Retry
        </Button>
      </Box>
    );
  }

  return (
    <Box>
      {/* Search bar */}
      {enableSearch && (
        <Box mb={2}>
          <input
            type="text"
            placeholder={searchPlaceholder}
            value={searchTerm}
            onChange={(e) => handleSearch(e.target.value)}
            style={{
              width: '100%',
              padding: '8px 12px',
              border: '1px solid #ccc',
              borderRadius: '4px',
              fontSize: '14px',
            }}
          />
        </Box>
      )}

      {/* Data count */}
      <Box mb={2} display="flex" justifyContent="space-between" alignItems="center">
        <Typography variant="body2" color="textSecondary">
          {allItems.length > 0 
            ? `Showing ${allItems.length} of ${totalCount} items`
            : 'No items found'
          }
        </Typography>
        {hasNextPage && (
          <Button
            variant="outlined"
            size="small"
            onClick={handleLoadMore}
            disabled={isFetchingNextPage}
          >
            {isFetchingNextPage ? 'Loading...' : 'Load More'}
          </Button>
        )}
      </Box>

      {/* Virtualized Table */}
      <VirtualizedTable
        columns={columns}
        data={allItems}
        height={height}
        itemHeight={itemHeight}
        loading={status === 'pending'}
        emptyMessage={emptyMessage}
      />

      {/* Load more button at bottom */}
      {hasNextPage && (
        <Box mt={2} display="flex" justifyContent="center">
          <Button
            variant="contained"
            onClick={handleLoadMore}
            disabled={isFetchingNextPage}
            startIcon={isFetchingNextPage ? <CircularProgress size={20} /> : null}
          >
            {isFetchingNextPage ? 'Loading More...' : 'Load More Data'}
          </Button>
        </Box>
      )}

      {/* Show when all data is loaded */}
      {!hasNextPage && allItems.length > 0 && (
        <Box mt={2} display="flex" justifyContent="center">
          <Typography variant="caption" color="textSecondary">
            All {totalCount} items loaded
          </Typography>
        </Box>
      )}
    </Box>
  );
}