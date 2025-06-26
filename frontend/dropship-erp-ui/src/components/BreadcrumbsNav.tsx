import { Breadcrumbs, Typography } from "@mui/material";
import { Link, useLocation } from "react-router-dom";

export default function BreadcrumbsNav() {
  const location = useLocation();
  const pathnames = location.pathname.split("/").filter(Boolean);
  return (
    <Breadcrumbs aria-label="breadcrumb" sx={{ mb: 2 }}>
      <Link to="/">Home</Link>
      {pathnames.map((value, index) => {
        const to = `/${pathnames.slice(0, index + 1).join("/")}`;
        const isLast = index === pathnames.length - 1;
        return isLast ? (
          <Typography color="text.primary" key={to}>
            {decodeURIComponent(value)}
          </Typography>
        ) : (
          <Link key={to} to={to}>
            {decodeURIComponent(value)}
          </Link>
        );
      })}
    </Breadcrumbs>
  );
}
