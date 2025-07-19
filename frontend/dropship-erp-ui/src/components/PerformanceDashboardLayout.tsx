import { Box, Paper, Typography, Tabs, Tab } from '@mui/material';
import { useState } from 'react';
import { colors, spacing } from '../theme/tokens';

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`dashboard-tabpanel-${index}`}
      aria-labelledby={`dashboard-tab-${index}`}
      {...other}
    >
      {value === index && (
        <Box sx={{ pt: spacing.lg }}>
          {children}
        </Box>
      )}
    </div>
  );
}

function a11yProps(index: number) {
  return {
    id: `dashboard-tab-${index}`,
    'aria-controls': `dashboard-tabpanel-${index}`,
  };
}

interface PerformanceDashboardLayoutProps {
  title: string;
  tabs: Array<{
    label: string;
    component: React.ReactNode;
  }>;
}

export default function PerformanceDashboardLayout({ title, tabs }: PerformanceDashboardLayoutProps) {
  const [tabValue, setTabValue] = useState(0);

  const handleTabChange = (_event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
  };

  return (
    <Box sx={{ width: '100%' }}>
      {/* Header */}
      <Box sx={{ mb: spacing.lg }}>
        <Typography variant="h4" component="h1" sx={{ mb: spacing.sm, color: colors.text }}>
          {title}
        </Typography>
        <Typography variant="body1" sx={{ color: colors.textSecondary }}>
          Comprehensive analytics and insights for business performance monitoring
        </Typography>
      </Box>

      {/* Main Content */}
      <Paper elevation={1} sx={{ borderRadius: '8px', overflow: 'hidden' }}>
        {/* Tab Navigation */}
        <Box sx={{ borderBottom: 1, borderColor: colors.divider }}>
          <Tabs
            value={tabValue}
            onChange={handleTabChange}
            aria-label="dashboard tabs"
            sx={{
              px: spacing.lg,
              '& .MuiTab-root': {
                textTransform: 'none',
                fontSize: '1rem',
                fontWeight: 500,
                minHeight: 64,
              },
            }}
          >
            {tabs.map((tab, index) => (
              <Tab key={index} label={tab.label} {...a11yProps(index)} />
            ))}
          </Tabs>
        </Box>

        {/* Tab Content */}
        <Box sx={{ p: spacing.lg }}>
          {tabs.map((tab, index) => (
            <TabPanel key={index} value={tabValue} index={index}>
              {tab.component}
            </TabPanel>
          ))}
        </Box>
      </Paper>
    </Box>
  );
}