import { Backdrop, CircularProgress } from "@mui/material";
import useLoading from "../useLoading";

export default function LoadingOverlay() {
  const loading = useLoading();
  return (
    <Backdrop
      open={loading}
      sx={{ color: "#fff", zIndex: (theme) => theme.zIndex.tooltip + 1 }}
    >
      <CircularProgress color="inherit" />
    </Backdrop>
  );
}
