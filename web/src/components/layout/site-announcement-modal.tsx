"use client";

import { useEffect, useMemo, useState } from "react";
import { Button, Checkbox, Modal, Typography } from "antd";

import { useConfigStore } from "@/stores/use-config-store";

const storageKeyPrefix = "infinite-canvas:site-announcement:";
const dismissedTodayKey = "infinite-canvas:site-announcement:dismissed-today";

export function SiteAnnouncementModal() {
    const announcement = useConfigStore((state) => state.publicSettings?.announcement);
    const [open, setOpen] = useState(false);
    const [dontShowToday, setDontShowToday] = useState(false);
    const storageKey = useMemo(() => `${storageKeyPrefix}${announcement?.version || "default"}`, [announcement?.version]);

    const isDismissedToday = () => {
        if (typeof window === "undefined") return false;
        const today = new Date().toISOString().slice(0, 10);
        const stored = window.localStorage.getItem(dismissedTodayKey);
        return stored === today;
    };

    useEffect(() => {
        if (!announcement?.enabled || !announcement.content?.trim()) {
            setOpen(false);
            return;
        }
        if (announcement.oncePerVersion && typeof window !== "undefined" && window.localStorage.getItem(storageKey) === "1") {
            setOpen(false);
            return;
        }
        if (isDismissedToday()) {
            setOpen(false);
            return;
        }
        setDontShowToday(false);
        setOpen(true);
    }, [announcement, storageKey]);

    const close = () => {
        if (dontShowToday && typeof window !== "undefined") {
            const today = new Date().toISOString().slice(0, 10);
            window.localStorage.setItem(dismissedTodayKey, today);
        } else if (announcement?.oncePerVersion && typeof window !== "undefined") {
            window.localStorage.setItem(storageKey, "1");
        }
        setDontShowToday(false);
        setOpen(false);
    };

    return (
        <Modal title={announcement?.title || "网站公告"} open={open} onCancel={close} footer={[
            <Checkbox key="dont-show" checked={dontShowToday} onChange={(e) => setDontShowToday(e.target.checked)}>
                今日不再显示
            </Checkbox>,
            <Button key="ok" type="primary" onClick={close}>知道了</Button>,
        ]} centered destroyOnHidden>
            <Typography.Paragraph style={{ whiteSpace: "pre-wrap", marginBottom: 0 }}>{announcement?.content}</Typography.Paragraph>
        </Modal>
    );
}
