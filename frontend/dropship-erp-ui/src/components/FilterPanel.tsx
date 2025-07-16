import { useState, useCallback } from "react";
import {
  Box,
  Accordion,
  AccordionSummary,
  AccordionDetails,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Button,
  Chip,
  Typography,
  IconButton,
  Paper,
} from "@mui/material";
import {
  ExpandMore as ExpandMoreIcon,
  Add as AddIcon,
  Delete as DeleteIcon,
  Clear as ClearIcon,
  FilterList as FilterListIcon,
} from "@mui/icons-material";
import { DatePicker } from "@mui/x-date-pickers/DatePicker";
import { LocalizationProvider } from "@mui/x-date-pickers";
import { AdapterDateFns } from "@mui/x-date-pickers/AdapterDateFns";
import type { FilterCondition, FilterParams, SortCondition } from "../useServerPagination";

export interface FilterField {
  key: string;
  label: string;
  type: "text" | "number" | "date" | "select" | "multiselect";
  options?: { value: string; label: string }[];
  operators?: string[];
}

export interface FilterPanelProps {
  fields: FilterField[];
  onFiltersChange: (filters: FilterParams | undefined) => void;
  onSortChange: (sort: SortCondition[]) => void;
  currentFilters?: FilterParams;
  currentSort?: SortCondition[];
  loading?: boolean;
}

const OPERATORS = {
  text: [
    { value: "contains", label: "Contains" },
    { value: "eq", label: "Equals" },
    { value: "starts_with", label: "Starts with" },
    { value: "ends_with", label: "Ends with" },
  ],
  number: [
    { value: "eq", label: "Equals" },
    { value: "neq", label: "Not equals" },
    { value: "gt", label: "Greater than" },
    { value: "gte", label: "Greater than or equal" },
    { value: "lt", label: "Less than" },
    { value: "lte", label: "Less than or equal" },
    { value: "between", label: "Between" },
  ],
  date: [
    { value: "eq", label: "On date" },
    { value: "gt", label: "After" },
    { value: "gte", label: "On or after" },
    { value: "lt", label: "Before" },
    { value: "lte", label: "On or before" },
    { value: "between", label: "Between" },
  ],
  select: [
    { value: "eq", label: "Equals" },
    { value: "neq", label: "Not equals" },
    { value: "in", label: "In list" },
    { value: "not_in", label: "Not in list" },
  ],
  multiselect: [
    { value: "in", label: "In list" },
    { value: "not_in", label: "Not in list" },
  ],
} as const;

