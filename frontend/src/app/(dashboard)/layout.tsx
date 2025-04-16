import { DashboardLayout } from '@/components';
import { AuthGuard } from '@/components';

// This layout will be shared by all protected routes in the (dashboard) group
export default function DashboardRouteLayout({ children }: { children: React.ReactNode }) {
    return (
        <AuthGuard>
            <DashboardLayout>{children}</DashboardLayout>
        </AuthGuard>
    );
}
