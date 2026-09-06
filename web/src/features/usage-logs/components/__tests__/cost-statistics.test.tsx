import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import i18next from 'i18next'
import { beforeAll, beforeEach, describe, expect, test, vi } from 'vitest'

import { getCostStatistics, getCostStatisticsLogs } from '../../api'
import { CostStatisticsPage } from '../cost-statistics'

vi.mock('../../api', () => ({
  exportCostStatistics: vi.fn(),
  getCostStatistics: vi.fn(),
  getCostStatisticsLogs: vi.fn(),
}))

const mockedGetCostStatistics = vi.mocked(getCostStatistics)
const mockedGetCostStatisticsLogs = vi.mocked(getCostStatisticsLogs)

function renderPage() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })

  return render(
    <QueryClientProvider client={queryClient}>
      <CostStatisticsPage />
    </QueryClientProvider>
  )
}

function statsResponse() {
  return {
    success: true,
    data: {
      rows: [
        {
          model_name: 'gpt-4o',
          channel_id: 12,
          currency: 'USD',
          request_count: 1,
          success_count: 1,
          failure_count: 0,
          prompt_tokens: 10,
          completion_tokens: 20,
          cached_tokens: 0,
          cache_creation_tokens: 0,
          image_tokens: 0,
          audio_tokens: 0,
          audio_input_tokens: 0,
          audio_output_tokens: 0,
          image_count: 0,
          audio_seconds: 0,
          audio_output_seconds: 0,
          video_seconds: 0,
          total_tokens: 30,
          actual_cost_micros: 100,
          revenue_micros: 200,
          gross_profit_micros: 100,
          profit_rate: 0.5,
          usage_missing_count: 0,
        },
      ],
      currency: 'USD',
      mixed_currency: false,
      request_count: 1,
      success_count: 1,
      failure_count: 0,
      prompt_tokens: 10,
      completion_tokens: 20,
      cached_tokens: 0,
      audio_tokens: 0,
      audio_input_tokens: 0,
      audio_output_tokens: 0,
      image_count: 0,
      audio_seconds: 0,
      video_seconds: 0,
      total_tokens: 30,
      actual_cost_micros: 100,
      revenue_micros: 200,
      gross_profit_micros: 100,
      profit_rate: 0.5,
    },
  }
}

beforeAll(() => {
  i18next.addResourceBundle('en', 'translation', {
    Available: 'Available',
    Channel: 'Channel',
    Details: 'Details',
    Error: 'Error',
    Failures: 'Failures',
    'Failed to load logs': 'Failed to load logs',
    'Gross profit': 'Gross profit',
    'Profit rate': 'Profit rate',
    Loading: 'Loading',
    Model: 'Model',
    'No Logs Found': 'No Logs Found',
    Next: 'Next',
    Previous: 'Previous',
    Requests: 'Requests',
    Status: 'Status',
    Success: 'Success',
    Time: 'Time',
    Total: 'Total',
    'Total tokens': 'Total tokens',
    Usage: 'Usage',
    Username: 'Username',
    'Upstream cost': 'Upstream cost',
    'User revenue': 'User revenue',
  })
})

beforeEach(() => {
  mockedGetCostStatistics.mockResolvedValue(statsResponse())
  mockedGetCostStatisticsLogs.mockResolvedValue({
    success: true,
    data: {
      items: [
        {
          id: 101,
          user_id: 7,
          created_at: 1_700_000_000,
          type: 2,
          model_name: 'gpt-4o',
          username: 'alice',
          channel_id: 12,
          group: 'default',
          request_id: 'request-safe-id',
          upstream_request_id: 'upstream-safe-id',
          status: 'ok',
          success: true,
          quota: 200,
          prompt_tokens: 10,
          completion_tokens: 20,
          cached_tokens: 0,
          audio_input_tokens: 0,
          audio_output_tokens: 0,
          image_count: 0,
          audio_seconds: 0,
          audio_output_seconds: 0,
          video_seconds: 0,
          total_tokens: 30,
          actual_cost_micros: 100,
          revenue_micros: 200,
          gross_profit_micros: 100,
          profit_rate: 0.5,
          currency: 'USD',
          usage_available: true,
          billing_event: 'settled',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    },
  })
})

describe('cost statistics drill-down', () => {
  test('expands a summary row, fetches its scoped logs, and keeps the action accessible', async () => {
    const user = userEvent.setup()
    renderPage()

    const expandButton = await screen.findByRole('button', { name: 'Details' })
    expect(expandButton).toHaveAttribute('aria-expanded', 'false')

    await user.click(expandButton)

    await waitFor(() => {
      expect(mockedGetCostStatisticsLogs).toHaveBeenCalledWith({
        model_name: 'gpt-4o',
        channel: 12,
        username: undefined,
        group: undefined,
        start_timestamp: undefined,
        end_timestamp: undefined,
        page: 1,
        page_size: 20,
      })
    })
    expect(expandButton).toHaveAttribute('aria-expanded', 'true')
    const detailRegion = await screen.findByRole('region', { name: 'Details' })
    expect(detailRegion.closest('td')).toHaveAttribute('colspan', '18')

    const detailTable = within(detailRegion).getByRole('table')
    expect(
      within(detailTable).getByRole('columnheader', { name: 'Gross profit' })
    ).toBeInTheDocument()
    expect(
      within(detailTable).getByRole('columnheader', { name: 'Profit rate' })
    ).toBeInTheDocument()

    const detailRow = within(detailTable).getByText('alice').closest('tr')
    if (!detailRow) throw new Error('Expected a request detail row')
    const detailCells = within(detailRow).getAllByRole('cell')
    expect(detailCells[6]).toHaveTextContent('USD 0.000100')
    expect(detailCells[7]).toHaveTextContent('50.00%')
    expect(within(detailTable).getByText('Available')).toBeInTheDocument()
    expect(screen.queryByText('secret prompt content')).toBeNull()
  })

  test('keeps the summary table header and empty row aligned to 18 columns', async () => {
    const response = statsResponse()
    response.data.rows = []
    mockedGetCostStatistics.mockResolvedValue(response)

    renderPage()

    const emptyCell = await screen.findByText('No cost data found')
    const summaryTable = emptyCell.closest('table')
    if (!summaryTable) throw new Error('Expected the summary table')
    expect(within(summaryTable).getAllByRole('columnheader')).toHaveLength(18)
    expect(emptyCell).toHaveAttribute('colspan', '18')
  })
})
