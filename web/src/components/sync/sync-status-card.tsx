import {
  RefreshCw,
  CheckCircle2,
  AlertTriangle,
  Clock,
  FileText,
  XCircle,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { useSyncStatus, useTriggerSync } from "@/hooks/use-sync";
import { formatDate } from "@/lib/utils";
import { cn } from "@/lib/utils";

export function SyncStatusCard() {
  const { data: status, isLoading } = useSyncStatus();
  const triggerSync = useTriggerSync();

  if (isLoading || !status) {
    return (
      <div className="rounded-lg border p-6">
        <p className="text-muted-foreground">Loading sync status...</p>
      </div>
    );
  }

  return (
    <div className="rounded-lg border p-6 space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold">Sync Status</h2>
        <Button
          size="sm"
          onClick={() => triggerSync.mutate()}
          disabled={status.is_running || triggerSync.isPending}
        >
          <RefreshCw
            className={cn(
              "mr-2 h-4 w-4",
              (status.is_running || triggerSync.isPending) && "animate-spin"
            )}
          />
          {status.is_running ? "Syncing..." : "Trigger Sync"}
        </Button>
      </div>

      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
        <StatCard
          icon={
            status.is_running ? (
              <RefreshCw className="h-5 w-5 text-blue-500 animate-spin" />
            ) : (
              <CheckCircle2 className="h-5 w-5 text-green-500" />
            )
          }
          label="Status"
          value={status.is_running ? "Running" : "Idle"}
        />
        <StatCard
          icon={<Clock className="h-5 w-5 text-muted-foreground" />}
          label="Last Sync"
          value={
            status.last_sync_at ? formatDate(status.last_sync_at) : "Never"
          }
        />
        <StatCard
          icon={<FileText className="h-5 w-5 text-blue-500" />}
          label="Synced"
          value={String(status.documents_synced)}
        />
        <StatCard
          icon={<XCircle className="h-5 w-5 text-red-500" />}
          label="Failed"
          value={String(status.documents_failed)}
        />
        <StatCard
          icon={<AlertTriangle className="h-5 w-5 text-yellow-500" />}
          label="DLQ"
          value={String(status.dlq_count)}
        />
      </div>

      {status.next_sync_at && (
        <p className="text-xs text-muted-foreground">
          Next sync scheduled: {formatDate(status.next_sync_at)}
        </p>
      )}
    </div>
  );
}

function StatCard({
  icon,
  label,
  value,
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
}) {
  return (
    <div className="flex items-center gap-3 rounded-md border p-3">
      {icon}
      <div>
        <p className="text-xs text-muted-foreground">{label}</p>
        <p className="text-sm font-semibold">{value}</p>
      </div>
    </div>
  );
}
