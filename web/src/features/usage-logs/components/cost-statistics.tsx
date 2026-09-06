import { useQuery } from '@tanstack/react-query'
import { ChevronDown, ChevronRight, Loader2 } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { formatTimestamp } from '@/lib/format'

import {
  exportCostStatistics,
  getCostStatistics,
  getCostStatisticsLogs,
} from '../api'
import type { UsageLog } from '../data/schema'
import type {
  CostStatistics,
  CostStatisticsLog,
  CostStatisticsLogsParams,
} from '../types'
import { DetailsDialog } from './dialogs/details-dialog'

const emptyStats: CostStatistics = {
  rows: [],
  currency: 'USD',
  mixed_currency: false,
  request_count: 0,
  success_count: 0,
  failure_count: 0,
  prompt_tokens: 0,
  completion_tokens: 0,
  cached_tokens: 0,
  audio_tokens: 0,
  audio_input_tokens: 0,
  audio_output_tokens: 0,
  image_count: 0,
  audio_seconds: 0,
  video_seconds: 0,
  total_tokens: 0,
  actual_cost_micros: 0,
  revenue_micros: 0,
  gross_profit_micros: 0,
  profit_rate: 0,
}

const detailPageSize = 20
const mainTableColumnCount = 18

type ReportParams = Record<string, string | number | undefined>

function toTimestamp(value: string, end: boolean) {
  if (!value) return undefined
  const date = new Date(`${value}T${end ? '23:59:59' : '00:00:00'}`)
  return Math.floor(date.getTime() / 1000)
}

function money(micros: number, currency: string) {
  return `${currency || 'USD'} ${(micros / 1_000_000).toFixed(6)}`
}

function summaryMoney(
  micros: number,
  currency: string,
  mixedCurrency: boolean,
  mixedLabel: string,
) {
  return mixedCurrency ? mixedLabel : money(micros, currency)
}

function rowKey(row: CostStatistics['rows'][number]) {
  return `${row.model_name}:${row.channel_id}:${row.currency}`
}

/**
 * Adapt the sanitized cost-log projection to the existing log details dialog.
 * Do not pass through content, token names, IPs, or the raw `other` payload.
 */
function toSafeUsageLog(log: CostStatisticsLog): UsageLog {
  return {
    id: log.id,
    user_id: log.user_id,
    created_at: log.created_at,
    type: log.type,
    content: '',
    username: log.username ?? '',
    token_name: '',
    model_name: log.model_name,
    quota: log.quota,
    prompt_tokens: log.prompt_tokens,
    completion_tokens: log.completion_tokens,
    use_time: 0,
    is_stream: false,
    channel: log.channel_id,
    channel_name: '',
    token_id: 0,
    group: log.group ?? '',
    ip: '',
    other: '',
    request_id: log.request_id ?? '',
    upstream_request_id: log.upstream_request_id ?? '',
  }
}

