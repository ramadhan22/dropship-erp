import { useState } from 'react';
import { Box, Typography, Card, CardContent } from '@mui/material';
import PerformanceDashboardLayout from './PerformanceDashboardLayout';
import DashboardFilterContainer, { type DashboardFilters } from './DashboardFilterContainer';
import ChartContainer from './ChartContainer';
import { colors, spacing } from '../theme/tokens';

// Account & Journal Analysis Component
function AccountAndJournalAnalysis() {
  return (
    <Box>
      {/* Account Metrics Cards */}
      <Box sx={{ 
        display: 'grid', 
        gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr', md: 'repeat(4, 1fr)' },
        gap: spacing.md,
        mb: spacing.lg 
      }}>
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" sx={{ color: colors.primary, mb: spacing.sm }}>
              Total Accounts
            </Typography>
            <Typography variant="h4" sx={{ color: colors.text, mb: spacing.xs }}>
              247
            </Typography>
            <Typography variant="body2" sx={{ color: colors.info }}>
              Active accounts
            </Typography>
          </CardContent>
        </Card>
        
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" sx={{ color: colors.primary, mb: spacing.sm }}>
              Journal Entries
            </Typography>
            <Typography variant="h4" sx={{ color: colors.text, mb: spacing.xs }}>
              18,946
            </Typography>
            <Typography variant="body2" sx={{ color: colors.success }}>
              This month
            </Typography>
          </CardContent>
        </Card>
        
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" sx={{ color: colors.primary, mb: spacing.sm }}>
              Unreconciled Items
            </Typography>
            <Typography variant="h4" sx={{ color: colors.text, mb: spacing.xs }}>
              89
            </Typography>
            <Typography variant="body2" sx={{ color: colors.warning }}>
              Needs attention
            </Typography>
          </CardContent>
        </Card>
        
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" sx={{ color: colors.primary, mb: spacing.sm }}>
              Balance Accuracy
            </Typography>
            <Typography variant="h4" sx={{ color: colors.text, mb: spacing.xs }}>
              99.7%
            </Typography>
            <Typography variant="body2" sx={{ color: colors.success }}>
              High accuracy rate
            </Typography>
          </CardContent>
        </Card>
      </Box>

      {/* Charts */}
      <Box sx={{ mt: spacing.lg }}>
        <ChartContainer
          title="Account Balance Trends"
          subtitle="Track balance changes across major account categories"
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
              Account balance trends chart will be displayed here
            </Typography>
          </Box>
        </ChartContainer>

        <ChartContainer
          title="Journal Entry Analysis"
          subtitle="Daily journal entry volume and transaction types"
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
              Journal entry analysis chart will be displayed here
            </Typography>
          </Box>
        </ChartContainer>
      </Box>
    </Box>
  );
}

// Tax & Compliance Tracking Component
function TaxAndComplianceTracking() {
  return (
    <Box>
      {/* Tax Metrics Cards */}
      <Box sx={{ 
        display: 'grid', 
        gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr', md: 'repeat(3, 1fr)' },
        gap: spacing.md,
        mb: spacing.lg 
      }}>
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" sx={{ color: colors.primary, mb: spacing.sm }}>
              VAT Collected
            </Typography>
            <Typography variant="h4" sx={{ color: colors.text, mb: spacing.xs }}>
              ₹24,56,789
            </Typography>
            <Typography variant="body2" sx={{ color: colors.success }}>
              Current period
            </Typography>
          </CardContent>
        </Card>
        
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" sx={{ color: colors.primary, mb: spacing.sm }}>
              Income Tax Due
            </Typography>
            <Typography variant="h4" sx={{ color: colors.text, mb: spacing.xs }}>
              ₹12,34,567
            </Typography>
            <Typography variant="body2" sx={{ color: colors.warning }}>
              Due in 15 days
            </Typography>
          </CardContent>
        </Card>
        
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" sx={{ color: colors.primary, mb: spacing.sm }}>
              Compliance Score
            </Typography>
            <Typography variant="h4" sx={{ color: colors.text, mb: spacing.xs }}>
              98.5%
            </Typography>
            <Typography variant="body2" sx={{ color: colors.success }}>
              Excellent compliance
            </Typography>
          </CardContent>
        </Card>
      </Box>

      <Box sx={{ mt: spacing.lg }}>
        <ChartContainer
          title="Tax Liability Tracking"
          subtitle="Monthly tax obligations and payment schedules"
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
              Tax liability tracking dashboard will be displayed here
            </Typography>
          </Box>
        </ChartContainer>

        <ChartContainer
          title="Compliance Status Overview"
          subtitle="Track compliance status across different regulatory requirements"
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
              Compliance status overview will be displayed here
            </Typography>
          </Box>
        </ChartContainer>
      </Box>
    </Box>
  );
}

export default function FinancialManagementDashboard() {
  const [filters, setFilters] = useState<DashboardFilters>({
    dateFrom: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000), // 30 days ago
    dateTo: new Date(),
    stores: [],
    channels: [],
    currency: 'IDR',
    period: 'monthly', // Default to monthly for financial management
  });

  const tabs = [
    {
      label: 'Account & Journal Analysis',
      component: <AccountAndJournalAnalysis />,
    },
    {
      label: 'Tax & Compliance Tracking',
      component: <TaxAndComplianceTracking />,
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
        title="Financial Management Dashboard"
        tabs={tabs}
      />
    </Box>
  );
}