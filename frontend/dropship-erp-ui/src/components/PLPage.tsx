import { useEffect, useState } from "react";
import {
  Box,
  Typography,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  CircularProgress,
  Alert,
} from "@mui/material";
import WarningIcon from "@mui/icons-material/Warning";
import { fetchProfitLoss } from "../api/pl";
import { listAllStores } from "../api";
import type { Store } from "../types";

interface PLRow {
  label: string;
  amount: number;
  percent?: number;
  previousAmount?: number;
  change?: number;
  changePercent?: number;
  manual?: boolean;
  indent?: number;
  group?: boolean;
}

interface ProfitLoss {
  pendapatanUsaha: PLRow[];
  totalPendapatanUsaha: number;
  prevTotalPendapatanUsaha?: number;
  hargaPokokPenjualan: PLRow[];
  totalHargaPokokPenjualan: number;
  prevTotalHargaPokokPenjualan?: number;
  labaKotor: { amount: number; percent?: number; previousAmount?: number; change?: number; changePercent?: number };
  bebanOperasional: PLRow[];
  totalBebanOperasional: number;
  prevTotalBebanOperasional?: number;
  bebanPemasaran: PLRow[];
  totalBebanPemasaran: number;
  prevTotalBebanPemasaran?: number;
  bebanAdministrasi: PLRow[];
  totalBebanAdministrasi: number;
  prevTotalBebanAdministrasi?: number;
  totalBebanUsaha: { amount: number; percent?: number; previousAmount?: number; change?: number; changePercent?: number };
  labaSebelumPajak: number;
  prevLabaSebelumPajak?: number;
  pajakPenghasilan: PLRow[];
  totalPajakPenghasilan: number;
  prevTotalPajakPenghasilan?: number;
  labaRugiBersih: { amount: number; percent?: number; previousAmount?: number; change?: number; changePercent?: number };
}

const months = [
  "Jan",
  "Feb",
  "Mar",
  "Apr",
  "May",
  "Jun",
  "Jul",
  "Aug",
  "Sep",
  "Oct",
  "Nov",
  "Dec",
];

const years = [2023, 2024, 2025];

function renderRows(
  rows: PLRow[] | null | undefined,
  fmt: Intl.NumberFormat,
  showComparison: boolean = true,
  indent = 1,
) {
  if (!rows) return null;
  return rows.map((r) => (
    <TableRow key={r.label} sx={r.manual ? { bgcolor: "#fdecea" } : undefined}>
      <TableCell sx={{ pl: 4 * (indent + (r.indent ?? 0)) }}>{r.label}</TableCell>
      <TableCell align="right">{r.group ? "" : fmt.format(r.amount)}</TableCell>
      <TableCell align="right">
        {!r.group && showComparison && r.change !== undefined ? (
          <Box sx={{ display: "flex", alignItems: "center", justifyContent: "flex-end", gap: 1 }}>
            <Typography
              variant="body2"
              sx={{
                color: r.change > 0 ? "success.main" : r.change < 0 ? "error.main" : "text.secondary",
                fontWeight: "medium",
              }}
            >
              {r.change > 0 ? "+" : ""}{fmt.format(r.change)}
            </Typography>
            <Typography
              variant="caption"
              sx={{
                color: r.changePercent && r.changePercent > 0 ? "success.main" : 
                       r.changePercent && r.changePercent < 0 ? "error.main" : "text.secondary",
                fontWeight: "medium",
              }}
            >
              ({r.changePercent && r.changePercent > 0 ? "+" : ""}{r.changePercent?.toFixed(1) ?? 0}%)
            </Typography>
          </Box>
        ) : showComparison ? (
          "—"
        ) : (
          r.percent != null && !r.group ? fmt.format(r.percent) : ""
        )}
      </TableCell>
    </TableRow>
  ));
}

