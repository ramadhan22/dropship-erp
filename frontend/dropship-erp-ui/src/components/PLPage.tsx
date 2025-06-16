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
  manual?: boolean;
}

interface ProfitLoss {
  pendapatanUsaha: PLRow[];
  totalPendapatanUsaha: number;
  hargaPokokPenjualan: PLRow[];
  totalHargaPokokPenjualan: number;
  labaKotor: { amount: number; percent: number };
  bebanOperasional: PLRow[];
  totalBebanOperasional: number;
  bebanAdministrasi: PLRow[];
  totalBebanAdministrasi: number;
  totalBebanUsaha: { amount: number; percent: number };
  labaSebelumPajak: number;
  pajakPenghasilan: PLRow[];
  totalPajakPenghasilan: number;
  labaRugiBersih: { amount: number; percent: number };
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

function renderRows(rows: PLRow[], fmt: Intl.NumberFormat, indent = 1) {
  return rows.map((r) => (
    <TableRow key={r.label} sx={r.manual ? { bgcolor: "#fdecea" } : undefined}>
      <TableCell sx={{ pl: 4 * indent }}>{r.label}</TableCell>
      <TableCell align="right">{fmt.format(r.amount)}</TableCell>
      <TableCell align="right">
        {r.percent != null ? fmt.format(r.percent) : ""}
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
        });
        setData(res.data);
      } catch (e: any) {
        setError(e.message);
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, [periodType, month, year, store]);

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
        LAPORAN LABA RUGI
      </Typography>
      <Typography variant="subtitle2" sx={{ textAlign: "center", mb: 2 }}>
        Periode yang Berakhir pada {endDate} (dalam satuan rupiah)
      </Typography>

      <Box sx={{ display: "flex", gap: 2, alignItems: "center", mb: 2 }}>
        <Typography sx={{ minWidth: 70 }}>Periode:</Typography>
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
      </Box>

      {loading && <CircularProgress />}
      {error && <Alert severity="error">{error}</Alert>}

      {data && (
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Keterangan</TableCell>
              <TableCell align="right">Jumlah</TableCell>
              <TableCell align="right">%</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            <TableRow>
              <TableCell colSpan={3} sx={{ bgcolor: "#eee", fontWeight: "bold" }}>
                Pendapatan Usaha
              </TableCell>
            </TableRow>
            {renderRows(data.pendapatanUsaha, numFmt)}
            <TableRow>
              <TableCell sx={{ fontWeight: "bold" }}>Jumlah Pendapatan Usaha</TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>
                {numFmt.format(data.totalPendapatanUsaha)}
              </TableCell>
              <TableCell />
            </TableRow>

            <TableRow>
              <TableCell colSpan={3} sx={{ bgcolor: "#eee", fontWeight: "bold" }}>
                Harga Pokok Penjualan
              </TableCell>
            </TableRow>
            {renderRows(data.hargaPokokPenjualan, numFmt)}
            <TableRow>
              <TableCell sx={{ fontWeight: "bold" }}>Jumlah Harga Pokok Penjualan</TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>
                {numFmt.format(data.totalHargaPokokPenjualan)}
              </TableCell>
              <TableCell />
            </TableRow>
            <TableRow>
              <TableCell sx={{ fontWeight: "bold" }}>Laba/Rugi Kotor</TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>
                {numFmt.format(data.labaKotor.amount)}
              </TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>
                {numFmt.format(data.labaKotor.percent)}
              </TableCell>
            </TableRow>

            <TableRow>
              <TableCell colSpan={3} sx={{ bgcolor: "#eee", fontWeight: "bold" }}>
                Beban Usaha
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell colSpan={3} sx={{ bgcolor: "#eee", fontWeight: "bold", pl: 4 }}>
                Beban Operasional
              </TableCell>
            </TableRow>
            {renderRows(data.bebanOperasional, numFmt, 2)}
            <TableRow>
              <TableCell sx={{ fontWeight: "bold" }}>Jumlah Beban Operasional</TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>
                {numFmt.format(data.totalBebanOperasional)}
              </TableCell>
              <TableCell />
            </TableRow>
            <TableRow>
              <TableCell colSpan={3} sx={{ bgcolor: "#eee", fontWeight: "bold", pl: 4 }}>
                Beban Administrasi
              </TableCell>
            </TableRow>
            {renderRows(data.bebanAdministrasi, numFmt, 2)}
            <TableRow>
              <TableCell sx={{ fontWeight: "bold" }}>Jumlah Beban Administrasi</TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>
                {numFmt.format(data.totalBebanAdministrasi)}
              </TableCell>
              <TableCell />
            </TableRow>
            <TableRow>
              <TableCell sx={{ fontWeight: "bold" }}>Jumlah Beban Usaha / Pengeluaran</TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>
                {numFmt.format(data.totalBebanUsaha.amount)}
              </TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>
                {numFmt.format(data.totalBebanUsaha.percent)}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell sx={{ fontWeight: "bold" }}>Laba Usaha Sebelum Pajak</TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>
                {numFmt.format(data.labaSebelumPajak)}
              </TableCell>
              <TableCell />
            </TableRow>

            <TableRow>
              <TableCell colSpan={3} sx={{ bgcolor: "#eee", fontWeight: "bold" }}>
                Pajak Penghasilan
              </TableCell>
            </TableRow>
            {renderRows(data.pajakPenghasilan, numFmt)}
            <TableRow>
              <TableCell sx={{ fontWeight: "bold" }}>Jumlah Pajak Penghasilan</TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold" }}>
                {numFmt.format(data.totalPajakPenghasilan)}
              </TableCell>
              <TableCell />
            </TableRow>
            <TableRow>
              <TableCell sx={{ fontWeight: "bold", fontSize: "1.1rem" }}>Laba/Rugi Bersih</TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold", fontSize: "1.1rem" }}>
                {numFmt.format(data.labaRugiBersih.amount)}
              </TableCell>
              <TableCell align="right" sx={{ fontWeight: "bold", fontSize: "1.1rem" }}>
                {numFmt.format(data.labaRugiBersih.percent)}
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )}

      <Box sx={{ display: "flex", alignItems: "center", mt: 4 }}>
        <Typography sx={{ flexGrow: 1 }}>Dibuat Oleh, ___________________</Typography>
        <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
          <WarningIcon color="warning" fontSize="small" />
          <Typography variant="caption">
            Catatan penting!!! Warna Merah adalah input data manual
          </Typography>
        </Box>
      </Box>
    </Box>
  );
}
