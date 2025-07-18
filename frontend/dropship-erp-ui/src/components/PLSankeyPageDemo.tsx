import { useState } from "react";
import {
  Box,
  Typography,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Card,
  CardContent,
} from "@mui/material";
import { ResponsiveSankey } from "@nivo/sankey";

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

// Sample data for demonstration
const sampleSankeyData: SankeyData = {
  nodes: [
    { id: "Product Sales", nodeColor: "#81C784" },
    { id: "Service Revenue", nodeColor: "#81C784" },
    { id: "Affiliate Income", nodeColor: "#81C784" },
    { id: "Total Revenue", nodeColor: "#4CAF50" },
    { id: "Cost of Goods Sold", nodeColor: "#F44336" },
    { id: "Operating Expenses", nodeColor: "#FF9800" },
    { id: "Marketing Expenses", nodeColor: "#2196F3" },
    { id: "Administrative Expenses", nodeColor: "#9C27B0" },
    { id: "Taxes", nodeColor: "#795548" },
    { id: "Net Income", nodeColor: "#8BC34A" },
  ],
  links: [
    // Revenue sources flowing to Total Revenue
    { source: "Product Sales", target: "Total Revenue", value: 150000000 },
    { source: "Service Revenue", target: "Total Revenue", value: 25000000 },
    { source: "Affiliate Income", target: "Total Revenue", value: 15000000 },
    
    // Total Revenue flowing to expenses and net income
    { source: "Total Revenue", target: "Cost of Goods Sold", value: 80000000 },
    { source: "Total Revenue", target: "Operating Expenses", value: 35000000 },
    { source: "Total Revenue", target: "Marketing Expenses", value: 25000000 },
    { source: "Total Revenue", target: "Administrative Expenses", value: 15000000 },
    { source: "Total Revenue", target: "Taxes", value: 10000000 },
    { source: "Total Revenue", target: "Net Income", value: 25000000 },
  ],
};

export default function PLSankeyPageDemo() {
  const now = new Date();
  const [periodType, setPeriodType] = useState<"Monthly" | "Yearly">("Monthly");
  const [month, setMonth] = useState(now.getMonth() + 1);
  const [year, setYear] = useState(now.getFullYear());
  const [store, setStore] = useState("All");
  const [sankeyData] = useState<SankeyData>(sampleSankeyData);

  const numFmt = new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    maximumFractionDigits: 0,
  });

  const endDate = new Date(
    year,
    periodType === "Monthly" ? month : 12,
    0,
  ).toLocaleDateString("id-ID", {
    day: "2-digit",
    month: "long",
    year: "numeric",
  });

  const totalRevenue = sankeyData.links
    .filter(link => link.target === "Total Revenue")
    .reduce((sum, link) => sum + link.value, 0);

  const totalExpenses = sankeyData.links
    .filter(link => link.source === "Total Revenue" && link.target !== "Net Income")
    .reduce((sum, link) => sum + link.value, 0);

  const netIncome = sankeyData.links
    .find(link => link.source === "Total Revenue" && link.target === "Net Income")?.value || 0;

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
        PROFIT & LOSS SANKEY DIAGRAM (DEMO)
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
            <MenuItem value="Store A">Store A</MenuItem>
            <MenuItem value="Store B">Store B</MenuItem>
          </Select>
        </FormControl>
      </Box>

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

      <Card>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Summary
          </Typography>
          <Box sx={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))", gap: 2 }}>
            <Box>
              <Typography variant="body2" color="text.secondary">Total Revenue</Typography>
              <Typography variant="h6" color="success.main">
                {numFmt.format(totalRevenue)}
              </Typography>
            </Box>
            <Box>
              <Typography variant="body2" color="text.secondary">Total Expenses</Typography>
              <Typography variant="h6" color="error.main">
                {numFmt.format(totalExpenses)}
              </Typography>
            </Box>
            <Box>
              <Typography variant="body2" color="text.secondary">Net Income</Typography>
              <Typography 
                variant="h6" 
                color={netIncome >= 0 ? "success.main" : "error.main"}
              >
                {numFmt.format(netIncome)}
              </Typography>
            </Box>
          </Box>
        </CardContent>
      </Card>

      <Box sx={{ mt: 2, p: 2, backgroundColor: "#f5f5f5", borderRadius: 1 }}>
        <Typography variant="body2" color="text.secondary">
          <strong>Note:</strong> This is a demonstration page with sample data. 
          The actual PLSankeyPage will connect to the backend API to fetch real profit & loss data.
        </Typography>
      </Box>
    </Box>
  );
}