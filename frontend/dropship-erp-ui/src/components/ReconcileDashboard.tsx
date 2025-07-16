import React, { useState, useEffect, useCallback } from "react";
import SortableTable from "./SortableTable";
import { fetchAPI } from "../api";
import type { ReconcileCandidate, ReconciledTransaction } from "../types";

export default function ReconcileDashboard() {
  const [candidates, setCandidates] = useState<ReconcileCandidate[]>([]);
  const [transactions, setTransactions] = useState<ReconciledTransaction[]>([]);
  const [loading, setLoading] = useState(false);
  const [reconcileLoading, setReconcileLoading] = useState(false);
  const [shop, setShop] = useState("");
  const [fromDate, setFromDate] = useState("");
  const [toDate, setToDate] = useState("");
  const [message, setMessage] = useState("");

  const fetchCandidates = useCallback(async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams();
      if (shop) params.append("shop", shop);
      if (fromDate) params.append("from", fromDate);
      if (toDate) params.append("to", toDate);
      params.append("limit", "100");
      params.append("offset", "0");

      const response = await fetchAPI(`/reconcile/candidates?${params}`);
      setCandidates(response.data || []);
    } catch (err) {
      console.error("Error fetching candidates:", err);
      setMessage("Error fetching reconcile candidates");
    } finally {
      setLoading(false);
    }
  }, [shop, fromDate, toDate]);

  const fetchTransactions = useCallback(async () => {
    if (!shop || !fromDate) return;
    
    try {
      const period = fromDate.substring(0, 7); // YYYY-MM format
      const response = await fetchAPI(`/reconcile/transactions?shop=${shop}&period=${period}`);
      setTransactions(response.data || []);
    } catch (err) {
      console.error("Error fetching transactions:", err);
    }
  }, [shop, fromDate]);

  useEffect(() => {
    fetchCandidates();
    fetchTransactions();
  }, [fetchCandidates, fetchTransactions]);

  const handleReconcileAll = async () => {
    if (!shop || !fromDate || !toDate) {
      setMessage("Please select shop and date range");
      return;
    }

    setReconcileLoading(true);
    setMessage("");

    try {
      const requestBody = {
        shop,
        from_date: fromDate,
        to_date: toDate,
      };

      const response = await fetchAPI("/reconcile/all", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(requestBody),
      });

      setMessage(response.message || "Reconcile all completed successfully");
      
      // Refresh the data
      fetchCandidates();
      fetchTransactions();
    } catch (err) {
      console.error("Error during reconcile all:", err);
      setMessage("Error during reconcile all operation");
    } finally {
      setReconcileLoading(false);
    }
  };

  const updateEscrowStatus = async (orderSN: string, status: string) => {
    try {
      await fetchAPI(`/reconcile/escrow/${orderSN}/status`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ status }),
      });
      
      setMessage(`Escrow status updated for order ${orderSN}`);
      fetchTransactions();
    } catch (err) {
      console.error("Error updating escrow status:", err);
      setMessage("Error updating escrow status");
    }
  };

  const candidateColumns = [
    { key: "kode_pesanan", label: "Kode Pesanan" },
    { key: "kode_invoice_channel", label: "Invoice Channel" },
    { key: "nama_toko", label: "Store" },
    { key: "status_pesanan_terakhir", label: "Status" },
    { key: "no_pesanan", label: "Shopee Order" },
    { key: "shopee_order_status", label: "Shopee Status" },
  ];

  const transactionColumns = [
    { key: "id", label: "ID" },
    { key: "shop_username", label: "Shop" },
    { key: "dropship_id", label: "Dropship ID" },
    { key: "shopee_id", label: "Shopee ID" },
    { key: "status", label: "Status" },
    { key: "matched_at", label: "Matched At" },
    {
      key: "actions",
      label: "Actions",
      render: (transaction: ReconciledTransaction) => (
        <div style={{ display: "flex", gap: "0.5rem" }}>
          <button
            onClick={() => updateEscrowStatus(transaction.shopee_id || "", "processed")}
            style={{
              padding: "0.25rem 0.5rem",
              border: "1px solid #ccc",
              borderRadius: "4px",
              backgroundColor: "#f8f9fa",
              cursor: "pointer",
            }}
          >
            Mark Processed
          </button>
        </div>
      ),
    },
  ];

  return (
    <div style={{ padding: "1rem" }}>
      <h1>Reconcile Dashboard</h1>

      {/* Filter Controls */}
      <div style={{ marginBottom: "1rem", display: "flex", gap: "1rem", alignItems: "center" }}>
        <div>
          <label style={{ display: "block", marginBottom: "0.25rem" }}>Shop:</label>
          <input
            type="text"
            value={shop}
            onChange={(e) => setShop(e.target.value)}
            placeholder="Enter shop name"
            style={{ padding: "0.5rem", border: "1px solid #ccc", borderRadius: "4px" }}
          />
        </div>
        <div>
          <label style={{ display: "block", marginBottom: "0.25rem" }}>From Date:</label>
          <input
            type="date"
            value={fromDate}
            onChange={(e) => setFromDate(e.target.value)}
            style={{ padding: "0.5rem", border: "1px solid #ccc", borderRadius: "4px" }}
          />
        </div>
        <div>
          <label style={{ display: "block", marginBottom: "0.25rem" }}>To Date:</label>
          <input
            type="date"
            value={toDate}
            onChange={(e) => setToDate(e.target.value)}
            style={{ padding: "0.5rem", border: "1px solid #ccc", borderRadius: "4px" }}
          />
        </div>
        <div style={{ marginTop: "1.5rem" }}>
          <button
            onClick={handleReconcileAll}
            disabled={reconcileLoading || !shop || !fromDate || !toDate}
            style={{
              padding: "0.5rem 1rem",
              backgroundColor: reconcileLoading ? "#ccc" : "#007bff",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: reconcileLoading ? "not-allowed" : "pointer",
              fontWeight: "bold",
            }}
          >
            {reconcileLoading ? "Processing..." : "Reconcile All"}
          </button>
        </div>
      </div>

      {/* Status Message */}
      {message && (
        <div
          style={{
            padding: "0.75rem",
            marginBottom: "1rem",
            backgroundColor: message.includes("Error") ? "#f8d7da" : "#d4edda",
            border: `1px solid ${message.includes("Error") ? "#f5c6cb" : "#c3e6cb"}`,
            borderRadius: "4px",
            color: message.includes("Error") ? "#721c24" : "#155724",
          }}
        >
          {message}
        </div>
      )}

      {/* Reconcile Candidates Section */}
      <div style={{ marginBottom: "2rem" }}>
        <h2>Reconcile Candidates</h2>
        <p style={{ color: "#666", marginBottom: "1rem" }}>
          Orders that need reconciliation or may have returned order escrow settlements.
        </p>
        {loading ? (
          <div>Loading candidates...</div>
        ) : (
          <SortableTable
            data={candidates}
            columns={candidateColumns}
            defaultSort="kode_pesanan"
          />
        )}
      </div>

      {/* Reconciled Transactions Section */}
      <div>
        <h2>Reconciled Transactions</h2>
        <p style={{ color: "#666", marginBottom: "1rem" }}>
          Previously reconciled transactions with escrow status management.
        </p>
        <SortableTable
          data={transactions}
          columns={transactionColumns}
          defaultSort="matched_at"
        />
      </div>
    </div>
  );
}