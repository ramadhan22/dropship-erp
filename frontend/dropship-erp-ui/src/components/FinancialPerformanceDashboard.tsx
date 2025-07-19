import { useState } from 'react';
import { Box, Typography, Card, CardContent } from '@mui/material';
import PerformanceDashboardLayout from './PerformanceDashboardLayout';
import DashboardFilterContainer, { type DashboardFilters } from './DashboardFilterContainer';
import ChartContainer from './ChartContainer';
import { colors, spacing } from '../theme/tokens';

// Revenue & Profit Analysis Component
function RevenueAndProfitAnalysis() {
  return (
    <Box>
      {/* Key Metrics Cards */}
      <Box sx={{ 
        display: 'grid', 
        gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr', md: 'repeat(4, 1fr)' },
        gap: spacing.md,
        mb: spacing.lg 
      }}>
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" sx={{ color: colors.primary, mb: spacing.sm }}>
              Total Revenue
            </Typography>
            <Typography variant="h4" sx={{ color: colors.text, mb: spacing.xs }}>
              ₹2,45,67,890
            </Typography>
            <Typography variant="body2" sx={{ color: colors.success }}>
              +12.5% from last month
            </Typography>
          </CardContent>
        </Card>
        
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" sx={{ color: colors.primary, mb: spacing.sm }}>
              Gross Profit
            </Typography>
            <Typography variant="h4" sx={{ color: colors.text, mb: spacing.xs }}>
              ₹1,23,45,678
            </Typography>
            <Typography variant="body2" sx={{ color: colors.success }}>
              +8.3% from last month
            </Typography>
          </CardContent>
        </Card>
        
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" sx={{ color: colors.primary, mb: spacing.sm }}>
              Profit Margin
            </Typography>
            <Typography variant="h4" sx={{ color: colors.text, mb: spacing.xs }}>
              50.2%
            </Typography>
            <Typography variant="body2" sx={{ color: colors.warning }}>
              -2.1% from last month
            </Typography>
          </CardContent>
        </Card>
        
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" sx={{ color: colors.primary, mb: spacing.sm }}>
              Net Profit
            </Typography>
            <Typography variant="h4" sx={{ color: colors.text, mb: spacing.xs }}>
              ₹98,76,543
            </Typography>
            <Typography variant="body2" sx={{ color: colors.success }}>
              +15.7% from last month
            </Typography>
          </CardContent>
        </Card>
      </Box>

      {/* Charts */}
      <Box sx={{ mt: spacing.lg }}>
        <ChartContainer
          title="Revenue Trend"
          subtitle="Monthly revenue comparison over the last 12 months"
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
              Revenue trend chart will be displayed here
            </Typography>
          </Box>
        </ChartContainer>

        <ChartContainer
          title="Profit Analysis"
          subtitle="Gross profit vs net profit comparison"
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
              Profit analysis chart will be displayed here
            </Typography>
          </Box>
        </ChartContainer>
      </Box>
    </Box>
  );
}

// Sales Performance Metrics Component
function SalesPerformanceMetrics() {
  return (
    <Box>
      {/* Sales Metrics Cards */}
      <Box sx={{ 
        display: 'grid', 
        gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr', md: 'repeat(3, 1fr)' },
        gap: spacing.md,
        mb: spacing.lg 
      }}>
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" sx={{ color: colors.primary, mb: spacing.sm }}>
              Total Orders
            </Typography>
            <Typography variant="h4" sx={{ color: colors.text, mb: spacing.xs }}>
              15,847
            </Typography>
            <Typography variant="body2" sx={{ color: colors.success }}>
              +18.2% from last month
            </Typography>
          </CardContent>
        </Card>
        
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" sx={{ color: colors.primary, mb: spacing.sm }}>
              Average Order Value
            </Typography>
            <Typography variant="h4" sx={{ color: colors.text, mb: spacing.xs }}>
              ₹1,548
            </Typography>
            <Typography variant="body2" sx={{ color: colors.warning }}>
              -3.5% from last month
            </Typography>
          </CardContent>
        </Card>
        
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" sx={{ color: colors.primary, mb: spacing.sm }}>
              Conversion Rate
            </Typography>
            <Typography variant="h4" sx={{ color: colors.text, mb: spacing.xs }}>
              3.24%
            </Typography>
            <Typography variant="body2" sx={{ color: colors.success }}>
              +0.8% from last month
            </Typography>
          </CardContent>
        </Card>
      </Box>

      <Box sx={{ mt: spacing.lg }}>
        <ChartContainer
          title="Sales Performance Overview"
          subtitle="Comprehensive sales metrics and trends"
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
              Sales performance dashboard will be displayed here
            </Typography>
          </Box>
        </ChartContainer>
      </Box>
    </Box>
  );
}

export default function FinancialPerformanceDashboard() {
  const [filters, setFilters] = useState<DashboardFilters>({
    dateFrom: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000), // 30 days ago
    dateTo: new Date(),
    stores: [],
    channels: [],
    currency: 'IDR',
    period: 'daily',
  });

  const tabs = [
    {
      label: 'Revenue & Profit Analysis',
      component: <RevenueAndProfitAnalysis />,
    },
    {
      label: 'Sales Performance Metrics',
      component: <SalesPerformanceMetrics />,
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
        title="Financial Performance Dashboard"
        tabs={tabs}
      />
    </Box>
  );
}