function CostStatisticsRowDetails(props: {
  row: CostStatistics['rows'][number]
  filters: ReportParams
  expanded: boolean
}) {
  const { t } = useTranslation()
  const [page, setPage] = useState(1)
  const [selectedLog, setSelectedLog] = useState<CostStatisticsLog | null>(null)
  const params: CostStatisticsLogsParams = {
    model_name: props.row.model_name,
    channel: props.row.channel_id || undefined,
    username:
      typeof props.filters.username === 'string'
        ? props.filters.username
        : undefined,
    group:
      typeof props.filters.group === 'string' ? props.filters.group : undefined,
    start_timestamp:
      typeof props.filters.start_timestamp === 'number'
        ? props.filters.start_timestamp
        : undefined,
    end_timestamp:
      typeof props.filters.end_timestamp === 'number'
        ? props.filters.end_timestamp
        : undefined,
    page,
    page_size: detailPageSize,
  }

  const { data, isFetching, isError } = useQuery({
    queryKey: ['cost-statistics-logs', params],
    queryFn: async () => {
      const result = await getCostStatisticsLogs(params)
      if (!result.success) {
        throw new Error(result.message || 'Failed to load logs')
      }
      return (
        result.data ?? {
          items: [],
          total: 0,
          page,
          page_size: detailPageSize,
        }
      )
    },
    enabled: props.expanded,
    placeholderData: (previousData) => previousData,
  })

  if (!props.expanded) return null

  const items = data?.items ?? []
  const total = data?.total ?? 0
  const detailRegionId = `cost-statistics-details-${rowKey(props.row)}`

  return (
    <tr id={detailRegionId} className='bg-muted/20 border-t'>
      <td colSpan={mainTableColumnCount} className='p-0'>
        <div className='space-y-3 p-3' role='region' aria-label={t('Details')}>
          <div className='flex flex-wrap items-center justify-between gap-2'>
            <div>
              <h3 className='text-sm font-semibold'>{t('Details')}</h3>
              <p className='text-muted-foreground text-xs'>
                {props.row.model_name} · {t('Channel')}{' '}
                {props.row.channel_id || '-'}
              </p>
            </div>
            {isFetching && (
              <span className='text-muted-foreground inline-flex items-center gap-1 text-xs'>
                <Loader2 className='size-3.5 animate-spin' aria-hidden='true' />
                {t('Loading')}
              </span>
            )}
          </div>

          {isError && (
            <p className='text-destructive text-sm'>
              {t('Failed to load logs')}
            </p>
          )}
          {!isError && items.length === 0 && !isFetching && (
            <p className='text-muted-foreground text-sm'>
              {t('No Logs Found')}
            </p>
          )}
          {!isError && (items.length > 0 || isFetching) && (
            <div className='overflow-x-auto rounded-md border'>
              <table className='w-full min-w-[980px] text-xs'>
                <thead className='bg-muted/40 text-left'>
                  <tr>
                    <th className='p-2'>{t('Time')}</th>
                    <th className='p-2'>{t('Username')}</th>
                    <th className='p-2'>{t('Status')}</th>
                    <th className='p-2'>{t('Total tokens')}</th>
                    <th className='p-2'>{t('Upstream cost')}</th>
                    <th className='p-2'>{t('User revenue')}</th>
                    <th className='p-2'>{t('Gross profit')}</th>
                    <th className='p-2'>{t('Profit rate')}</th>
                    <th className='p-2'>{t('Usage')}</th>
                    <th className='p-2'>
                      <span className='sr-only'>{t('Details')}</span>
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {items.map((log) => {
                    let status = log.status || '-'
                    if (log.success === true) status = t('Success')
                    if (log.success === false) status = t('Error')
                    return (
                      <tr key={log.id} className='border-t'>
                        <td className='p-2 font-mono'>
                          {formatTimestamp(log.created_at)}
                        </td>
                        <td className='p-2'>{log.username || '-'}</td>
                        <td className='p-2'>{status}</td>
                        <td className='p-2 font-mono'>{log.total_tokens}</td>
                        <td className='p-2 font-mono'>
                          {money(log.actual_cost_micros, log.currency)}
                        </td>
                        <td className='p-2 font-mono'>
                          {money(log.revenue_micros, log.currency)}
                        </td>
                        <td className='p-2 font-mono'>
                          {money(log.gross_profit_micros, log.currency)}
                        </td>
                        <td className='p-2 font-mono'>
                          {(log.profit_rate * 100).toFixed(2)}%
                        </td>
                        <td className='p-2'>
                          {log.usage_available
                            ? t('Available')
                            : t('Missing usage')}
                        </td>
                        <td className='p-2'>
                          <Button
                            type='button'
                            variant='outline'
                            size='xs'
                            onClick={() => setSelectedLog(log)}
                          >
                            {t('Details')}
                          </Button>
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>
          )}

          {total > detailPageSize && (
            <div className='flex items-center justify-end gap-2'>
              <span className='text-muted-foreground text-xs'>
                {t('Page')} {page} / {Math.ceil(total / detailPageSize)}
              </span>
              <Button
                type='button'
                variant='outline'
                size='xs'
                disabled={page <= 1 || isFetching}
                onClick={() => setPage((value) => Math.max(1, value - 1))}
              >
                {t('Previous')}
              </Button>
              <Button
                type='button'
                variant='outline'
                size='xs'
                disabled={page * detailPageSize >= total || isFetching}
                onClick={() => setPage((value) => value + 1)}
              >
                {t('Next')}
              </Button>
            </div>
          )}
        </div>
        {selectedLog && (
          <DetailsDialog
            log={toSafeUsageLog(selectedLog)}
            isAdmin
            isRoot={false}
            open
            onOpenChange={(open) => {
              if (!open) setSelectedLog(null)
            }}
          />
        )}
      </td>
    </tr>
  )
}

function CostStatisticsRow(props: {
  row: CostStatistics['rows'][number]
  filters: ReportParams
  expanded: boolean
  onToggle: () => void
}) {
  const { t } = useTranslation()
  const key = rowKey(props.row)
  const detailsId = `cost-statistics-details-${key}`

  return (
    <>
      <tr className='border-t' aria-expanded={props.expanded}>
        <td className='p-2'>
          <Button
            type='button'
            variant='ghost'
            size='icon-sm'
            aria-label={t('Details')}
            aria-expanded={props.expanded}
            aria-controls={detailsId}
            onClick={props.onToggle}
          >
            {props.expanded ? (
              <ChevronDown aria-hidden='true' />
            ) : (
              <ChevronRight aria-hidden='true' />
            )}
          </Button>
        </td>
        <td className='p-2 font-mono'>{props.row.model_name}</td>
        <td className='p-2'>{props.row.channel_id || '-'}</td>
        <td className='p-2'>{props.row.request_count}</td>
        <td className='p-2'>
          {props.row.success_count} / {props.row.failure_count}
        </td>
        <td className='p-2'>{props.row.prompt_tokens}</td>
        <td className='p-2'>{props.row.completion_tokens}</td>
        <td className='p-2'>{props.row.cached_tokens}</td>
        <td className='p-2'>
          {props.row.audio_input_tokens} / {props.row.audio_output_tokens}
        </td>
        <td className='p-2'>{props.row.image_count}</td>
        <td className='p-2'>
          {props.row.audio_seconds} / {props.row.audio_output_seconds}
        </td>
        <td className='p-2'>{props.row.video_seconds}</td>
        <td className='p-2'>{props.row.total_tokens}</td>
        <td className='p-2'>
          {money(props.row.actual_cost_micros, props.row.currency)}
        </td>
        <td className='p-2'>
          {money(props.row.revenue_micros, props.row.currency)}
        </td>
        <td className='p-2'>
          {money(props.row.gross_profit_micros, props.row.currency)}
        </td>
        <td className='p-2'>{(props.row.profit_rate * 100).toFixed(2)}%</td>
        <td className='p-2'>{props.row.usage_missing_count}</td>
      </tr>
      <CostStatisticsRowDetails
        row={props.row}
        filters={props.filters}
        expanded={props.expanded}
      />
    </>
  )
}

export function CostStatisticsPage() {
  const { t } = useTranslation()
  const [model, setModel] = useState('')
  const [username, setUsername] = useState('')
  const [channel, setChannel] = useState('')
  const [group, setGroup] = useState('')
  const [start, setStart] = useState('')
  const [end, setEnd] = useState('')
  const [filters, setFilters] = useState<ReportParams>({})
  const [expandedRow, setExpandedRow] = useState<string | null>(null)

  const { data, isFetching } = useQuery({
    queryKey: ['cost-statistics', filters],
    queryFn: async () => {
      const result = await getCostStatistics(filters)
      return result.success ? result.data || emptyStats : emptyStats
    },
    placeholderData: emptyStats,
  })
  const stats = data || emptyStats

  const reportParams: ReportParams = {
    model_name: model || undefined,
    username: username || undefined,
    channel: channel ? Number(channel) : undefined,
    group: group || undefined,
    start_timestamp: toTimestamp(start, false),
    end_timestamp: toTimestamp(end, true),
  }

  const applyFilters = () => {
    setExpandedRow(null)
    setFilters(reportParams)
  }

  const downloadCsv = async () => {
    try {
      const blob = await exportCostStatistics(reportParams)
      const url = URL.createObjectURL(blob)
      const anchor = document.createElement('a')
      anchor.href = url
      anchor.download = 'cost-statistics.csv'
      anchor.click()
      URL.revokeObjectURL(url)
    } catch {
      toast.error(t('Failed to export cost report'))
    }
  }

  return (
    <div className='flex flex-col gap-4 overflow-auto'>
      <div className='rounded-lg border p-4'>
        <p className='text-muted-foreground mb-3 text-sm'>
          {t(
            'Revenue is derived from the quota recorded at request time. Cost is the immutable upstream snapshot recorded in the same log.'
          )}
        </p>
        <div className='grid gap-2 md:grid-cols-3 lg:grid-cols-6'>
          <Input
            aria-label={t('Model')}
            placeholder={t('Model')}
            value={model}
            onChange={(event) => setModel(event.target.value)}
          />
          <Input
            aria-label={t('Username')}
            placeholder={t('Username')}
            value={username}
            onChange={(event) => setUsername(event.target.value)}
          />
          <Input
            aria-label={t('Channel ID')}
            placeholder={t('Channel ID')}
            inputMode='numeric'
            value={channel}
            onChange={(event) => setChannel(event.target.value)}
          />
          <Input
            aria-label={t('Group')}
            placeholder={t('Group')}
            value={group}
            onChange={(event) => setGroup(event.target.value)}
          />
          <Input
            aria-label={t('Start date')}
            type='date'
            value={start}
            onChange={(event) => setStart(event.target.value)}
          />
          <Input
            aria-label={t('End date')}
            type='date'
            value={end}
            onChange={(event) => setEnd(event.target.value)}
          />
        </div>
        <div className='mt-3 flex flex-wrap gap-2'>
          <Button onClick={applyFilters} disabled={isFetching}>
            {t('Refresh report')}
          </Button>
          <Button variant='outline' onClick={downloadCsv} disabled={isFetching}>
            {t('Export CSV')}
          </Button>
        </div>
      </div>
      <div className='grid gap-3 md:grid-cols-4'>
        {[
          [t('Requests'), stats.request_count],
          [t('Successes'), stats.success_count],
          [t('Failures'), stats.failure_count],
          [t('Total tokens'), stats.total_tokens],
          [
            t('Upstream cost'),
            summaryMoney(
              stats.actual_cost_micros,
              stats.currency,
              stats.mixed_currency,
              t('Mixed currencies'),
            ),
          ],
          [
            t('User revenue'),
            summaryMoney(
              stats.revenue_micros,
              stats.currency,
              stats.mixed_currency,
              t('Mixed currencies'),
            ),
          ],
          [
            t('Gross profit'),
            summaryMoney(
              stats.gross_profit_micros,
              stats.currency,
              stats.mixed_currency,
              t('Mixed currencies'),
            ),
          ],
          [t('Profit rate'), `${(stats.profit_rate * 100).toFixed(2)}%`],
        ].map(([label, value]) => (
          <div key={String(label)} className='rounded-lg border p-3'>
            <div className='text-muted-foreground text-xs'>{label}</div>
            <div className='mt-1 font-mono font-semibold tabular-nums'>
              {value}
            </div>
          </div>
        ))}
      </div>
      <div className='overflow-x-auto rounded-lg border'>
        <table className='w-full min-w-[1220px] text-sm'>
          <thead className='bg-muted/40 text-left'>
            <tr>
              <th className='p-2'>
                <span className='sr-only'>{t('Details')}</span>
              </th>
              {[
                t('Model'),
                t('Channel'),
                t('Requests'),
                t('Success / failure'),
                t('Input tokens'),
                t('Output tokens'),
                t('Cached tokens'),
                t('Audio input / output tokens'),
                t('Images'),
                t('Audio input / output seconds'),
                t('Video seconds'),
                t('Total tokens'),
                t('Upstream cost'),
                t('Revenue'),
                t('Gross profit'),
                t('Profit rate'),
                t('Missing usage'),
              ].map((label) => (
                <th key={label} className='p-2'>
                  {label}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {stats.rows.map((row) => {
              const key = rowKey(row)
              return (
                <CostStatisticsRow
                  key={key}
                  row={row}
                  filters={filters}
                  expanded={expandedRow === key}
                  onToggle={() =>
                    setExpandedRow((current) => (current === key ? null : key))
                  }
                />
              )
            })}
            {stats.rows.length === 0 && (
              <tr>
                <td
                  colSpan={mainTableColumnCount}
                  className='text-muted-foreground p-8 text-center'
                >
                  {t('No cost data found')}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
