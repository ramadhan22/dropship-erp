import { BrowserRouter, Link, Route, Routes } from "react-router-dom";
import BreadcrumbsNav from "./components/BreadcrumbsNav";
import BalanceSheetPage from "./components/BalanceSheetPage";
import ChannelPage from "./components/ChannelPage";
import DropshipImport from "./components/DropshipImport";
import MetricsPage from "./components/MetricsPage";
import ReconcileForm from "./components/ReconcileForm";
import ReconcileDashboard from "./components/ReconcileDashboard";
import ShopeeSalesPage from "./components/ShopeeSalesPage";
import ShopeeAffiliatePage from "./components/ShopeeAffiliatePage";
import AccountPage from "./components/AccountPage";
import ExpensePage from "./components/ExpensePage";
import JournalPage from "./components/JournalPage";
import AdInvoicePage from "./components/AdInvoicePage";
import PLPage from "./components/PLPage";
import GLPage from "./components/GLPage";
import SalesSummaryPage from "./components/SalesSummaryPage";
import SalesProfitPage from "./components/SalesProfitPage";
import KasAccountPage from "./components/KasAccountPage";
import PendingBalancePage from "./components/PendingBalancePage";
import StoreDetailPage from "./components/StoreDetailPage";
import WithdrawalPage from "./components/WithdrawalPage";
import ShopeeAdjustmentPage from "./components/ShopeeAdjustmentPage";
import TaxPaymentPage from "./components/TaxPaymentPage";
import ShopeeOrderDetailPage from "./components/ShopeeOrderDetailPage";
import WalletTransactionPage from "./components/WalletTransactionPage";
import AdsTopupPage from "./components/AdsTopupPage";

export default function App() {
  return (
    <BrowserRouter>
      <nav style={{ padding: "1rem", borderBottom: "1px solid #ccc" }}>
        <Link to="/">Home</Link> | <Link to="/dropship">Dropship Import</Link> |{" "}
        <Link to="/shopee">Shopee Sales</Link> |{" "}
        <Link to="/sales-profit">Sales Profit</Link> |{" "}
        <Link to="/shopee/adjustments">Adjustments</Link> |{" "}
        <Link to="/shopee/affiliate">Affiliate Sales</Link> |{" "}
        <Link to="/reconcile">Reconcile</Link> |{" "}
        <Link to="/metrics">Metrics</Link> | <Link to="/pl">P&L</Link> |{" "}
        <Link to="/balance">Balance Sheet</Link> | <Link to="/gl">GL</Link> |{" "}
        <Link to="/channels">Channels</Link> |{" "}
        <Link to="/accounts">Accounts</Link> |{" "}
        <Link to="/expenses">Expenses</Link> |{" "}
        <Link to="/ads">Ads Invoice</Link> | <Link to="/journal">Journal</Link>{" "}
        | <Link to="/kas">Kas</Link> |{" "}
        <Link to="/pending-balance">Pending Balance</Link> |{" "}
        <Link to="/order-details">Order Details</Link> |{" "}
        <Link to="/reconcile/dashboard">Reconcile Dashboard</Link> |{" "}
        <Link to="/tax-payment">Tax Payment</Link> |{" "}
        <Link to="/withdrawals">Withdrawals</Link> |{" "}
        <Link to="/ads-topups">Ads Topup</Link> |{" "}
        <Link to="/wallet-transactions">Wallet Txn</Link>
      </nav>
      <div style={{ padding: "1rem" }}>
        <BreadcrumbsNav />
        <Routes>
          <Route path="/" element={<SalesSummaryPage />} />
          <Route path="/dropship" element={<DropshipImport />} />
          <Route path="/shopee" element={<ShopeeSalesPage />} />
          <Route path="/sales-profit" element={<SalesProfitPage />} />
          <Route
            path="/shopee/adjustments"
            element={<ShopeeAdjustmentPage />}
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
          <Route
            path="/wallet-transactions"
            element={<WalletTransactionPage />}
          />
        </Routes>
      </div>
    </BrowserRouter>
  );
}