export default function PLPage() {
  const now = new Date();
  const [periodType, setPeriodType] = useState<"Monthly" | "Yearly">("Monthly");
  const [month, setMonth] = useState(now.getMonth() + 1);
  const [year, setYear] = useState(now.getFullYear());
  const [store, setStore] = useState("All");
  const [stores, setStores] = useState<Store[]>([]);
  const [data, setData] = useState<ProfitLoss | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showComparison, setShowComparison] = useState(true);

  const numFmt = new Intl.NumberFormat("id-ID", {
    maximumFractionDigits: 0,
  });

  useEffect(() => {
    listAllStores().then((s) => setStores(s));
  }, []);

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      setError(null);
      setData(null);
      try {
        const res = await fetchProfitLoss({
          type: periodType,
          month: periodType === "Monthly" ? month : undefined,
          year,
          store: store === "All" ? "" : store,
          comparison: showComparison,
        });
        setData(res.data);
      } catch (e: any) {
        setError(e.message);
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, [periodType, month, year, store, showComparison]);

  const endDate = new Date(
    year,
    periodType === "Monthly" ? month : 12,
    0,
  ).toLocaleDateString("id-ID", {
    day: "2-digit",
    month: "long",
    year: "numeric",
  });

  return (
    <Box sx={{ p: 2 }}>
      <Box sx={{ display: "flex", alignItems: "center", mb: 1 }}>
        <img src="/logo.png" alt="Logo" style={{ height: 40, marginRight: 16 }} />
        <Typography variant="h6" sx={{ flexGrow: 1, textAlign: "center" }}>
          Dropship × Ecommerce ERP
        </Typography>
      </Box>
      <Typography
        variant="subtitle1"
        sx={{ fontWeight: "bold", textAlign: "center" }}
      >
        PROFIT & LOSS STATEMENT
      </Typography>
      <Typography variant="subtitle2" sx={{ textAlign: "center", mb: 2 }}>
        For the period ended {endDate}
      </Typography>

      <Box sx={{ display: "flex", gap: 2, alignItems: "center", mb: 2, flexWrap: "wrap" }}>
        <Typography sx={{ minWidth: 50 }}>Period:</Typography>
        <FormControl size="small">
          <InputLabel id="type-label">Type</InputLabel>
          <Select
            labelId="type-label"
            value={periodType}
            label="Type"
            onChange={(e) => setPeriodType(e.target.value as any)}
          >
            <MenuItem value="Monthly">Monthly</MenuItem>
            <MenuItem value="Yearly">Yearly</MenuItem>
          </Select>
        </FormControl>
        {periodType === "Monthly" && (
          <FormControl size="small">
            <InputLabel id="month-label">Month</InputLabel>
            <Select
              labelId="month-label"
              value={month}
              label="Month"
              onChange={(e) => setMonth(Number(e.target.value))}
            >
              {months.map((m, idx) => (
                <MenuItem key={m} value={idx + 1}>
                  {m}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
        )}
        <FormControl size="small">
          <InputLabel id="year-label">Year</InputLabel>
          <Select
            labelId="year-label"
            value={year}
            label="Year"
            onChange={(e) => setYear(Number(e.target.value))}
          >
            {years.map((y) => (
              <MenuItem key={y} value={y}>
                {y}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
        <Typography sx={{ minWidth: 50 }}>Store:</Typography>
        <FormControl size="small" sx={{ minWidth: 160 }}>
          <InputLabel id="store-label">Store</InputLabel>
          <Select
            labelId="store-label"
            value={store}
            label="Store"
            onChange={(e) => setStore(e.target.value)}
          >
            <MenuItem value="All">All</MenuItem>
            {stores.map((s) => (
              <MenuItem key={s.store_id} value={s.nama_toko}>
                {s.nama_toko}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
        <FormControl size="small">
          <InputLabel id="comparison-label">View</InputLabel>
          <Select
            labelId="comparison-label"
            value={showComparison ? "comparison" : "percentage"}
            label="View"
            onChange={(e) => setShowComparison(e.target.value === "comparison")}
          >
            <MenuItem value="comparison">Period Comparison</MenuItem>
            <MenuItem value="percentage">Percentage View</MenuItem>
          </Select>
        </FormControl>
      </Box>

      {loading && <CircularProgress />}
      {error && <Alert severity="error">{error}</Alert>}

      {data && (
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell sx={{ fontWeight: "bold" }}>Description</TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>Amount</TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>
                {showComparison ? "Change vs Previous" : "% of Revenue"}
              </TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            <TableRow>
              <TableCell colSpan={3} sx={{ bgcolor: "#eee", fontWeight: "bold" }}>
                REVENUE
              </TableCell>
            </TableRow>
            {renderRows(data.pendapatanUsaha, numFmt, showComparison)}
            <TableRow>
              <TableCell sx={{ fontWeight: "bold" }}>Total Revenue</TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>
                {numFmt.format(data.totalPendapatanUsaha)}
              </TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>
                {showComparison && data.prevTotalPendapatanUsaha !== undefined ? (
                  <Box sx={{ display: "flex", alignItems: "center", justifyContent: "flex-end", gap: 1 }}>
                    <Typography
                      variant="body2"
                      sx={{
                        color: (data.totalPendapatanUsaha - data.prevTotalPendapatanUsaha) > 0 ? "success.main" : 
                               (data.totalPendapatanUsaha - data.prevTotalPendapatanUsaha) < 0 ? "error.main" : "text.secondary",
                        fontWeight: "bold",
                      }}
                    >
                      {(data.totalPendapatanUsaha - data.prevTotalPendapatanUsaha) > 0 ? "+" : ""}
                      {numFmt.format(data.totalPendapatanUsaha - data.prevTotalPendapatanUsaha)}
                    </Typography>
                    <Typography
                      variant="caption"
                      sx={{
                        color: data.prevTotalPendapatanUsaha > 0 ? 
                               (((data.totalPendapatanUsaha - data.prevTotalPendapatanUsaha) / data.prevTotalPendapatanUsaha) * 100 > 0 ? "success.main" : 
                                ((data.totalPendapatanUsaha - data.prevTotalPendapatanUsaha) / data.prevTotalPendapatanUsaha) * 100 < 0 ? "error.main" : "text.secondary") : "text.secondary",
                        fontWeight: "bold",
                      }}
                    >
                      ({data.prevTotalPendapatanUsaha > 0 ? 
                         (((data.totalPendapatanUsaha - data.prevTotalPendapatanUsaha) / data.prevTotalPendapatanUsaha) * 100 > 0 ? "+" : "") : ""}
                      {data.prevTotalPendapatanUsaha > 0 ? 
                         (((data.totalPendapatanUsaha - data.prevTotalPendapatanUsaha) / data.prevTotalPendapatanUsaha) * 100).toFixed(1) : "100.0"}%)
                    </Typography>
                  </Box>
                ) : ""}
              </TableCell>
            </TableRow>

            <TableRow>
              <TableCell colSpan={3} sx={{ bgcolor: "#eee", fontWeight: "bold" }}>
                COST OF GOODS SOLD
              </TableCell>
            </TableRow>
            {renderRows(data.hargaPokokPenjualan, numFmt, showComparison)}
            <TableRow>
              <TableCell sx={{ fontWeight: "bold" }}>Total Cost of Goods Sold</TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>
                {numFmt.format(data.totalHargaPokokPenjualan)}
              </TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>
                {showComparison && data.prevTotalHargaPokokPenjualan !== undefined ? (
                  <Box sx={{ display: "flex", alignItems: "center", justifyContent: "flex-end", gap: 1 }}>
                    <Typography
                      variant="body2"
                      sx={{
                        color: (data.totalHargaPokokPenjualan - data.prevTotalHargaPokokPenjualan) > 0 ? "error.main" : 
                               (data.totalHargaPokokPenjualan - data.prevTotalHargaPokokPenjualan) < 0 ? "success.main" : "text.secondary",
                        fontWeight: "bold",
                      }}
                    >
                      {(data.totalHargaPokokPenjualan - data.prevTotalHargaPokokPenjualan) > 0 ? "+" : ""}
                      {numFmt.format(data.totalHargaPokokPenjualan - data.prevTotalHargaPokokPenjualan)}
                    </Typography>
                    <Typography
                      variant="caption"
                      sx={{
                        color: data.prevTotalHargaPokokPenjualan > 0 ? 
                               (((data.totalHargaPokokPenjualan - data.prevTotalHargaPokokPenjualan) / data.prevTotalHargaPokokPenjualan) * 100 > 0 ? "error.main" : 
                                ((data.totalHargaPokokPenjualan - data.prevTotalHargaPokokPenjualan) / data.prevTotalHargaPokokPenjualan) * 100 < 0 ? "success.main" : "text.secondary") : "text.secondary",
                        fontWeight: "bold",
                      }}
                    >
                      ({data.prevTotalHargaPokokPenjualan > 0 ? 
                         (((data.totalHargaPokokPenjualan - data.prevTotalHargaPokokPenjualan) / data.prevTotalHargaPokokPenjualan) * 100 > 0 ? "+" : "") : ""}
                      {data.prevTotalHargaPokokPenjualan > 0 ? 
                         (((data.totalHargaPokokPenjualan - data.prevTotalHargaPokokPenjualan) / data.prevTotalHargaPokokPenjualan) * 100).toFixed(1) : "100.0"}%)
                    </Typography>
                  </Box>
                ) : ""}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell sx={{ fontWeight: "bold" }}>Gross Profit</TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>
                {numFmt.format(data.labaKotor.amount)}
              </TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>
                {showComparison && data.labaKotor.change !== undefined ? (
                  <Box sx={{ display: "flex", alignItems: "center", justifyContent: "flex-end", gap: 1 }}>
                    <Typography
                      variant="body2"
                      sx={{
                        color: data.labaKotor.change > 0 ? "success.main" : 
                               data.labaKotor.change < 0 ? "error.main" : "text.secondary",
                        fontWeight: "bold",
                      }}
                    >
                      {data.labaKotor.change > 0 ? "+" : ""}{numFmt.format(data.labaKotor.change)}
                    </Typography>
                    <Typography
                      variant="caption"
                      sx={{
                        color: data.labaKotor.changePercent && data.labaKotor.changePercent > 0 ? "success.main" : 
                               data.labaKotor.changePercent && data.labaKotor.changePercent < 0 ? "error.main" : "text.secondary",
                        fontWeight: "bold",
                      }}
                    >
                      ({data.labaKotor.changePercent && data.labaKotor.changePercent > 0 ? "+" : ""}{data.labaKotor.changePercent?.toFixed(1) ?? 0}%)
                    </Typography>
                  </Box>
                ) : (
                  data.labaKotor.percent != null ? numFmt.format(data.labaKotor.percent) : ""
                )}
              </TableCell>
            </TableRow>

            <TableRow>
              <TableCell colSpan={3} sx={{ bgcolor: "#eee", fontWeight: "bold" }}>
                OPERATING EXPENSES
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell colSpan={3} sx={{ bgcolor: "#f5f5f5", fontWeight: "bold", pl: 4 }}>
                Operating Expenses
              </TableCell>
            </TableRow>
            {renderRows(data.bebanOperasional, numFmt, showComparison, 2)}
            <TableRow>
              <TableCell sx={{ fontWeight: "bold" }}>Total Operating Expenses</TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>
                {numFmt.format(data.totalBebanOperasional)}
              </TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>
                {showComparison && data.prevTotalBebanOperasional !== undefined ? (
                  <Box sx={{ display: "flex", alignItems: "center", justifyContent: "flex-end", gap: 1 }}>
                    <Typography
                      variant="body2"
                      sx={{
                        color: (data.totalBebanOperasional - data.prevTotalBebanOperasional) > 0 ? "error.main" : 
                               (data.totalBebanOperasional - data.prevTotalBebanOperasional) < 0 ? "success.main" : "text.secondary",
                        fontWeight: "bold",
                      }}
                    >
                      {(data.totalBebanOperasional - data.prevTotalBebanOperasional) > 0 ? "+" : ""}
                      {numFmt.format(data.totalBebanOperasional - data.prevTotalBebanOperasional)}
                    </Typography>
                    <Typography
                      variant="caption"
                      sx={{
                        color: data.prevTotalBebanOperasional > 0 ? 
                               (((data.totalBebanOperasional - data.prevTotalBebanOperasional) / data.prevTotalBebanOperasional) * 100 > 0 ? "error.main" : 
                                ((data.totalBebanOperasional - data.prevTotalBebanOperasional) / data.prevTotalBebanOperasional) * 100 < 0 ? "success.main" : "text.secondary") : "text.secondary",
                        fontWeight: "bold",
                      }}
                    >
                      ({data.prevTotalBebanOperasional > 0 ? 
                         (((data.totalBebanOperasional - data.prevTotalBebanOperasional) / data.prevTotalBebanOperasional) * 100 > 0 ? "+" : "") : ""}
                      {data.prevTotalBebanOperasional > 0 ? 
                         (((data.totalBebanOperasional - data.prevTotalBebanOperasional) / data.prevTotalBebanOperasional) * 100).toFixed(1) : "100.0"}%)
                    </Typography>
                  </Box>
                ) : ""}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell colSpan={3} sx={{ bgcolor: "#f5f5f5", fontWeight: "bold", pl: 4 }}>
                Marketing Expenses
              </TableCell>
            </TableRow>
            {renderRows(data.bebanPemasaran, numFmt, showComparison, 2)}
            <TableRow>
              <TableCell sx={{ fontWeight: "bold" }}>Total Marketing Expenses</TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>
                {numFmt.format(data.totalBebanPemasaran)}
              </TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>
                {showComparison && data.prevTotalBebanPemasaran !== undefined ? (
                  <Box sx={{ display: "flex", alignItems: "center", justifyContent: "flex-end", gap: 1 }}>
                    <Typography
                      variant="body2"
                      sx={{
                        color: (data.totalBebanPemasaran - data.prevTotalBebanPemasaran) > 0 ? "error.main" : 
                               (data.totalBebanPemasaran - data.prevTotalBebanPemasaran) < 0 ? "success.main" : "text.secondary",
                        fontWeight: "bold",
                      }}
                    >
                      {(data.totalBebanPemasaran - data.prevTotalBebanPemasaran) > 0 ? "+" : ""}
                      {numFmt.format(data.totalBebanPemasaran - data.prevTotalBebanPemasaran)}
                    </Typography>
                    <Typography
                      variant="caption"
                      sx={{
                        color: data.prevTotalBebanPemasaran > 0 ? 
                               (((data.totalBebanPemasaran - data.prevTotalBebanPemasaran) / data.prevTotalBebanPemasaran) * 100 > 0 ? "error.main" : 
                                ((data.totalBebanPemasaran - data.prevTotalBebanPemasaran) / data.prevTotalBebanPemasaran) * 100 < 0 ? "success.main" : "text.secondary") : "text.secondary",
                        fontWeight: "bold",
                      }}
                    >
                      ({data.prevTotalBebanPemasaran > 0 ? 
                         (((data.totalBebanPemasaran - data.prevTotalBebanPemasaran) / data.prevTotalBebanPemasaran) * 100 > 0 ? "+" : "") : ""}
                      {data.prevTotalBebanPemasaran > 0 ? 
                         (((data.totalBebanPemasaran - data.prevTotalBebanPemasaran) / data.prevTotalBebanPemasaran) * 100).toFixed(1) : "100.0"}%)
                    </Typography>
                  </Box>
                ) : ""}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell colSpan={3} sx={{ bgcolor: "#f5f5f5", fontWeight: "bold", pl: 4 }}>
                Administrative Expenses
              </TableCell>
            </TableRow>
            {renderRows(data.bebanAdministrasi, numFmt, showComparison, 2)}
            <TableRow>
              <TableCell sx={{ fontWeight: "bold" }}>Total Administrative Expenses</TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>
                {numFmt.format(data.totalBebanAdministrasi)}
              </TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>
                {showComparison && data.prevTotalBebanAdministrasi !== undefined ? (
                  <Box sx={{ display: "flex", alignItems: "center", justifyContent: "flex-end", gap: 1 }}>
                    <Typography
                      variant="body2"
                      sx={{
                        color: (data.totalBebanAdministrasi - data.prevTotalBebanAdministrasi) > 0 ? "error.main" : 
                               (data.totalBebanAdministrasi - data.prevTotalBebanAdministrasi) < 0 ? "success.main" : "text.secondary",
                        fontWeight: "bold",
                      }}
                    >
                      {(data.totalBebanAdministrasi - data.prevTotalBebanAdministrasi) > 0 ? "+" : ""}
                      {numFmt.format(data.totalBebanAdministrasi - data.prevTotalBebanAdministrasi)}
                    </Typography>
                    <Typography
                      variant="caption"
                      sx={{
                        color: data.prevTotalBebanAdministrasi > 0 ? 
                               (((data.totalBebanAdministrasi - data.prevTotalBebanAdministrasi) / data.prevTotalBebanAdministrasi) * 100 > 0 ? "error.main" : 
                                ((data.totalBebanAdministrasi - data.prevTotalBebanAdministrasi) / data.prevTotalBebanAdministrasi) * 100 < 0 ? "success.main" : "text.secondary") : "text.secondary",
                        fontWeight: "bold",
                      }}
                    >
                      ({data.prevTotalBebanAdministrasi > 0 ? 
                         (((data.totalBebanAdministrasi - data.prevTotalBebanAdministrasi) / data.prevTotalBebanAdministrasi) * 100 > 0 ? "+" : "") : ""}
                      {data.prevTotalBebanAdministrasi > 0 ? 
                         (((data.totalBebanAdministrasi - data.prevTotalBebanAdministrasi) / data.prevTotalBebanAdministrasi) * 100).toFixed(1) : "100.0"}%)
                    </Typography>
                  </Box>
                ) : ""}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell sx={{ fontWeight: "bold" }}>Total Operating Expenses</TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>
                {numFmt.format(data.totalBebanUsaha.amount)}
              </TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>
                {showComparison && data.totalBebanUsaha.change !== undefined ? (
                  <Box sx={{ display: "flex", alignItems: "center", justifyContent: "flex-end", gap: 1 }}>
                    <Typography
                      variant="body2"
                      sx={{
                        color: data.totalBebanUsaha.change > 0 ? "error.main" : 
                               data.totalBebanUsaha.change < 0 ? "success.main" : "text.secondary",
                        fontWeight: "bold",
                      }}
                    >
                      {data.totalBebanUsaha.change > 0 ? "+" : ""}{numFmt.format(data.totalBebanUsaha.change)}
                    </Typography>
                    <Typography
                      variant="caption"
                      sx={{
                        color: data.totalBebanUsaha.changePercent && data.totalBebanUsaha.changePercent > 0 ? "error.main" : 
                               data.totalBebanUsaha.changePercent && data.totalBebanUsaha.changePercent < 0 ? "success.main" : "text.secondary",
                        fontWeight: "bold",
                      }}
                    >
                      ({data.totalBebanUsaha.changePercent && data.totalBebanUsaha.changePercent > 0 ? "+" : ""}{data.totalBebanUsaha.changePercent?.toFixed(1) ?? 0}%)
                    </Typography>
                  </Box>
                ) : (
                  data.totalBebanUsaha.percent != null ? numFmt.format(data.totalBebanUsaha.percent) : ""
                )}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell sx={{ fontWeight: "bold" }}>Operating Income</TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>
                {numFmt.format(data.labaSebelumPajak)}
              </TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>
                {showComparison && data.prevLabaSebelumPajak !== undefined ? (
                  <Box sx={{ display: "flex", alignItems: "center", justifyContent: "flex-end", gap: 1 }}>
                    <Typography
                      variant="body2"
                      sx={{
                        color: (data.labaSebelumPajak - data.prevLabaSebelumPajak) > 0 ? "success.main" : 
                               (data.labaSebelumPajak - data.prevLabaSebelumPajak) < 0 ? "error.main" : "text.secondary",
                        fontWeight: "bold",
                      }}
                    >
                      {(data.labaSebelumPajak - data.prevLabaSebelumPajak) > 0 ? "+" : ""}
                      {numFmt.format(data.labaSebelumPajak - data.prevLabaSebelumPajak)}
                    </Typography>
                    <Typography
                      variant="caption"
                      sx={{
                        color: data.prevLabaSebelumPajak > 0 ? 
                               (((data.labaSebelumPajak - data.prevLabaSebelumPajak) / data.prevLabaSebelumPajak) * 100 > 0 ? "success.main" : 
                                ((data.labaSebelumPajak - data.prevLabaSebelumPajak) / data.prevLabaSebelumPajak) * 100 < 0 ? "error.main" : "text.secondary") : "text.secondary",
                        fontWeight: "bold",
                      }}
                    >
                      ({data.prevLabaSebelumPajak > 0 ? 
                         (((data.labaSebelumPajak - data.prevLabaSebelumPajak) / data.prevLabaSebelumPajak) * 100 > 0 ? "+" : "") : ""}
                      {data.prevLabaSebelumPajak > 0 ? 
                         (((data.labaSebelumPajak - data.prevLabaSebelumPajak) / data.prevLabaSebelumPajak) * 100).toFixed(1) : "100.0"}%)
                    </Typography>
                  </Box>
                ) : ""}
              </TableCell>
            </TableRow>

            <TableRow>
              <TableCell colSpan={3} sx={{ bgcolor: "#eee", fontWeight: "bold" }}>
                TAXES
              </TableCell>
            </TableRow>
            {renderRows(data.pajakPenghasilan, numFmt, showComparison)}
            <TableRow>
              <TableCell sx={{ fontWeight: "bold" }}>Total Taxes</TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>
                {numFmt.format(data.totalPajakPenghasilan)}
              </TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>
                {showComparison && data.prevTotalPajakPenghasilan !== undefined ? (
                  <Box sx={{ display: "flex", alignItems: "center", justifyContent: "flex-end", gap: 1 }}>
                    <Typography
                      variant="body2"
                      sx={{
                        color: (data.totalPajakPenghasilan - data.prevTotalPajakPenghasilan) > 0 ? "error.main" : 
                               (data.totalPajakPenghasilan - data.prevTotalPajakPenghasilan) < 0 ? "success.main" : "text.secondary",
                        fontWeight: "bold",
                      }}
                    >
                      {(data.totalPajakPenghasilan - data.prevTotalPajakPenghasilan) > 0 ? "+" : ""}
                      {numFmt.format(data.totalPajakPenghasilan - data.prevTotalPajakPenghasilan)}
                    </Typography>
                    <Typography
                      variant="caption"
                      sx={{
                        color: data.prevTotalPajakPenghasilan > 0 ? 
                               (((data.totalPajakPenghasilan - data.prevTotalPajakPenghasilan) / data.prevTotalPajakPenghasilan) * 100 > 0 ? "error.main" : 
                                ((data.totalPajakPenghasilan - data.prevTotalPajakPenghasilan) / data.prevTotalPajakPenghasilan) * 100 < 0 ? "success.main" : "text.secondary") : "text.secondary",
                        fontWeight: "bold",
                      }}
                    >
                      ({data.prevTotalPajakPenghasilan > 0 ? 
                         (((data.totalPajakPenghasilan - data.prevTotalPajakPenghasilan) / data.prevTotalPajakPenghasilan) * 100 > 0 ? "+" : "") : ""}
                      {data.prevTotalPajakPenghasilan > 0 ? 
                         (((data.totalPajakPenghasilan - data.prevTotalPajakPenghasilan) / data.prevTotalPajakPenghasilan) * 100).toFixed(1) : "100.0"}%)
                    </Typography>
                  </Box>
                ) : ""}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell sx={{ fontWeight: "bold", fontSize: "1.1rem" }}>Net Income</TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold", fontSize: "1.1rem" }}>
                {numFmt.format(data.labaRugiBersih.amount)}
              </TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold", fontSize: "1.1rem" }}>
                {showComparison && data.labaRugiBersih.change !== undefined ? (
                  <Box sx={{ display: "flex", alignItems: "center", justifyContent: "flex-end", gap: 1 }}>
                    <Typography
                      variant="body2"
                      sx={{
                        color: data.labaRugiBersih.change > 0 ? "success.main" : 
                               data.labaRugiBersih.change < 0 ? "error.main" : "text.secondary",
                        fontWeight: "bold",
                        fontSize: "1.1rem",
                      }}
                    >
                      {data.labaRugiBersih.change > 0 ? "+" : ""}{numFmt.format(data.labaRugiBersih.change)}
                    </Typography>
                    <Typography
                      variant="caption"
                      sx={{
                        color: data.labaRugiBersih.changePercent && data.labaRugiBersih.changePercent > 0 ? "success.main" : 
                               data.labaRugiBersih.changePercent && data.labaRugiBersih.changePercent < 0 ? "error.main" : "text.secondary",
                        fontWeight: "bold",
                        fontSize: "0.9rem",
                      }}
                    >
                      ({data.labaRugiBersih.changePercent && data.labaRugiBersih.changePercent > 0 ? "+" : ""}{data.labaRugiBersih.changePercent?.toFixed(1) ?? 0}%)
                    </Typography>
                  </Box>
                ) : (
                  data.labaRugiBersih.percent != null ? numFmt.format(data.labaRugiBersih.percent) : ""
                )}
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )}

      <Box sx={{ display: "flex", alignItems: "center", justifyContent: "space-between", mt: 4 }}>
        <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
          <WarningIcon color="warning" fontSize="small" />
          <Typography variant="caption">
            Red background indicates manual data entry
          </Typography>
        </Box>
        <Typography variant="body2" sx={{ fontStyle: "italic" }}>
          Prepared by: ___________________
        </Typography>
      </Box>
    </Box>
  );
}
