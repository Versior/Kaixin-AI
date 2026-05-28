"use client";

import { useEffect, useMemo, useState } from "react";
import { Modal, Typography } from "antd";

import { useConfigStore } from "@/stores/use-config-store";

const storageKeyPrefix = "infinite-canvas:site-announcement:";

export function SiteAnnouncementModal() {
    const announcement = useConfigStore((state) => state.publicSettings?.announcement);
    const [open, setOpen] = useState(false);
    const storageKey = useMemo(() => `${storageKeyPrefix}${announcement?.version || "default"}`, [announcement?.version]);

    useEffect(() => {
        if (!announcement?.enabled || !announcement.content?.trim()) {
            setOpen(false);
            return;
        }
        if (announcement.oncePerVersion && typeof window !== "undefined" && window.localStorage.getItem(storageKey) === "1") {
            setOpen(false);
            return;
        }
        setOpen(true);
    }, [announcement, storageKey]);

    const close = () => {
        if (announcement?.oncePerVersion && typeof window !== "undefined") {
            window.localStorage.setItem(storageKey, "1");
        }
        setOpen(false);
    };

    return (
        <Modal title={announcement?.title || "网站公告"} open={open} onCancel={close} onOk={close} okText="知道了" cancelButtonProps={{ style: { display: "none" } }} centered destroyOnHidden>
            <Typography.Paragraph style={{ whiteSpace: "pre-wrap", marginBottom: 0 }}>{announcement?.content}</Typography.Paragraph>
        </Modal>
    );
}
