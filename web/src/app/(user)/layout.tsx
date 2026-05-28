"use client";

import type { ReactNode } from "react";
import { usePathname } from "next/navigation";

import { AppTopNav } from "@/components/layout/app-top-nav";
import { RequireLogin } from "@/components/layout/require-login";
import { SiteAnnouncementModal } from "@/components/layout/site-announcement-modal";

const publicPaths = new Set(["/", "/login", "/prompts"]);

export default function UserLayout({ children }: { children: ReactNode }) {
    return (
        <div className="flex h-dvh flex-col overflow-hidden bg-background text-foreground">
            <AppTopNav />
            <SiteAnnouncementModal />
            <div className="min-h-0 flex-1 overflow-hidden">
                <RequireLoginWrapper>{children}</RequireLoginWrapper>
            </div>
        </div>
    );
}

function RequireLoginWrapper({ children }: { children: ReactNode }) {
    const pathname = usePathname();
    return publicPaths.has(pathname) ? <>{children}</> : <RequireLogin>{children}</RequireLogin>;
}
