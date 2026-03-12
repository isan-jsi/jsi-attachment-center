import { SyncStatusCard } from "@/components/sync/sync-status-card";
import { DlqTable } from "@/components/sync/dlq-table";
import { SyncLogFeed } from "@/components/sync/sync-log-feed";

export default function SyncDashboard() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Sync Dashboard</h1>
        <p className="text-sm text-muted-foreground">
          Monitor and manage document synchronization.
        </p>
      </div>

      <SyncStatusCard />

      <div className="grid gap-6 lg:grid-cols-2">
        <DlqTable />
        <SyncLogFeed />
      </div>
    </div>
  );
}
