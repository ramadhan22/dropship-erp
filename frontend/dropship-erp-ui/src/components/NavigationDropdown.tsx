import { useState, useCallback } from 'react';
import { 
  Button, 
  Menu, 
  MenuItem, 
  ListItemIcon, 
  ListItemText
} from '@mui/material';
import { Link } from 'react-router-dom';
import { 
  KeyboardArrowDown,
  Home as HomeIcon,
  Dashboard as DashboardIcon,
  CloudUpload as ImportIcon,
  Store as ShopeeIcon,
  AccountBalance as FinancialIcon,
  Business as ManagementIcon,
  Campaign as AdsIcon,
  Payment as TransactionIcon,
  Business as BusinessIcon,
  Payment as PaymentIcon,
  Analytics as AnalyticsIcon,
  TrendingUp as TrendingUpIcon,
  Assessment as AssessmentIcon,
  Insights as InsightsIcon
} from '@mui/icons-material';

interface NavigationItem {
  label: string;
  path: string;
  icon?: React.ReactNode;
}

interface NavigationDropdownProps {
  label: string;
  icon?: React.ReactNode;
  items: NavigationItem[];
}

export const NavigationDropdown = ({ label, icon, items }: NavigationDropdownProps) => {
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const open = Boolean(anchorEl);

  const handleClick = useCallback((event: React.MouseEvent<HTMLButtonElement>) => {
    setAnchorEl(event.currentTarget);
  }, []);

  const handleClose = useCallback(() => {
    setAnchorEl(null);
  }, []);

  return (
    <>
      <Button
        color="inherit"
        onClick={handleClick}
        endIcon={<KeyboardArrowDown />}
        startIcon={icon}
        sx={{ textTransform: 'none' }}
      >
        {label}
      </Button>
      <Menu
        anchorEl={anchorEl}
        open={open}
        onClose={handleClose}
        MenuListProps={{
          'aria-labelledby': 'basic-button',
        }}
        anchorOrigin={{
          vertical: 'bottom',
          horizontal: 'left',
        }}
        transformOrigin={{
          vertical: 'top',
          horizontal: 'left',
        }}
      >
        {items.map((item) => (
          <MenuItem
            key={item.path}
            component={Link}
            to={item.path}
            onClick={handleClose}
            sx={{ 
              minWidth: 180,
              '&:hover': {
                backgroundColor: 'rgba(0, 0, 0, 0.04)',
              }
            }}
          >
            {item.icon && (
              <ListItemIcon sx={{ minWidth: '36px !important' }}>
                {item.icon}
              </ListItemIcon>
            )}
            <ListItemText>{item.label}</ListItemText>
          </MenuItem>
        ))}
      </Menu>
    </>
  );
};

// Navigation sections configuration
export const navigationSections = [
  {
    label: 'Home',
    icon: <HomeIcon fontSize="small" />,
    items: [
      { label: 'Home', path: '/', icon: <HomeIcon fontSize="small" /> },
      { label: 'Dashboard', path: '/dashboard', icon: <DashboardIcon fontSize="small" /> },
    ]
  },
  {
    label: 'Performance Dashboard',
    icon: <AnalyticsIcon fontSize="small" />,
    items: [
      { label: 'Financial Performance', path: '/performance-dashboard/financial', icon: <TrendingUpIcon fontSize="small" /> },
      { label: 'Operational Analytics', path: '/performance-dashboard/operational', icon: <AssessmentIcon fontSize="small" /> },
      { label: 'Marketplace Integration', path: '/performance-dashboard/marketplace', icon: <ShopeeIcon fontSize="small" /> },
      { label: 'Financial Management', path: '/performance-dashboard/management', icon: <InsightsIcon fontSize="small" /> },
    ]
  },
  {
    label: 'Operations',
    icon: <ImportIcon fontSize="small" />,
    items: [
      { label: 'Dropship Import', path: '/dropship', icon: <ImportIcon fontSize="small" /> },
      { label: 'Batch History', path: '/batches', icon: <DashboardIcon fontSize="small" /> },
    ]
  },
  {
    label: 'Shopee',
    icon: <ShopeeIcon fontSize="small" />,
    items: [
      { label: 'Shopee Sales', path: '/shopee', icon: <ShopeeIcon fontSize="small" /> },
      { label: 'Sales Profit', path: '/sales-profit', icon: <FinancialIcon fontSize="small" /> },
      { label: 'Adjustments', path: '/shopee/adjustments', icon: <ManagementIcon fontSize="small" /> },
      { label: 'Shipping Discrepancies', path: '/shipping-discrepancies', icon: <PaymentIcon fontSize="small" /> },
      { label: 'Affiliate Sales', path: '/shopee/affiliate', icon: <BusinessIcon fontSize="small" /> },
      { label: 'Order Details', path: '/order-details', icon: <DashboardIcon fontSize="small" /> },
      { label: 'Order Returns', path: '/order-returns', icon: <PaymentIcon fontSize="small" /> },
    ]
  },
  {
    label: 'Financial',
    icon: <FinancialIcon fontSize="small" />,
    items: [
      { label: 'Reconcile', path: '/reconcile', icon: <FinancialIcon fontSize="small" /> },
      { label: 'Reconcile Dashboard', path: '/reconcile/dashboard', icon: <DashboardIcon fontSize="small" /> },
      { label: 'P&L Report', path: '/pl', icon: <FinancialIcon fontSize="small" /> },
      { label: 'P&L Sankey Diagram', path: '/pl-sankey', icon: <DashboardIcon fontSize="small" /> },
      { label: 'P&L Sankey Demo', path: '/pl-sankey-demo', icon: <DashboardIcon fontSize="small" /> },
      { label: 'Balance Sheet', path: '/balance', icon: <FinancialIcon fontSize="small" /> },
      { label: 'General Ledger', path: '/gl', icon: <BusinessIcon fontSize="small" /> },
      { label: 'Metrics', path: '/metrics', icon: <DashboardIcon fontSize="small" /> },
    ]
  },
  {
    label: 'Management',
    icon: <ManagementIcon fontSize="small" />,
    items: [
      { label: 'Accounts', path: '/accounts', icon: <FinancialIcon fontSize="small" /> },
      { label: 'Channels', path: '/channels', icon: <BusinessIcon fontSize="small" /> },
      { label: 'Expenses', path: '/expenses', icon: <PaymentIcon fontSize="small" /> },
      { label: 'Journal', path: '/journal', icon: <BusinessIcon fontSize="small" /> },
      { label: 'Kas Accounts', path: '/kas', icon: <FinancialIcon fontSize="small" /> },
      { label: 'Pending Balance', path: '/pending-balance', icon: <FinancialIcon fontSize="small" /> },
    ]
  },
  {
    label: 'Ads',
    icon: <AdsIcon fontSize="small" />,
    items: [
      { label: 'Ads Invoice', path: '/ads', icon: <PaymentIcon fontSize="small" /> },
      { label: 'Ads Topup', path: '/ads-topups', icon: <AdsIcon fontSize="small" /> },
      { label: 'Ads Performance', path: '/ads-performance', icon: <DashboardIcon fontSize="small" /> },
    ]
  },
  {
    label: 'Transactions',
    icon: <TransactionIcon fontSize="small" />,
    items: [
      { label: 'Withdrawals', path: '/withdrawals', icon: <PaymentIcon fontSize="small" /> },
      { label: 'Tax Payment', path: '/tax-payment', icon: <PaymentIcon fontSize="small" /> },
      { label: 'Wallet Transactions', path: '/wallet-transactions', icon: <PaymentIcon fontSize="small" /> },
    ]
  },
];