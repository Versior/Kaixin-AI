import type { Metadata } from "next";
import { AntdRegistry } from "@ant-design/nextjs-registry";
import { AppProviders } from "@/components/layout/app-providers";
import "antd/dist/reset.css";
import "./globals.css";
import React from "react";

export const metadata: Metadata = {
    title: "灵感事务所",
    description: "高质感 AI 生图平台",
};

export default function RootLayout({
    children,
}: Readonly<{
    children: React.ReactNode;
}>) {
    return (
        <html lang="zh-CN" suppressHydrationWarning className="font-sans">
            <body
                className="bg-background text-foreground antialiased"
                style={{
                    fontFamily: '"SF Pro Display","SF Pro Text","PingFang SC","Microsoft YaHei","Helvetica Neue",sans-serif',
                }}
            >
                <script
                    dangerouslySetInnerHTML={{
                        __html: `!function(){try{var e=JSON.parse(localStorage.getItem("linggan-sws:theme_store")||localStorage.getItem("infinite-canvas:theme_store")||"null"),t=e&&e.state&&(e.state.theme==="light"||e.state.theme==="dark")?e.state.theme:window.matchMedia("(prefers-color-scheme:dark)").matches?"dark":"light";document.documentElement.classList.toggle("dark",t==="dark"),document.documentElement.style.colorScheme=t}catch(e){document.documentElement.classList.add("dark"),document.documentElement.style.colorScheme="dark"}}()`,
                    }}
                />
                <AntdRegistry>
                    <AppProviders>{children}</AppProviders>
                </AntdRegistry>
            </body>
        </html>
    );
}
