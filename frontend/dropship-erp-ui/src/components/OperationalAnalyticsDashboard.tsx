import { useState } from 'react';
import { Box, Typography, Card, CardContent } from '@mui/material';
import PerformanceDashboardLayout from './PerformanceDashboardLayout';
import DashboardFilterContainer, { type DashboardFilters } from './DashboardFilterContainer';
import ChartContainer from './ChartContainer';
import { colors, spacing } from '../theme/tokens';

// Dropship Purchase Analysis Component
function DropshipPurchaseAnalysis() {
  return (
    <Box>
      {/* Purchase Metrics Cards */}
      <Box sx={{ 
        display: 'grid', 
        gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr', md: 'repeat(4, 1fr)' },
        gap: spacing.md,
        mb: spacing.lg 
      }}>
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" sx={{ color: colors.primary, mb: spacing.sm }}>
              Total Purchases
            </Typography>
            <Typography variant="h4" sx={{ color: colors.text, mb: spacing.xs }}>
              8,943
            </Typography>
            <Typography variant="body2" sx={{ color: colors.success }}>
              +22.1% from last month
            </Typography>
          </CardContent>
        </Card>
        
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" sx={{ color: colors.primary, mb: spacing.sm }}>
              Purchase Value
            </Typography>
            <Typography variant="h4" sx={{ color: colors.text, mb: spacing.xs }}>
              ₹1,89,45,632
            </Typography>
            <Typography variant="body2" sx={{ color: colors.success }}>
              +15.8% from last month
            </Typography>
          </CardContent>
        </Card>
        
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" sx={{ color: colors.primary, mb: spacing.sm }}>
              Average Purchase
            </Typography>
            <Typography variant="h4" sx={{ color: colors.text, mb: spacing.xs }}>
              ₹2,119
            </Typography>
            <Typography variant="body2" sx={{ color: colors.warning }}>
              -4.2% from last month
            </Typography>
          </CardContent>
        </Card>
        
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" sx={{ color: colors.primary, mb: spacing.sm }}>
              Pending Orders
            </Typography>
            <Typography variant="h4" sx={{ color: colors.text, mb: spacing.xs }}>
              156
            </Typography>
            <Typography variant="body2" sx={{ color: colors.error }}>
              +12 from yesterday
            </Typography>
          </CardContent>
        </Card>
      </Box>

      {/* Charts */}
      <Box sx={{ mt: spacing.lg }}>
        <ChartContainer
          title="Dropship Purchase Trends"
          subtitle="Daily purchase volume and value analysis"
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
              Dropship purchase trend chart will be displayed here
            </Typography>
          </Box>
        </ChartContainer>

        <ChartContainer
          title="Purchase Status Distribution"
          subtitle="Breakdown of purchase statuses across different suppliers"
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
              Purchase status distribution chart will be displayed here
            </Typography>
          </Box>
        </ChartContainer>
      </Box>
    </Box>
  );
}

// Reconciliation Dashboard Component
function ReconciliationDashboard() {
  return (
    <Box>
      {/* Reconciliation Metrics Cards */}
      <Box sx={{ 
        display: 'grid', 
        gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr', md: 'repeat(3, 1fr)' },
        gap: spacing.md,
        mb: spacing.lg 
      }}>
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" sx={{ color: colors.primary, mb: spacing.sm }}>
              Reconciled Transactions
            </Typography>
            <Typography variant="h4" sx={{ color: colors.text, mb: spacing.xs }}>
              12,547
            </Typography>
            <Typography variant="body2" sx={{ color: colors.success }}>
              94.2% success rate
            </Typography>
          </CardContent>
        </Card>
        
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" sx={{ color: colors.primary, mb: spacing.sm }}>
              Pending Reconciliation
            </Typography>
            <Typography variant="h4" sx={{ color: colors.text, mb: spacing.xs }}>
              342
            </Typography>
            <Typography variant="body2" sx={{ color: colors.warning }}>
              Requires attention
            </Typography>
          </CardContent>
        </Card>
        
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" sx={{ color: colors.primary, mb: spacing.sm }}>
              Discrepancies Found
            </Typography>
            <Typography variant="h4" sx={{ color: colors.text, mb: spacing.xs }}>
              23
            </Typography>
            <Typography variant="body2" sx={{ color: colors.error }}>
              Needs manual review
            </Typography>
          </CardContent>
        </Card>
      </Box>

      <Box sx={{ mt: spacing.lg }}>
        <ChartContainer
          title="Reconciliation Performance"
          subtitle="Daily reconciliation processing and accuracy metrics"
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
              Reconciliation performance dashboard will be displayed here
            </Typography>
          </Box>
        </ChartContainer>
      </Box>
    </Box>
  );
}

export default function OperationalAnalyticsDashboard() {
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
      label: 'Dropship Purchase Analysis',
      component: <DropshipPurchaseAnalysis />,
    },
    {
      label: 'Reconciliation Dashboard',
      component: <ReconciliationDashboard />,
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
        title="Operational Analytics Dashboard"
        tabs={tabs}
      />
    </Box>
  );
}