export default function FilterPanel({
  fields,
  onFiltersChange,
  onSortChange,
  currentFilters,
  currentSort = [],
  loading = false,
}: FilterPanelProps) {
  const [expanded, setExpanded] = useState(false);
  const [conditions, setConditions] = useState<FilterCondition[]>(
    currentFilters?.filters?.conditions || []
  );
  const [logic, setLogic] = useState<"AND" | "OR">(
    (currentFilters?.filters?.logic as "AND" | "OR") || "AND"
  );
  const [sortConditions, setSortConditions] = useState<SortCondition[]>(currentSort);

  const addCondition = useCallback(() => {
    const newCondition: FilterCondition = {
      field: fields[0]?.key || "",
      operator: "contains",
      value: "",
    };
    setConditions((prev) => [...prev, newCondition]);
  }, [fields]);

  const updateCondition = useCallback((index: number, updates: Partial<FilterCondition>) => {
    setConditions((prev) =>
      prev.map((condition, i) => (i === index ? { ...condition, ...updates } : condition))
    );
  }, []);

  const removeCondition = useCallback((index: number) => {
    setConditions((prev) => prev.filter((_, i) => i !== index));
  }, []);

  const addSort = useCallback(() => {
    const newSort: SortCondition = {
      field: fields[0]?.key || "",
      direction: "asc",
    };
    setSortConditions((prev) => [...prev, newSort]);
  }, [fields]);

  const updateSort = useCallback((index: number, updates: Partial<SortCondition>) => {
    setSortConditions((prev) =>
      prev.map((sort, i) => (i === index ? { ...sort, ...updates } : sort))
    );
  }, []);

  const removeSort = useCallback((index: number) => {
    setSortConditions((prev) => prev.filter((_, i) => i !== index));
  }, []);

  const applyFilters = useCallback(() => {
    const validConditions = conditions.filter(
      (c) => c.field && c.operator && (c.value !== "" || c.operator === "is_null" || c.operator === "is_not_null")
    );

    const filterParams: FilterParams | undefined = validConditions.length > 0 ? {
      filters: {
        logic,
        conditions: validConditions,
      },
    } : undefined;

    onFiltersChange(filterParams);
    onSortChange(sortConditions);
  }, [conditions, logic, sortConditions, onFiltersChange, onSortChange]);

  const clearFilters = useCallback(() => {
    setConditions([]);
    setSortConditions([]);
    onFiltersChange(undefined);
    onSortChange([]);
  }, [onFiltersChange, onSortChange]);

  const getFieldOperators = (field: FilterField) => {
    if (field.operators) {
      return field.operators.map(op => ({ value: op, label: op }));
    }
    return OPERATORS[field.type] || OPERATORS.text;
  };

  const renderValueInput = (condition: FilterCondition, index: number) => {
    const field = fields.find((f) => f.key === condition.field);
    if (!field) return null;

    if (condition.operator === "is_null" || condition.operator === "is_not_null") {
      return null; // No value input needed
    }

    if (condition.operator === "between") {
      if (field.type === "date") {
        return (
          <LocalizationProvider dateAdapter={AdapterDateFns}>
            <Box sx={{ display: "flex", gap: 1 }}>
              <DatePicker
                label="From"
                value={condition.values?.[0] ? new Date(condition.values[0]) : null}
                onChange={(date) =>
                  updateCondition(index, {
                    values: [date?.toISOString().split('T')[0] || "", condition.values?.[1] || ""],
                  })
                }
                slotProps={{
                  textField: {
                    size: "small"
                  }
                }}
              />
              <DatePicker
                label="To"
                value={condition.values?.[1] ? new Date(condition.values[1]) : null}
                onChange={(date) =>
                  updateCondition(index, {
                    values: [condition.values?.[0] || "", date?.toISOString().split('T')[0] || ""],
                  })
                }
                slotProps={{
                  textField: {
                    size: "small"
                  }
                }}
              />
            </Box>
          </LocalizationProvider>
        );
      }
      return (
        <Box sx={{ display: "flex", gap: 1 }}>
          <TextField
            size="small"
            type="number"
            label="From"
            value={condition.values?.[0] || ""}
            onChange={(e) =>
              updateCondition(index, {
                values: [e.target.value, condition.values?.[1] || ""],
              })
            }
          />
          <TextField
            size="small"
            type="number"
            label="To"
            value={condition.values?.[1] || ""}
            onChange={(e) =>
              updateCondition(index, {
                values: [condition.values?.[0] || "", e.target.value],
              })
            }
          />
        </Box>
      );
    }

    if (field.type === "date") {
      return (
        <LocalizationProvider dateAdapter={AdapterDateFns}>
          <DatePicker
            label="Value"
            value={condition.value ? new Date(condition.value) : null}
            onChange={(date) =>
              updateCondition(index, { value: date?.toISOString().split('T')[0] || "" })
            }
            slotProps={{
              textField: {
                size: "small"
              }
            }}
          />
        </LocalizationProvider>
      );
    }

    if (field.type === "select" && field.options) {
      if (condition.operator === "in" || condition.operator === "not_in") {
        return (
          <FormControl size="small" sx={{ minWidth: 200 }}>
            <InputLabel>Values</InputLabel>
            <Select
              multiple
              value={condition.values || []}
              onChange={(e) =>
                updateCondition(index, { values: Array.isArray(e.target.value) ? e.target.value : [e.target.value] })
              }
              label="Values"
              renderValue={(selected) => (
                <Box sx={{ display: "flex", flexWrap: "wrap", gap: 0.5 }}>
                  {(selected as string[]).map((value) => (
                    <Chip key={value} label={field.options?.find(o => o.value === value)?.label || value} size="small" />
                  ))}
                </Box>
              )}
            >
              {field.options.map((option) => (
                <MenuItem key={option.value} value={option.value}>
                  {option.label}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
        );
      }
      return (
        <FormControl size="small" sx={{ minWidth: 150 }}>
          <InputLabel>Value</InputLabel>
          <Select
            value={condition.value || ""}
            onChange={(e) => updateCondition(index, { value: e.target.value })}
            label="Value"
          >
            {field.options.map((option) => (
              <MenuItem key={option.value} value={option.value}>
                {option.label}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
      );
    }

    return (
      <TextField
        size="small"
        label="Value"
        type={field.type === "number" ? "number" : "text"}
        value={condition.value || ""}
        onChange={(e) => updateCondition(index, { value: e.target.value })}
      />
    );
  };

  const activeFiltersCount = conditions.filter(c => c.field && c.operator && (c.value !== "" || c.operator === "is_null" || c.operator === "is_not_null")).length;

  return (
    <Paper elevation={1} sx={{ mb: 2 }}>
      <Accordion expanded={expanded} onChange={(_, isExpanded) => setExpanded(isExpanded)}>
        <AccordionSummary expandIcon={<ExpandMoreIcon />}>
          <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
            <FilterListIcon />
            <Typography variant="h6">Filters & Sorting</Typography>
            {activeFiltersCount > 0 && (
              <Chip size="small" label={`${activeFiltersCount} active`} color="primary" />
            )}
            {sortConditions.length > 0 && (
              <Chip size="small" label={`${sortConditions.length} sorts`} color="secondary" />
            )}
          </Box>
        </AccordionSummary>
        <AccordionDetails>
          <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
            {/* Filter Conditions */}
            <Box>
              <Box sx={{ display: "flex", justifyContent: "between", alignItems: "center", mb: 2 }}>
                <Typography variant="subtitle1">Filter Conditions</Typography>
                <FormControl size="small" sx={{ minWidth: 100 }}>
                  <InputLabel>Logic</InputLabel>
                  <Select
                    value={logic}
                    onChange={(e) => setLogic(e.target.value as "AND" | "OR")}
                    label="Logic"
                  >
                    <MenuItem value="AND">AND</MenuItem>
                    <MenuItem value="OR">OR</MenuItem>
                  </Select>
                </FormControl>
              </Box>

              {conditions.map((condition, index) => (
                <Box key={index} sx={{ mb: 2, display: "flex", gap: 2, alignItems: "center", flexWrap: "wrap" }}>
                  <FormControl size="small" sx={{ minWidth: 150 }}>
                    <InputLabel>Field</InputLabel>
                    <Select
                      value={condition.field}
                      onChange={(e) => updateCondition(index, { field: e.target.value, operator: "contains", value: "" })}
                      label="Field"
                    >
                      {fields.map((field) => (
                        <MenuItem key={field.key} value={field.key}>
                          {field.label}
                        </MenuItem>
                      ))}
                    </Select>
                  </FormControl>
                  <FormControl size="small" sx={{ minWidth: 120 }}>
                    <InputLabel>Operator</InputLabel>
                    <Select
                      value={condition.operator}
                      onChange={(e) => updateCondition(index, { operator: e.target.value, value: "", values: [] })}
                      label="Operator"
                    >
                      {getFieldOperators(fields.find(f => f.key === condition.field) || fields[0]).map((op) => (
                        <MenuItem key={op.value} value={op.value}>
                          {op.label}
                        </MenuItem>
                      ))}
                    </Select>
                  </FormControl>
                  <Box sx={{ minWidth: 200 }}>
                    {renderValueInput(condition, index)}
                  </Box>
                  <IconButton onClick={() => removeCondition(index)} color="error">
                    <DeleteIcon />
                  </IconButton>
                </Box>
              ))}

              <Button startIcon={<AddIcon />} onClick={addCondition} variant="outlined">
                Add Filter
              </Button>
            </Box>

            {/* Sort Conditions */}
            <Box>
              <Typography variant="subtitle1" sx={{ mb: 2 }}>
                Sort Conditions
              </Typography>

              {sortConditions.map((sort, index) => (
                <Box key={index} sx={{ mb: 2, display: "flex", gap: 2, alignItems: "center" }}>
                  <FormControl size="small" sx={{ minWidth: 150 }}>
                    <InputLabel>Field</InputLabel>
                    <Select
                      value={sort.field}
                      onChange={(e) => updateSort(index, { field: e.target.value })}
                      label="Field"
                    >
                      {fields.map((field) => (
                        <MenuItem key={field.key} value={field.key}>
                          {field.label}
                        </MenuItem>
                      ))}
                    </Select>
                  </FormControl>
                  <FormControl size="small" sx={{ minWidth: 120 }}>
                    <InputLabel>Direction</InputLabel>
                    <Select
                      value={sort.direction}
                      onChange={(e) => updateSort(index, { direction: e.target.value as "asc" | "desc" })}
                      label="Direction"
                    >
                      <MenuItem value="asc">Ascending</MenuItem>
                      <MenuItem value="desc">Descending</MenuItem>
                    </Select>
                  </FormControl>
                  <IconButton onClick={() => removeSort(index)} color="error">
                    <DeleteIcon />
                  </IconButton>
                </Box>
              ))}

              <Button startIcon={<AddIcon />} onClick={addSort} variant="outlined">
                Add Sort
              </Button>
            </Box>

            {/* Action Buttons */}
            <Box sx={{ display: "flex", gap: 2, pt: 2, borderTop: 1, borderColor: "divider" }}>
              <Button 
                variant="contained" 
                onClick={applyFilters} 
                disabled={loading}
                startIcon={<FilterListIcon />}
              >
                Apply Filters
              </Button>
              <Button 
                variant="outlined" 
                onClick={clearFilters} 
                disabled={loading}
                startIcon={<ClearIcon />}
              >
                Clear All
              </Button>
            </Box>
          </Box>
        </AccordionDetails>
      </Accordion>
    </Paper>
  );
}