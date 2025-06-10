import { BrowserRouter, Link, Route, Routes } from "react-router-dom";
import BalanceSheetPage from "./components/BalanceSheetPage";
import DropshipImport from "./components/DropshipImport";
import MetricsPage from "./components/MetricsPage";
import ReconcileForm from "./components/ReconcileForm";
import ShopeeImport from "./components/ShopeeImport";

export default function App() {
  return (
    <BrowserRouter>
      <nav style={{ padding: "1rem", borderBottom: "1px solid #ccc" }}>
        <Link to="/dropship">Dropship Import</Link> |{" "}
        <Link to="/shopee">Shopee Import</Link> |{" "}
        <Link to="/reconcile">Reconcile</Link> |{" "}
        <Link to="/metrics">Metrics</Link> |{" "}
        <Link to="/balance">Balance Sheet</Link>
      </nav>
      <div style={{ padding: "1rem" }}>
        <Routes>
          <Route path="/dropship" element={<DropshipImport />} />
          <Route path="/shopee" element={<ShopeeImport />} />
          <Route path="/reconcile" element={<ReconcileForm />} />
          <Route path="/metrics" element={<MetricsPage />} />
          <Route path="/balance" element={<BalanceSheetPage />} />
        </Routes>
      </div>
    </BrowserRouter>
  );
}
