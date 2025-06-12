// File: src/components/MetricsPage.test.tsx

import "@testing-library/jest-dom";
import { render } from "@testing-library/react";
import { fireEvent, screen, waitFor } from "@testing-library/dom";
import * as api from "../api";
import MetricsPage from "./MetricsPage";

// Mock the API module so we don’t make real HTTP calls
jest.mock("../api", () => ({
  computeMetrics: jest.fn(),
  fetchMetrics: jest.fn(),
}));

describe("MetricsPage Component", () => {
  it("computes then fetches metrics successfully", async () => {
    // Arrange: stub out the API calls
    (api.computeMetrics as jest.Mock).mockResolvedValue({} as any);
    (api.fetchMetrics as jest.Mock).mockResolvedValue({
      data: { net_profit: 10 },
    } as any);

    // Render the component
    render(<MetricsPage />);

    // Fill in form fields
    fireEvent.change(screen.getByLabelText(/Shop/i), {
      target: { value: "ShopX" },
    });
    fireEvent.change(screen.getByLabelText(/Period/i), {
      target: { value: "2025-05" },
    });

    // Act: click Compute and assert success message
    fireEvent.click(screen.getByRole("button", { name: /Compute/i }));
    await waitFor(() =>
      expect(screen.getByText(/Metrics computed!/i)).toBeInTheDocument()
    );

    // Act: click Fetch and assert the fetched value is shown
    fireEvent.click(screen.getByRole("button", { name: /Fetch/i }));
    await waitFor(() => expect(screen.getByText(/10/i)).toBeInTheDocument());
  });
});
