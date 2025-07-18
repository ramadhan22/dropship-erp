import { useEffect, useState } from "react";
import {
  Box,
  Typography,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  CircularProgress,
  Alert,
  Card,
  CardContent,
} from "@mui/material";
import { ResponsiveSankey } from "@nivo/sankey";
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

interface SankeyNode {
  id: string;
  nodeColor?: string;
}

interface SankeyLink {
  source: string;
  target: string;
  value: number;
}

interface SankeyData {
  nodes: SankeyNode[];
  links: SankeyLink[];
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

const transformToSankeyData = (data: ProfitLoss): SankeyData => {
  const nodes: SankeyNode[] = [];
  const links: SankeyLink[] = [];

  // Add nodes
  nodes.push({ id: "Total Revenue", nodeColor: "#4CAF50" });
  nodes.push({ id: "Cost of Goods Sold", nodeColor: "#F44336" });
  nodes.push({ id: "Operating Expenses", nodeColor: "#FF9800" });
  nodes.push({ id: "Marketing Expenses", nodeColor: "#2196F3" });
  nodes.push({ id: "Administrative Expenses", nodeColor: "#9C27B0" });
  nodes.push({ id: "Taxes", nodeColor: "#795548" });
  nodes.push({ id: "Net Income", nodeColor: "#8BC34A" });

  // Add revenue source nodes
  data.pendapatanUsaha?.forEach(row => {
    if (row.amount > 0) {
      nodes.push({ id: row.label, nodeColor: "#81C784" });
      links.push({
        source: row.label,
        target: "Total Revenue",
        value: Math.abs(row.amount),
      });
    }
  });

  // Add main flows from Total Revenue
  if (data.totalHargaPokokPenjualan > 0) {
    links.push({
      source: "Total Revenue",
      target: "Cost of Goods Sold",
      value: data.totalHargaPokokPenjualan,
    });
  }

  if (data.totalBebanOperasional > 0) {
    links.push({
      source: "Total Revenue",
      target: "Operating Expenses",
      value: data.totalBebanOperasional,
    });
  }

  if (data.totalBebanPemasaran > 0) {
    links.push({
      source: "Total Revenue",
      target: "Marketing Expenses",
      value: data.totalBebanPemasaran,
    });
  }

  if (data.totalBebanAdministrasi > 0) {
    links.push({
      source: "Total Revenue",
      target: "Administrative Expenses",
      value: data.totalBebanAdministrasi,
    });
  }

  if (data.totalPajakPenghasilan > 0) {
    links.push({
      source: "Total Revenue",
      target: "Taxes",
      value: data.totalPajakPenghasilan,
    });
  }

  // Net Income (what remains)
  if (data.labaRugiBersih.amount > 0) {
    links.push({
      source: "Total Revenue",
      target: "Net Income",
      value: data.labaRugiBersih.amount,
    });
  }

  return { nodes, links };
};

export default function PLSankeyPage() {
  const now = new Date();
  const [periodType, setPeriodType] = useState<"Monthly" | "Yearly">("Monthly");
  const [month, setMonth] = useState(now.getMonth() + 1);
  const [year, setYear] = useState(now.getFullYear());
  const [store, setStore] = useState("All");
  const [stores, setStores] = useState<Store[]>([]);
  const [data, setData] = useState<ProfitLoss | null>(null);
  const [sankeyData, setSankeyData] = useState<SankeyData | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const numFmt = new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
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
      setSankeyData(null);
      try {
        const res = await fetchProfitLoss({
          type: periodType,
          month: periodType === "Monthly" ? month : undefined,
          year,
          store: store === "All" ? "" : store,
          comparison: false,
        });
        setData(res.data);
        setSankeyData(transformToSankeyData(res.data));
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
        PROFIT & LOSS SANKEY DIAGRAM
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
      </Box>

      {loading && (
        <Box sx={{ display: "flex", justifyContent: "center", my: 4 }}>
          <CircularProgress />
        </Box>
      )}
      
      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      {sankeyData && !loading && (
        <Card sx={{ mb: 2 }}>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              Money Flow Visualization
            </Typography>
            <Typography variant="body2" color="text.secondary" gutterBottom>
              This diagram shows how revenue flows through various expenses to net income.
              The width of each flow represents the relative amount of money.
            </Typography>
            
            <Box sx={{ height: 600, mt: 2 }}>
              <ResponsiveSankey
                data={sankeyData}
                margin={{ top: 40, right: 160, bottom: 40, left: 50 }}
                align="justify"
                colors={{ scheme: 'category10' }}
                nodeOpacity={1}
                nodeHoverOthersOpacity={0.35}
                nodeThickness={18}
                nodeSpacing={24}
                nodeBorderWidth={0}
                nodeBorderColor={{
                  from: 'color',
                  modifiers: [['darker', 0.8]],
                }}
                linkOpacity={0.5}
                linkHoverOthersOpacity={0.1}
                linkContract={3}
                enableLinkGradient={true}
                labelPosition="outside"
                labelOrientation="vertical"
                labelPadding={16}
                labelTextColor={{
                  from: 'color',
                  modifiers: [['darker', 1]],
                }}
                legends={[
                  {
                    anchor: 'bottom-right',
                    direction: 'column',
                    translateX: 130,
                    itemWidth: 100,
                    itemHeight: 14,
                    itemDirection: 'right-to-left',
                    itemsSpacing: 2,
                    itemTextColor: '#999',
                    symbolSize: 14,
                    effects: [
                      {
                        on: 'hover',
                        style: {
                          itemTextColor: '#000',
                        },
                      },
                    ],
                  },
                ]}
              />
            </Box>
          </CardContent>
        </Card>
      )}

      {data && !loading && (
        <Card>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              Summary
            </Typography>
            <Box sx={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))", gap: 2 }}>
              <Box>
                <Typography variant="body2" color="text.secondary">Total Revenue</Typography>
                <Typography variant="h6" color="success.main">
                  {numFmt.format(data.totalPendapatanUsaha)}
                </Typography>
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">Total Expenses</Typography>
                <Typography variant="h6" color="error.main">
                  {numFmt.format(
                    data.totalHargaPokokPenjualan + 
                    data.totalBebanOperasional + 
                    data.totalBebanPemasaran + 
                    data.totalBebanAdministrasi + 
                    data.totalPajakPenghasilan
                  )}
                </Typography>
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">Net Income</Typography>
                <Typography 
                  variant="h6" 
                  color={data.labaRugiBersih.amount >= 0 ? "success.main" : "error.main"}
                >
                  {numFmt.format(data.labaRugiBersih.amount)}
                </Typography>
              </Box>
            </Box>
          </CardContent>
        </Card>
      )}
    </Box>
  );
}