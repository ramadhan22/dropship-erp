import { BrowserRouter, Link, Route, Routes } from "react-router-dom";
import BalanceSheetPage from "./components/BalanceSheetPage";
import ChannelPage from "./components/ChannelPage";
import DropshipImport from "./components/DropshipImport";
import MetricsPage from "./components/MetricsPage";
import ReconcileForm from "./components/ReconcileForm";
import ShopeeSalesPage from "./components/ShopeeSalesPage";

export default function App() {
  return (
    <BrowserRouter>
      <nav style={{ padding: "1rem", borderBottom: "1px solid #ccc" }}>
        <Link to="/dropship">Dropship Import</Link> |{" "}
        <Link to="/shopee">Shopee Sales</Link> |{" "}
        <Link to="/reconcile">Reconcile</Link> |{" "}
        <Link to="/metrics">Metrics</Link> |{" "}
        <Link to="/balance">Balance Sheet</Link> |{" "}
        <Link to="/channels">Channels</Link>
      </nav>
      <div style={{ padding: "1rem" }}>
        <Routes>
          <Route path="/dropship" element={<DropshipImport />} />
          <Route path="/shopee" element={<ShopeeSalesPage />} />
          <Route path="/reconcile" element={<ReconcileForm />} />
          <Route path="/metrics" element={<MetricsPage />} />
          <Route path="/balance" element={<BalanceSheetPage />} />
          <Route path="/channels" element={<ChannelPage />} />
        </Routes>
      </div>
    </BrowserRouter>
  );
}
