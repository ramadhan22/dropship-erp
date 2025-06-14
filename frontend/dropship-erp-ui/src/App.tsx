import { BrowserRouter, Link, Route, Routes } from "react-router-dom";
import BalanceSheetPage from "./components/BalanceSheetPage";
import ChannelPage from "./components/ChannelPage";
import DropshipImport from "./components/DropshipImport";
import MetricsPage from "./components/MetricsPage";
import ReconcileForm from "./components/ReconcileForm";
import ReconcileDashboard from "./components/ReconcileDashboard";
import ShopeeSalesPage from "./components/ShopeeSalesPage";
import ShopeeAffiliateImport from "./components/ShopeeAffiliateImport";
import ShopeeImport from "./components/ShopeeImport";
import AccountPage from "./components/AccountPage";
import ExpensePage from "./components/ExpensePage";
import JournalPage from "./components/JournalPage";
import PLPage from "./components/PLPage";
import GLPage from "./components/GLPage";
import SalesSummaryPage from "./components/SalesSummaryPage";

export default function App() {
  return (
    <BrowserRouter>
      <nav style={{ padding: "1rem", borderBottom: "1px solid #ccc" }}>
        <Link to="/">Home</Link> | <Link to="/dropship">Dropship Import</Link> |{" "}
        <Link to="/shopee">Shopee Sales</Link> |{" "}
        <Link to="/shopee/affiliate">Affiliate Import</Link> |{" "}
        <Link to="/reconcile">Reconcile</Link> |{" "}
        <Link to="/metrics">Metrics</Link> | <Link to="/pl">P&L</Link> |{" "}
        <Link to="/balance">Balance Sheet</Link> | <Link to="/gl">GL</Link> |{" "}
        <Link to="/channels">Channels</Link> |{" "}
        <Link to="/accounts">Accounts</Link> |{" "}
        <Link to="/expenses">Expenses</Link> |{" "}
        <Link to="/journal">Journal</Link> |{" "}
        <Link to="/reconcile/dashboard">Reconcile Dashboard</Link>
      </nav>
      <div style={{ padding: "1rem" }}>
        <Routes>
          <Route path="/" element={<SalesSummaryPage />} />
          <Route path="/dropship" element={<DropshipImport />} />
          <Route path="/shopee" element={<ShopeeSalesPage />} />
          <Route path="/shopee/affiliate" element={<ShopeeAffiliateImport />} />
          <Route path="/reconcile" element={<ReconcileForm />} />
          <Route path="/reconcile/dashboard" element={<ReconcileDashboard />} />
          <Route path="/metrics" element={<MetricsPage />} />
          <Route path="/pl" element={<PLPage />} />
          <Route path="/balance" element={<BalanceSheetPage />} />
          <Route path="/gl" element={<GLPage />} />
          <Route path="/channels" element={<ChannelPage />} />
          <Route path="/accounts" element={<AccountPage />} />
          <Route path="/expenses" element={<ExpensePage />} />
          <Route path="/journal" element={<JournalPage />} />
        </Routes>
      </div>
    </BrowserRouter>
  );
}
