import { useState } from 'react';
import { Box, Typography, Card, CardContent } from '@mui/material';
import PerformanceDashboardLayout from './PerformanceDashboardLayout';
import DashboardFilterContainer, { type DashboardFilters } from './DashboardFilterContainer';
import ChartContainer from './ChartContainer';
import { colors, spacing } from '../theme/tokens';

// Shopee Performance Analytics Component
function ShopeePerformanceAnalytics() {
  return (
    <Box>
      {/* Shopee Metrics Cards */}
      <Box sx={{ 
        display: 'grid', 
        gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr', md: 'repeat(4, 1fr)' },
        gap: spacing.md,
        mb: spacing.lg 
      }}>
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" sx={{ color: colors.primary, mb: spacing.sm }}>
              Shopee Orders
            </Typography>
            <Typography variant="h4" sx={{ color: colors.text, mb: spacing.xs }}>
              9,847
            </Typography>
            <Typography variant="body2" sx={{ color: colors.success }}>
              +19.5% from last month
            </Typography>
          </CardContent>
        </Card>
        
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" sx={{ color: colors.primary, mb: spacing.sm }}>
              Shopee Revenue
            </Typography>
            <Typography variant="h4" sx={{ color: colors.text, mb: spacing.xs }}>
              ₹1,67,89,234
            </Typography>
            <Typography variant="body2" sx={{ color: colors.success }}>
              +24.1% from last month
            </Typography>
          </CardContent>
        </Card>
        
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" sx={{ color: colors.primary, mb: spacing.sm }}>
              Commission Fees
            </Typography>
            <Typography variant="h4" sx={{ color: colors.text, mb: spacing.xs }}>
              ₹8,39,461
            </Typography>
            <Typography variant="body2" sx={{ color: colors.warning }}>
              5% of total revenue
            </Typography>
          </CardContent>
        </Card>
        
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" sx={{ color: colors.primary, mb: spacing.sm }}>
              Return Rate
            </Typography>
            <Typography variant="h4" sx={{ color: colors.text, mb: spacing.xs }}>
              2.8%
            </Typography>
            <Typography variant="body2" sx={{ color: colors.success }}>
              -0.5% from last month
            </Typography>
          </CardContent>
        </Card>
      </Box>

      {/* Charts */}
      <Box sx={{ mt: spacing.lg }}>
        <ChartContainer
          title="Shopee Sales Performance"
          subtitle="Daily sales trends and order volume on Shopee platform"
          height={350}
        >
          <Box sx={{ 
            height: '100%', 
            display: 'flex', 
            alignItems: 'center', 
            justifyContent: 'center',
            backgroundColor: colors.surface,
            borderRadius: '4px'
          }}>
            <Typography variant="body1" sx={{ color: colors.textSecondary }}>
              Shopee performance analytics chart will be displayed here
            </Typography>
          </Box>
        </ChartContainer>

        <ChartContainer
          title="Fee Structure Analysis"
          subtitle="Breakdown of Shopee fees and commission structure"
          height={350}
        >
          <Box sx={{ 
            height: '100%', 
            display: 'flex', 
            alignItems: 'center', 
            justifyContent: 'center',
            backgroundColor: colors.surface,
            borderRadius: '4px'
          }}>
            <Typography variant="body1" sx={{ color: colors.textSecondary }}>
              Fee structure analysis chart will be displayed here
            </Typography>
          </Box>
        </ChartContainer>
      </Box>
    </Box>
  );
}

// Multi-channel Comparison Component
function MultiChannelComparison() {
  return (
    <Box>
      {/* Channel Comparison Cards */}
      <Box sx={{ 
        display: 'grid', 
        gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr', md: 'repeat(3, 1fr)' },
        gap: spacing.md,
        mb: spacing.lg 
      }}>
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" sx={{ color: colors.primary, mb: spacing.sm }}>
              Shopee Share
            </Typography>
            <Typography variant="h4" sx={{ color: colors.text, mb: spacing.xs }}>
              68.5%
            </Typography>
            <Typography variant="body2" sx={{ color: colors.success }}>
              Highest performing channel
            </Typography>
          </CardContent>
        </Card>
        
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" sx={{ color: colors.primary, mb: spacing.sm }}>
              Tokopedia Share
            </Typography>
            <Typography variant="h4" sx={{ color: colors.text, mb: spacing.xs }}>
              21.3%
            </Typography>
            <Typography variant="body2" sx={{ color: colors.info }}>
              Second largest channel
            </Typography>
          </CardContent>
        </Card>
        
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" sx={{ color: colors.primary, mb: spacing.sm }}>
              Other Channels
            </Typography>
            <Typography variant="h4" sx={{ color: colors.text, mb: spacing.xs }}>
              10.2%
            </Typography>
            <Typography variant="body2" sx={{ color: colors.textSecondary }}>
              Lazada, Bukalapak, etc.
            </Typography>
          </CardContent>
        </Card>
      </Box>

      <Box sx={{ mt: spacing.lg }}>
        <ChartContainer
          title="Channel Performance Comparison"
          subtitle="Revenue and order comparison across all marketplace channels"
          height={400}
        >
          <Box sx={{ 
            height: '100%', 
            display: 'flex', 
            alignItems: 'center', 
            justifyContent: 'center',
            backgroundColor: colors.surface,
            borderRadius: '4px'
          }}>
            <Typography variant="body1" sx={{ color: colors.textSecondary }}>
              Multi-channel comparison dashboard will be displayed here
            </Typography>
          </Box>
        </ChartContainer>

        <ChartContainer
          title="Channel Efficiency Metrics"
          subtitle="Conversion rates, average order values, and customer satisfaction by channel"
          height={350}
        >
          <Box sx={{ 
            height: '100%', 
            display: 'flex', 
            alignItems: 'center', 
            justifyContent: 'center',
            backgroundColor: colors.surface,
            borderRadius: '4px'
          }}>
            <Typography variant="body1" sx={{ color: colors.textSecondary }}>
              Channel efficiency metrics will be displayed here
            </Typography>
          </Box>
        </ChartContainer>
      </Box>
    </Box>
  );
}

export default function MarketplaceIntegrationDashboard() {
  const [filters, setFilters] = useState<DashboardFilters>({
    dateFrom: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000), // 30 days ago
    dateTo: new Date(),
    stores: [],
    channels: ['shopee', 'tokopedia'], // Default to main channels
    currency: 'IDR',
    period: 'daily',
  });

  const tabs = [
    {
      label: 'Shopee Performance Analytics',
      component: <ShopeePerformanceAnalytics />,
    },
    {
      label: 'Multi-channel Comparison',
      component: <MultiChannelComparison />,
    },
  ];

  return (
    <Box>
      <DashboardFilterContainer
        filters={filters}
        onFiltersChange={setFilters}
        showCurrencyToggle={true}
        showPeriodToggle={true}
      />
      
      <PerformanceDashboardLayout
        title="Marketplace Integration Dashboard"
        tabs={tabs}
      />
    </Box>
  );
}