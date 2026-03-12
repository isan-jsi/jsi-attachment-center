import { useState } from "react";
import { CheckCircle2, XCircle, Clock, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useSyncLogs } from "@/hooks/use-sync";
import { formatDate } from "@/lib/utils";

function StatusIcon({ status }: { status: string }) {
  switch (status.toLowerCase()) {
    case "completed":
    case "success":
      return <CheckCircle2 className="h-4 w-4 text-green-500" />;
    case "failed":
    case "error":
      return <XCircle className="h-4 w-4 text-red-500" />;
    case "running":
    case "in_progress":
      return <Loader2 className="h-4 w-4 text-blue-500 animate-spin" />;
    default:
      return <Clock className="h-4 w-4 text-muted-foreground" />;
  }
}

export function SyncLogFeed() {
  const [page, setPage] = useState(1);
  const pageSize = 15;

  const { data, isLoading } = useSyncLogs({
    page,
    page_size: pageSize,
    sort_by: "started_at",
    sort_dir: "desc",
  });

  const logs = data?.data ?? [];
  const meta = data?.meta;
  const totalPages = meta?.total_pages ?? 1;

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold">Sync Logs</h2>

      {isLoading ? (
        <p className="text-muted-foreground">Loading logs...</p>
      ) : logs.length === 0 ? (
        <div className="rounded-md border p-8 text-center text-muted-foreground">
          No sync logs available.
        </div>
      ) : (
        <div className="space-y-2">
          {logs.map((log) => (
            <div
              key={log.id}
              className="flex items-start gap-3 rounded-md border p-3"
            >
              <StatusIcon status={log.status} />
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium">
                    {log.sync_type}
                  </span>
                  <span className="text-xs text-muted-foreground">
                    {formatDate(log.started_at)}
                  </span>
                </div>
                <p className="text-xs text-muted-foreground">
                  Processed: {log.documents_processed} | Failed:{" "}
                  {log.documents_failed}
                  {log.completed_at && (
                    <> | Completed: {formatDate(log.completed_at)}</>
                  )}
                </p>
                {log.error_message && (
                  <p className="mt-1 text-xs text-red-600 truncate" title={log.error_message}>
                    {log.error_message}
                  </p>
                )}
              </div>
            </div>
          ))}
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
