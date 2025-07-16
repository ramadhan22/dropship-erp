import { Skeleton, Box, Card, CardContent } from '@mui/material';

interface SkeletonLoaderProps {
  variant?: 'table' | 'card' | 'form' | 'dashboard';
  rows?: number;
  columns?: number;
}

export const SkeletonLoader = ({ 
  variant = 'table', 
  rows = 5, 
  columns = 4 
}: SkeletonLoaderProps) => {
  
  if (variant === 'table') {
    return (
      <Box>
        {/* Table header */}
        <Box display="flex" gap={1} mb={1}>
          {Array.from({ length: columns }).map((_, index) => (
            <Skeleton key={index} variant="text" width="20%" height={40} />
          ))}
        </Box>
        
        {/* Table rows */}
        {Array.from({ length: rows }).map((_, rowIndex) => (
          <Box key={rowIndex} display="flex" gap={1} mb={1}>
            {Array.from({ length: columns }).map((_, colIndex) => (
              <Skeleton key={colIndex} variant="text" width="20%" height={30} />
            ))}
          </Box>
        ))}
      </Box>
    );
  }

  if (variant === 'card') {
    return (
      <Box display="flex" flexDirection="column" gap={2}>
        {Array.from({ length: rows }).map((_, index) => (
          <Card key={index}>
            <CardContent>
              <Skeleton variant="text" width="60%" height={24} sx={{ mb: 1 }} />
              <Skeleton variant="text" width="100%" height={16} sx={{ mb: 1 }} />
              <Skeleton variant="text" width="80%" height={16} />
            </CardContent>
          </Card>
        ))}
      </Box>
    );
  }

  if (variant === 'form') {
    return (
      <Box display="flex" flexDirection="column" gap={2}>
        {Array.from({ length: rows }).map((_, index) => (
          <Box key={index}>
            <Skeleton variant="text" width="30%" height={20} sx={{ mb: 1 }} />
            <Skeleton variant="rectangular" width="100%" height={56} />
          </Box>
        ))}
      </Box>
    );
  }

  if (variant === 'dashboard') {
    return (
      <Box>
        {/* Dashboard header */}
        <Skeleton variant="text" width="40%" height={40} sx={{ mb: 2 }} />
        
        {/* Stats cards */}
        <Box display="flex" gap={2} mb={3}>
          {Array.from({ length: 4 }).map((_, index) => (
            <Card key={index} sx={{ flex: 1 }}>
              <CardContent>
                <Skeleton variant="text" width="70%" height={20} sx={{ mb: 1 }} />
                <Skeleton variant="text" width="50%" height={32} />
              </CardContent>
            </Card>
          ))}
        </Box>
        
        {/* Chart area */}
        <Card>
          <CardContent>
            <Skeleton variant="text" width="30%" height={24} sx={{ mb: 2 }} />
            <Skeleton variant="rectangular" width="100%" height={300} />
          </CardContent>
        </Card>
      </Box>
    );
  }

  // Default fallback
  return (
    <Box>
      {Array.from({ length: rows }).map((_, index) => (
        <Skeleton key={index} variant="text" width="100%" height={40} sx={{ mb: 1 }} />
      ))}
    </Box>
  );
};