// File: src/components/DropshipImport.test.tsx

import "@testing-library/jest-dom";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import * as api from "../api";
import DropshipImport from "./DropshipImport";

// Mock the api module so importDropship calls do not hit the network
jest.mock("../api", () => ({
  importDropship: jest.fn(),
}));

describe("DropshipImport", () => {
  it("renders input and button", () => {
    render(<DropshipImport />);
    expect(screen.getByLabelText(/CSV file/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Import/i })).toBeInTheDocument();
  });

  it("shows success message on successful import", async () => {
    // Arrange: mock implementation of importDropship to resolve
    (api.importDropship as jest.Mock).mockResolvedValue({
      data: { inserted: 3 },
    });

    render(<DropshipImport />);
    const file = new File(["data"], "data.csv", { type: "text/csv" });
    fireEvent.change(screen.getByLabelText(/CSV file/i), {
      target: { files: [file] },
    });
    fireEvent.click(screen.getByRole("button", { name: /Import/i }));

    await waitFor(() =>
      expect(
        screen.getByText(/Imported 3 rows successfully!/i),
      ).toBeInTheDocument(),
    );
  });

  it("shows error message on failure", async () => {
    (api.importDropship as jest.Mock).mockRejectedValue(new Error("boom"));

    render(<DropshipImport />);
    const file = new File(["bad"], "bad.csv", { type: "text/csv" });
    fireEvent.change(screen.getByLabelText(/CSV file/i), {
      target: { files: [file] },
    });
    fireEvent.click(screen.getByRole("button", { name: /Import/i }));

    await waitFor(() => expect(screen.getByText(/boom/i)).toBeInTheDocument());
  });
});
