import { useState } from "react";
import { RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useSyncDlq, useRetryDlq, useRetryAllDlq } from "@/hooks/use-sync";
import { formatDate } from "@/lib/utils";

export function DlqTable() {
  const [page, setPage] = useState(1);
  const pageSize = 10;

  const { data, isLoading } = useSyncDlq({ page, page_size: pageSize });
  const retryOne = useRetryDlq();
  const retryAll = useRetryAllDlq();

  const entries = data?.data ?? [];
  const meta = data?.meta;
  const totalPages = meta?.total_pages ?? 1;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold">Dead Letter Queue</h2>
        {entries.length > 0 && (
          <Button
            variant="outline"
            size="sm"
            onClick={() => retryAll.mutate()}
            disabled={retryAll.isPending}
          >
            <RefreshCw className="mr-2 h-4 w-4" />
            Retry All
          </Button>
        )}
      </div>

      {isLoading ? (
        <p className="text-muted-foreground">Loading...</p>
      ) : entries.length === 0 ? (
        <div className="rounded-md border p-8 text-center text-muted-foreground">
          No items in the dead letter queue.
        </div>
      ) : (
        <div className="overflow-auto rounded-md border">
          <table className="w-full">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="px-3 py-2 text-left text-sm font-medium text-muted-foreground">
                  Document
                </th>
                <th className="px-3 py-2 text-left text-sm font-medium text-muted-foreground">
                  Error
                </th>
                <th className="px-3 py-2 text-left text-sm font-medium text-muted-foreground">
                  Retries
                </th>
                <th className="px-3 py-2 text-left text-sm font-medium text-muted-foreground">
                  Created
                </th>
                <th className="px-3 py-2 text-right text-sm font-medium text-muted-foreground">
                  Action
                </th>
              </tr>
            </thead>
            <tbody>
              {entries.map((entry) => (
                <tr key={entry.id} className="border-b">
                  <td className="px-3 py-2 text-sm">
                    {entry.document_title ?? `Doc #${entry.document_id}`}
                  </td>
                  <td className="max-w-xs truncate px-3 py-2 text-sm text-red-600" title={entry.error_message}>
                    {entry.error_message}
                  </td>
                  <td className="px-3 py-2 text-sm text-muted-foreground">
                    {entry.retry_count}/{entry.max_retries}
                  </td>
                  <td className="px-3 py-2 text-sm text-muted-foreground">
                    {formatDate(entry.created_at)}
                  </td>
                  <td className="px-3 py-2 text-right">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => retryOne.mutate(entry.id)}
                      disabled={retryOne.isPending}
                    >
                      <RefreshCw className="h-4 w-4" />
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={page <= 1}
            onClick={() => setPage((p) => p - 1)}
          >
            Previous
          </Button>
          <span className="text-sm text-muted-foreground">
            Page {page} of {totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            disabled={page >= totalPages}
            onClick={() => setPage((p) => p + 1)}
          >
            Next
          </Button>
        </div>
      )}
    </div>
  );
}
