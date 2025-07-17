import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { jest } from '@jest/globals';
import AdsPerformancePage from '../components/AdsPerformancePage';

// Mock the ads performance API
jest.mock('../api/adsPerformance', () => ({
  fetchAdsCampaigns: jest.fn(),
  fetchAdsPerformanceSummary: jest.fn(),
  syncHistoricalAdsPerformance: jest.fn(),
}));

// Mock fetch for stores endpoint
const mockFetch = jest.fn();
global.fetch = mockFetch;

describe('AdsPerformancePage Store Filter', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('should populate store dropdown with data from /api/stores/all', async () => {
    // Mock the stores API response
    mockFetch.mockImplementation((url) => {
      if (url === '/api/stores/all') {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve([
            { store_id: 1, nama_toko: 'MR eStore Shopee' },
            { store_id: 2, nama_toko: 'MR Barista Gear' }
          ])
        });
      }
      // Mock other API calls
      return Promise.resolve({
        ok: true,
        json: () => Promise.resolve({ campaigns: [], totalCampaigns: 0 })
      });
    });

    render(<AdsPerformancePage />);

    // Wait for initial data loading
    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith('/api/stores/all');
    });

    // Click on "Sync Historical Data" button to open dialog
    const syncButton = screen.getByText('Sync Historical Data');
    fireEvent.click(syncButton);

    // Wait for dialog to open
    await waitFor(() => {
      expect(screen.getByText('Sync Historical Ads Performance')).toBeInTheDocument();
    });

    // Check if store dropdown is populated
    const storeSelect = screen.getByLabelText('Store');
    expect(storeSelect).toBeInTheDocument();

    // The dropdown should not be empty - we should be able to find store options
    // This is where the bug would be manifested if the stores are not populated
    console.log('Store select element:', storeSelect);
  });

  test('should handle empty stores response', async () => {
    // Mock empty stores response
    mockFetch.mockImplementation((url) => {
      if (url === '/api/stores/all') {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve([]) // Empty array
        });
      }
      return Promise.resolve({
        ok: true,
        json: () => Promise.resolve({ campaigns: [], totalCampaigns: 0 })
      });
    });

    render(<AdsPerformancePage />);

    // Wait for initial data loading
    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith('/api/stores/all');
    });

    // Click on "Sync Historical Data" button to open dialog
    const syncButton = screen.getByText('Sync Historical Data');
    fireEvent.click(syncButton);

    // Wait for dialog to open
    await waitFor(() => {
      expect(screen.getByText('Sync Historical Ads Performance')).toBeInTheDocument();
    });

    // The dropdown should be empty and the submit button should be disabled
    const submitButton = screen.getByText('Start Sync');
    expect(submitButton).toBeDisabled();
  });

  test('should handle stores API error', async () => {
    // Mock stores API error
    mockFetch.mockImplementation((url) => {
      if (url === '/api/stores/all') {
        return Promise.resolve({
          ok: false,
          status: 500
        });
      }
      return Promise.resolve({
        ok: true,
        json: () => Promise.resolve({ campaigns: [], totalCampaigns: 0 })
      });
    });

    // Spy on console.error to check if error is logged
    const consoleSpy = jest.spyOn(console, 'error').mockImplementation(() => {});

    render(<AdsPerformancePage />);

    // Wait for error to be logged
    await waitFor(() => {
      expect(consoleSpy).toHaveBeenCalledWith('Error fetching stores:', expect.any(Error));
    });

    consoleSpy.mockRestore();
  });
});