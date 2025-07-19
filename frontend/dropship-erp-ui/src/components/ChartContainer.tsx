import { 
  Box, 
  Paper, 
  Typography, 
  CircularProgress, 
  Alert,
  IconButton,
  Tooltip
} from '@mui/material';
import { FullscreenOutlined, GetAppOutlined } from '@mui/icons-material';
import { colors, spacing } from '../theme/tokens';

interface ChartContainerProps {
  title: string;
  subtitle?: string;
  children: React.ReactNode;
  loading?: boolean;
  error?: string;
  height?: number | string;
  onFullscreen?: () => void;
  onExport?: () => void;
  actions?: React.ReactNode;
}

export default function ChartContainer({
  title,
  subtitle,
  children,
  loading = false,
  error,
  height = 400,
  onFullscreen,
  onExport,
  actions,
}: ChartContainerProps) {
  return (
    <Paper 
      elevation={1} 
      sx={{ 
        borderRadius: '8px', 
        overflow: 'hidden',
        mb: spacing.lg,
      }}
    >
      {/* Header */}
      <Box sx={{ 
        p: spacing.md, 
        borderBottom: 1, 
        borderColor: colors.divider,
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'flex-start'
      }}>
        <Box sx={{ flex: 1 }}>
          <Typography variant="h6" component="h3" sx={{ color: colors.text, mb: spacing.xs }}>
            {title}
          </Typography>
          {subtitle && (
            <Typography variant="body2" sx={{ color: colors.textSecondary }}>
              {subtitle}
            </Typography>
          )}
        </Box>
        
        <Box sx={{ display: 'flex', gap: spacing.xs }}>
          {actions}
          {onExport && (
            <Tooltip title="Export Data">
              <IconButton size="small" onClick={onExport}>
                <GetAppOutlined fontSize="small" />
              </IconButton>
            </Tooltip>
          )}
          {onFullscreen && (
            <Tooltip title="Fullscreen">
              <IconButton size="small" onClick={onFullscreen}>
                <FullscreenOutlined fontSize="small" />
              </IconButton>
            </Tooltip>
          )}
        </Box>
      </Box>

      {/* Content */}
      <Box sx={{ 
        position: 'relative',
        height: typeof height === 'number' ? `${height}px` : height,
        minHeight: 200
      }}>
        {loading && (
          <Box sx={{ 
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            backgroundColor: 'rgba(255, 255, 255, 0.8)',
            zIndex: 1
          }}>
            <CircularProgress />
          </Box>
        )}
        
        {error && (
          <Box sx={{ p: spacing.md }}>
            <Alert severity="error">
              {error}
            </Alert>
          </Box>
        )}
        
        {!loading && !error && (
          <Box sx={{ 
            height: '100%',
            width: '100%',
            p: spacing.md
          }}>
            {children}
          </Box>
        )}
      </Box>
    </Paper>
  );
}