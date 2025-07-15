import React, { useState, useMemo, useCallback } from 'react';
import {
  Table,
  TableHead,
  TableRow,
  TableCell,
  TableSortLabel,
  Box,
  Typography,
} from '@mui/material';
import { FixedSizeList as List } from 'react-window';
import type { Column } from './SortableTable';

interface VirtualizedTableProps<T extends Record<string, any>> {
  columns: Column<T>[];
  data: T[];
  height?: number;
  itemHeight?: number;
  defaultSort?: { key: keyof T; direction?: 'asc' | 'desc' };
  onSortChange?: (key: keyof T, direction: 'asc' | 'desc') => void;
  loading?: boolean;
  emptyMessage?: string;
}

interface RowItemProps<T> {
  index: number;
  style: React.CSSProperties;
  data: {
    items: T[];
    columns: Column<T>[];
  };
}

// Row component for react-window
const RowItem = <T extends Record<string, any>>({ index, style, data }: RowItemProps<T>) => {
  const { items, columns } = data;
  const row = items[index];

  return (
    <div style={style}>
      <TableRow
        sx={{
          display: 'flex',
          '& > *': {
            flex: 1,
            borderBottom: '1px solid #e0e0e0',
            minHeight: '53px',
            display: 'flex',
            alignItems: 'center',
            padding: '8px 16px',
          },
        }}
      >
        {columns.map((col, colIndex) => (
          <TableCell
            key={colIndex}
            align={col.align}
            sx={{
              flex: 1,
              borderBottom: 'none',
              padding: '8px 16px',
            }}
          >
            {col.render
              ? col.render(col.key ? (row as any)[col.key] : undefined, row)
              : col.key
                ? String((row as any)[col.key])
                : null}
          </TableCell>
        ))}
      </TableRow>
    </div>
  );
};

export default function VirtualizedTable<T extends Record<string, any>>({
  columns,
  data,
  height = 400,
  itemHeight = 53,
  defaultSort,
  onSortChange,
  loading = false,
  emptyMessage = 'No data available',
}: VirtualizedTableProps<T>) {
  const [sortKey, setSortKey] = useState<keyof T | null>(
    defaultSort?.key ?? null,
  );
  const [direction, setDirection] = useState<'asc' | 'desc'>(
    defaultSort?.direction ?? 'asc',
  );

  const sortedData = useMemo(() => {
    if (onSortChange) {
      return data; // External sorting
    }

    if (!sortKey) {
      return data;
    }

    const sorted = [...data];
    sorted.sort((a, b) => {
      const aVal = a[sortKey];
      const bVal = b[sortKey];
      
      if (aVal === bVal) return 0;
      if (aVal == null) return -1;
      if (bVal == null) return 1;
      
      const comparison = aVal > bVal ? 1 : -1;
      return comparison * (direction === 'asc' ? 1 : -1);
    });
    
    return sorted;
  }, [data, sortKey, direction, onSortChange]);

  const handleSort = useCallback((key: keyof T) => {
    let dir: 'asc' | 'desc' = 'asc';
    if (sortKey === key) {
      dir = direction === 'asc' ? 'desc' : 'asc';
      setDirection(dir);
    } else {
      setSortKey(key);
      dir = 'asc';
      setDirection(dir);
    }
    
    if (onSortChange) {
      onSortChange(key, dir);
    }
  }, [sortKey, direction, onSortChange]);

  const itemData = useMemo(() => ({
    items: sortedData,
    columns,
  }), [sortedData, columns]);

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" height={height}>
        <Typography>Loading...</Typography>
      </Box>
    );
  }

  if (data.length === 0) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" height={height}>
        <Typography color="textSecondary">{emptyMessage}</Typography>
      </Box>
    );
  }

  return (
    <Box>
      {/* Table Header */}
      <Table size="small">
        <TableHead>
          <TableRow
            sx={{
              display: 'flex',
              '& > *': {
                flex: 1,
              },
            }}
          >
            {columns.map((col, index) => (
              <TableCell key={index} align={col.align} sx={{ flex: 1 }}>
                {col.key ? (
                  <TableSortLabel
                    active={sortKey === col.key}
                    direction={sortKey === col.key ? direction : 'asc'}
                    onClick={() => handleSort(col.key!)}
                  >
                    {col.label}
                  </TableSortLabel>
                ) : (
                  col.label
                )}
              </TableCell>
            ))}
          </TableRow>
        </TableHead>
      </Table>

      {/* Virtualized Table Body */}
      <Box
        sx={{
          height,
          border: '1px solid #e0e0e0',
          borderTop: 'none',
        }}
      >
        <List
          height={height}
          width="100%"
          itemCount={sortedData.length}
          itemSize={itemHeight}
          itemData={itemData as any}
        >
          {RowItem}
        </List>
      </Box>

      {/* Row count indicator */}
      <Box mt={1} display="flex" justifyContent="space-between" alignItems="center">
        <Typography variant="caption" color="textSecondary">
          Showing {sortedData.length} rows
        </Typography>
      </Box>
    </Box>
  );
}