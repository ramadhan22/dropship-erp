import "@testing-library/jest-dom";
import { render } from "@testing-library/react";
import { fireEvent, screen, waitFor } from "@testing-library/dom";
import * as api from "../api";
import ChannelPage from "./ChannelPage";

jest.mock("../api", () => ({
  createJenisChannel: jest.fn(),
  createStore: jest.fn(),
  listJenisChannels: jest.fn(),
  listStores: jest.fn(),
}));

it("creates channel and store", async () => {
  (api.listJenisChannels as jest.Mock).mockResolvedValue({
    data: [{ jenis_channel_id: 1, jenis_channel: "Tokopedia" }],
  });
  (api.listStores as jest.Mock).mockResolvedValue({ data: [] });

  render(<ChannelPage />);

  await waitFor(() => expect(api.listJenisChannels).toHaveBeenCalled());

  fireEvent.change(screen.getByLabelText(/New Channel/i), {
    target: { value: "Toko" },
  });
  fireEvent.click(screen.getByRole("button", { name: /Add Channel/i }));
  await waitFor(() =>
    expect(api.createJenisChannel).toHaveBeenCalledWith("Toko")
  );

  fireEvent.change(screen.getByLabelText(/Channel Select/i), {
    target: { value: "1" },
  });
  await waitFor(() => expect(api.listStores).toHaveBeenCalledWith(1));

  fireEvent.change(screen.getByLabelText(/Store Name/i), {
    target: { value: "ShopA" },
  });
  fireEvent.click(screen.getByRole("button", { name: /Add Store/i }));
  await waitFor(() =>
    expect(api.createStore).toHaveBeenCalledWith(1, "ShopA")
  );
});
