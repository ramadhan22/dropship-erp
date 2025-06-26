import "@testing-library/jest-dom";
import { render } from "@testing-library/react";
import { fireEvent, screen, waitFor } from "@testing-library/dom";
import * as api from "../api";
import ChannelPage from "./ChannelPage";

jest.mock("../api", () => ({
  createJenisChannel: jest.fn(),
  createStore: jest.fn(),
  updateStore: jest.fn(),
  deleteStore: jest.fn(),
  listJenisChannels: jest.fn(),
  listAllStoresDirect: jest.fn(),
  fetchShopeeAuthURL: jest.fn(),
}));

it("creates channel and store", async () => {
  (api.listJenisChannels as jest.Mock).mockResolvedValue({
    data: [{ jenis_channel_id: 1, jenis_channel: "Tokopedia" }],
  });
  (api.listAllStoresDirect as jest.Mock).mockResolvedValue({ data: [] });
  (api.fetchShopeeAuthURL as jest.Mock).mockResolvedValue({ data: { url: "u" } });
  (api.createJenisChannel as jest.Mock).mockResolvedValue({
    data: { jenis_channel_id: 2 },
  });
  (api.createStore as jest.Mock).mockResolvedValue({ data: { store_id: 3 } });

  render(<ChannelPage />);

  await waitFor(() => expect(api.listJenisChannels).toHaveBeenCalled());

  fireEvent.click(screen.getByRole("button", { name: /New Channel/i }));
  fireEvent.change(screen.getByLabelText(/Channel Name/i), {
    target: { value: "Toko" },
  });
  fireEvent.click(screen.getByRole("button", { name: /^Save$/i }));
  await waitFor(() => expect(api.createJenisChannel).toHaveBeenCalledWith("Toko"));
  await waitFor(() => expect(screen.queryByText(/Add Channel/i)).not.toBeInTheDocument());

  fireEvent.click(screen.getAllByRole("button", { name: /New Store/i })[0]);
  fireEvent.mouseDown(screen.getByLabelText(/Channel/));
  fireEvent.click(await screen.findByText("Tokopedia"));
  fireEvent.change(screen.getByLabelText(/Store Name/i), {
    target: { value: "ShopA" },
  });
  fireEvent.click(screen.getByRole("button", { name: /^Save$/i }));
  await waitFor(() => expect(api.createStore).toHaveBeenCalledWith(1, "ShopA"));
});
