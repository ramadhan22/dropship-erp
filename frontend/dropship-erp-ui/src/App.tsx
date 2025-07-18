import { BrowserRouter, Route, Routes } from "react-router-dom";
import { Suspense, lazy } from "react";
import { CircularProgress, Box, AppBar, Toolbar, Typography, Container } from "@mui/material";
import BreadcrumbsNav from "./components/BreadcrumbsNav";
import { ToastProvider } from "./components/ToastProvider";
import { NavigationDropdown, navigationSections } from "./components/NavigationDropdown";
import { colors, spacing } from "./theme/tokens";

// Immediate load for homepage and key navigation
import SalesSummaryPage from "./components/SalesSummaryPage";

// Lazy load all other pages for code splitting
const BalanceSheetPage = lazy(() => import("./components/BalanceSheetPage"));
const ChannelPage = lazy(() => import("./components/ChannelPage"));
const DropshipImport = lazy(() => import("./components/DropshipImport"));
const MetricsPage = lazy(() => import("./components/MetricsPage"));
const ReconcileForm = lazy(() => import("./components/ReconcileForm"));
const ReconcileDashboard = lazy(() => import("./components/ReconcileDashboard"));
const ShopeeSalesPage = lazy(() => import("./components/ShopeeSalesPage"));
const ShopeeAffiliatePage = lazy(() => import("./components/ShopeeAffiliatePage"));
const AccountPage = lazy(() => import("./components/AccountPage"));
const ExpensePage = lazy(() => import("./components/ExpensePage"));
const JournalPage = lazy(() => import("./components/JournalPage"));
const AdInvoicePage = lazy(() => import("./components/AdInvoicePage"));
const PLPage = lazy(() => import("./components/PLPage"));
const GLPage = lazy(() => import("./components/GLPage"));
const SalesProfitPage = lazy(() => import("./components/SalesProfitPage"));
const Dashboard = lazy(() => import("./components/Dashboard"));
const KasAccountPage = lazy(() => import("./components/KasAccountPage"));
const PendingBalancePage = lazy(() => import("./components/PendingBalancePage"));
const StoreDetailPage = lazy(() => import("./components/StoreDetailPage"));
const WithdrawalPage = lazy(() => import("./components/WithdrawalPage"));
const ShopeeAdjustmentPage = lazy(() => import("./components/ShopeeAdjustmentPage"));
const ShippingDiscrepancyPage = lazy(() => import("./components/ShippingDiscrepancyPage"));
const TaxPaymentPage = lazy(() => import("./components/TaxPaymentPage"));
const ShopeeOrderDetailPage = lazy(() => import("./components/ShopeeOrderDetailPage"));
const WalletTransactionPage = lazy(() => import("./components/WalletTransactionPage"));
const AdsTopupPage = lazy(() => import("./components/AdsTopupPage"));
const BatchHistoryPage = lazy(() => import("./components/BatchHistoryPage"));
const AdsPerformancePage = lazy(() => import("./components/AdsPerformancePage"));

// Loading fallback component
const LoadingFallback = () => (
  <Box display="flex" justifyContent="center" alignItems="center" minHeight="200px">
    <CircularProgress />
  </Box>
);

export default function App() {
  return (
    <ToastProvider>
      <BrowserRouter>
        <AppBar position="static" sx={{ backgroundColor: colors.primary }}>
          <Toolbar>
            <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
              Dropship ERP
            </Typography>
            <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
              {navigationSections.map((section) => (
                <NavigationDropdown
                  key={section.label}
                  label={section.label}
                  icon={section.icon}
                  items={section.items}
                />
              ))}
            </Box>
          </Toolbar>
        </AppBar>
        
        <Container maxWidth="xl" sx={{ py: spacing.md }}>
          <BreadcrumbsNav />
          <Suspense fallback={<LoadingFallback />}>
            <Routes>
              <Route path="/" element={<SalesSummaryPage />} />
              <Route path="/dashboard" element={<Dashboard />} />
              <Route path="/dropship" element={<DropshipImport />} />
              <Route path="/shopee" element={<ShopeeSalesPage />} />
              <Route path="/sales-profit" element={<SalesProfitPage />} />
              <Route
                path="/shopee/adjustments"
                element={<ShopeeAdjustmentPage />}
              />
              <Route
                path="/shipping-discrepancies"
                element={<ShippingDiscrepancyPage />}
              />
              <Route path="/shopee/affiliate" element={<ShopeeAffiliatePage />} />
              <Route path="/reconcile" element={<ReconcileForm />} />
              <Route path="/reconcile/dashboard" element={<ReconcileDashboard />} />
              <Route path="/metrics" element={<MetricsPage />} />
              <Route path="/pl" element={<PLPage />} />
              <Route path="/balance" element={<BalanceSheetPage />} />
              <Route path="/gl" element={<GLPage />} />
              <Route path="/channels" element={<ChannelPage />} />
              <Route path="/accounts" element={<AccountPage />} />
              <Route path="/expenses" element={<ExpensePage />} />
              <Route path="/ads" element={<AdInvoicePage />} />
              <Route path="/journal" element={<JournalPage />} />
              <Route path="/kas" element={<KasAccountPage />} />
              <Route path="/stores/:id" element={<StoreDetailPage />} />
              <Route path="/pending-balance" element={<PendingBalancePage />} />
              <Route path="/tax-payment" element={<TaxPaymentPage />} />
              <Route path="/order-details" element={<ShopeeOrderDetailPage />} />
              <Route path="/withdrawals" element={<WithdrawalPage />} />
              <Route path="/ads-topups" element={<AdsTopupPage />} />
              <Route path="/ads-performance" element={<AdsPerformancePage />} />
              <Route path="/batches" element={<BatchHistoryPage />} />
              <Route
                path="/wallet-transactions"
                element={<WalletTransactionPage />}
              />
            </Routes>
          </Suspense>
        </Container>
      </BrowserRouter>
    </ToastProvider>
  );
}
