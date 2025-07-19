import { 
  Box, 
  FormControl, 
  InputLabel, 
  Select, 
  MenuItem, 
  Chip, 
  OutlinedInput,
  Typography,
  Paper
} from '@mui/material';
import { DatePicker } from '@mui/x-date-pickers/DatePicker';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';
import { AdapterDateFns } from '@mui/x-date-pickers/AdapterDateFns';
import { colors, spacing } from '../theme/tokens';

export interface DashboardFilters {
  dateFrom: Date | null;
  dateTo: Date | null;
  stores: string[];
  channels: string[];
  currency?: string;
  period?: string;
}

interface DashboardFilterContainerProps {
  filters: DashboardFilters;
  onFiltersChange: (filters: DashboardFilters) => void;
  availableStores?: Array<{ id: string; name: string }>;
  availableChannels?: Array<{ id: string; name: string }>;
  showCurrencyToggle?: boolean;
  showPeriodToggle?: boolean;
}

const ITEM_HEIGHT = 48;
const ITEM_PADDING_TOP = 8;
const MenuProps = {
  PaperProps: {
    style: {
      maxHeight: ITEM_HEIGHT * 4.5 + ITEM_PADDING_TOP,
      width: 250,
    },
  },
};

const DEFAULT_STORES = [
  { id: 'store1', name: 'Main Store' },
  { id: 'store2', name: 'Secondary Store' },
  { id: 'store3', name: 'Outlet Store' },
];

const DEFAULT_CHANNELS = [
  { id: 'shopee', name: 'Shopee' },
  { id: 'tokopedia', name: 'Tokopedia' },
  { id: 'lazada', name: 'Lazada' },
];

export default function DashboardFilterContainer({
  filters,
  onFiltersChange,
  availableStores = DEFAULT_STORES,
  availableChannels = DEFAULT_CHANNELS,
  showCurrencyToggle = true,
  showPeriodToggle = true,
}: DashboardFilterContainerProps) {
  
  const handleDateFromChange = (date: Date | null) => {
    onFiltersChange({ ...filters, dateFrom: date });
  };

  const handleDateToChange = (date: Date | null) => {
    onFiltersChange({ ...filters, dateTo: date });
  };

  const handleStoresChange = (event: any) => {
    const value = event.target.value;
    onFiltersChange({ ...filters, stores: typeof value === 'string' ? value.split(',') : value });
  };

  const handleChannelsChange = (event: any) => {
    const value = event.target.value;
    onFiltersChange({ ...filters, channels: typeof value === 'string' ? value.split(',') : value });
  };

  const handleCurrencyChange = (event: any) => {
    onFiltersChange({ ...filters, currency: event.target.value });
  };

  const handlePeriodChange = (event: any) => {
    onFiltersChange({ ...filters, period: event.target.value });
  };

  return (
    <Paper elevation={1} sx={{ p: spacing.lg, mb: spacing.lg, borderRadius: '8px' }}>
      <Typography variant="h6" sx={{ mb: spacing.md, color: colors.text }}>
        Filters
      </Typography>
      
      <LocalizationProvider dateAdapter={AdapterDateFns}>
        <Box sx={{ 
          display: 'grid', 
          gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr', md: 'repeat(auto-fit, minmax(200px, 1fr))' },
          gap: spacing.md,
          alignItems: 'start'
        }}>
          {/* Date Range */}
          <DatePicker
            label="Date From"
            value={filters.dateFrom}
            onChange={handleDateFromChange}
            slotProps={{
              textField: {
                size: 'small',
                fullWidth: true,
              },
            }}
          />
          
          <DatePicker
            label="Date To"
            value={filters.dateTo}
            onChange={handleDateToChange}
            slotProps={{
              textField: {
                size: 'small',
                fullWidth: true,
              },
            }}
          />

          {/* Store Selection */}
          <FormControl size="small" fullWidth>
            <InputLabel id="stores-label">Stores</InputLabel>
            <Select
              labelId="stores-label"
              id="stores-select"
              multiple
              value={filters.stores}
              onChange={handleStoresChange}
              input={<OutlinedInput id="select-stores" label="Stores" />}
              renderValue={(selected) => (
                <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
                  {selected.map((value) => (
                    <Chip 
                      key={value} 
                      label={availableStores.find(store => store.id === value)?.name || value}
                      size="small" 
                    />
                  ))}
                </Box>
              )}
              MenuProps={MenuProps}
            >
              {availableStores.map((store) => (
                <MenuItem key={store.id} value={store.id}>
                  {store.name}
                </MenuItem>
              ))}
            </Select>
          </FormControl>

          {/* Channel Selection */}
          <FormControl size="small" fullWidth>
            <InputLabel id="channels-label">Channels</InputLabel>
            <Select
              labelId="channels-label"
              id="channels-select"
              multiple
              value={filters.channels}
              onChange={handleChannelsChange}
              input={<OutlinedInput id="select-channels" label="Channels" />}
              renderValue={(selected) => (
                <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
                  {selected.map((value) => (
                    <Chip 
                      key={value} 
                      label={availableChannels.find(channel => channel.id === value)?.name || value}
                      size="small" 
                    />
                  ))}
                </Box>
              )}
              MenuProps={MenuProps}
            >
              {availableChannels.map((channel) => (
                <MenuItem key={channel.id} value={channel.id}>
                  {channel.name}
                </MenuItem>
              ))}
            </Select>
          </FormControl>

          {/* Currency Toggle */}
          {showCurrencyToggle && (
            <FormControl size="small" fullWidth>
              <InputLabel id="currency-label">Currency</InputLabel>
              <Select
                labelId="currency-label"
                id="currency-select"
                value={filters.currency || 'IDR'}
                onChange={handleCurrencyChange}
                label="Currency"
              >
                <MenuItem value="IDR">IDR</MenuItem>
                <MenuItem value="USD">USD</MenuItem>
                <MenuItem value="SGD">SGD</MenuItem>
              </Select>
            </FormControl>
          )}

          {/* Period Toggle */}
          {showPeriodToggle && (
            <FormControl size="small" fullWidth>
              <InputLabel id="period-label">Period</InputLabel>
              <Select
                labelId="period-label"
                id="period-select"
                value={filters.period || 'daily'}
                onChange={handlePeriodChange}
                label="Period"
              >
                <MenuItem value="daily">Daily</MenuItem>
                <MenuItem value="weekly">Weekly</MenuItem>
                <MenuItem value="monthly">Monthly</MenuItem>
                <MenuItem value="quarterly">Quarterly</MenuItem>
              </Select>
            </FormControl>
          )}
        </Box>
      </LocalizationProvider>
    </Paper>
  );
}