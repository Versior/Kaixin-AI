"use client";

import type { ReactNode } from "react";
import { Spin } from "antd";
import Link from "next/link";

import { useUserStore } from "@/stores/use-user-store";

export function RequireLogin({ children }: { children: ReactNode }) {
    const user = useUserStore((state) => state.user);
    const isReady = useUserStore((state) => state.isReady);

    if (!isReady) {
        return (
            <div className="flex h-full items-center justify-center bg-background text-stone-500 dark:text-stone-400">
                <Spin />
            </div>
        );
    }

    if (!user) {
        return (
            <main className="flex h-full items-center justify-center bg-background px-6 text-center text-stone-950 dark:text-stone-100">
                <section className="max-w-md rounded-3xl border border-stone-200 bg-white p-8 shadow-sm dark:border-stone-800 dark:bg-stone-900">
                    <h1 className="text-2xl font-semibold">请先登录</h1>
                    <p className="mt-3 text-sm leading-6 text-stone-500 dark:text-stone-400">灵感库、我的素材、生图工作台和生成历史都按账号保存。登录后才能使用。</p>
                    <Link href="/login" className="mt-6 inline-flex rounded-full bg-stone-950 px-5 py-2.5 text-sm font-medium text-white transition hover:bg-stone-800 dark:bg-stone-100 dark:text-stone-950 dark:hover:bg-white">
                        去登录
                    </Link>
                </section>
            </main>
        );
    }

    return <>{children}</>;
}
