"use client";

import type { ReactNode } from "react";

import { AppTopNav } from "@/components/layout/app-top-nav";
import { SiteAnnouncementModal } from "@/components/layout/site-announcement-modal";

export default function UserLayout({ children }: { children: ReactNode }) {
    return (
        <div className="flex h-dvh flex-col overflow-hidden bg-background text-foreground">
            <AppTopNav />
            <SiteAnnouncementModal />
            <div className="min-h-0 flex-1 overflow-hidden">{children}</div>
        </div>
    );
}
