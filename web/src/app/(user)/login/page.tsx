"use client";

import { LockOutlined, MailOutlined, UserOutlined } from "@ant-design/icons";
import { App, Button, Form, Input, Segmented, Space } from "antd";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";

import { fetchCurrentUser, sendVerificationCode } from "@/services/api/auth";
import { useConfigStore } from "@/stores/use-config-store";
import { useUserStore } from "@/stores/use-user-store";

type LoginFormValues = {
    username: string;
    password: string;
    email?: string;
    verificationCode?: string;
    confirmPassword?: string;
};

export default function LoginPage() {
    return (
        <Suspense fallback={null}>
            <LoginContent />
        </Suspense>
    );
}

function LoginContent() {
    const { message } = App.useApp();
    const router = useRouter();
    const searchParams = useSearchParams();
    const login = useUserStore((state) => state.login);
    const register = useUserStore((state) => state.register);
    const setSession = useUserStore((state) => state.setSession);
    const isLoading = useUserStore((state) => state.isLoading);
    const linuxDoEnabled = useConfigStore((state) => state.publicSettings?.auth?.linuxDo?.enabled === true);
    const allowRegister = useConfigStore((state) => state.publicSettings?.auth?.allowRegister !== false);
    const smtpEnabled = useConfigStore((state) => state.publicSettings?.auth?.smtpEnabled === true);
    const [mode, setMode] = useState<"login" | "register">("login");
    const redirect = searchParams.get("redirect") || "/";
    const [form] = Form.useForm<LoginFormValues>();

    // 发送验证码状态
    const [codeSent, setCodeSent] = useState(false);
    const [countdown, setCountdown] = useState(0);
    const [sendingCode, setSendingCode] = useState(false);

    useEffect(() => {
        const token = searchParams.get("token");
        const error = searchParams.get("error");
        if (error) message.error(error);
        if (!token) return;
        void fetchCurrentUser(token).then((user) => {
            setSession(token, user);
            message.success("登录成功");
            router.replace(redirect.startsWith("/") ? redirect : "/");
            router.refresh();
        });
    }, [message, redirect, router, searchParams, setSession]);

    useEffect(() => {
        if (!allowRegister && mode === "register") setMode("login");
    }, [allowRegister, mode]);

    // 验证码倒计时
    useEffect(() => {
        if (countdown <= 0) return;
        const timer = setTimeout(() => setCountdown((c) => c - 1), 1000);
        return () => clearTimeout(timer);
    }, [countdown]);

    // 切换模式时重置验证码状态
    useEffect(() => {
        setCodeSent(false);
        setCountdown(0);
    }, [mode]);

    const handleSendCode = async (email: string) => {
        if (!email) {
            message.warning("请先填写邮箱");
            return;
        }
        setSendingCode(true);
        try {
            await sendVerificationCode(email);
            message.success("验证码已发送，请查收邮件");
            setCodeSent(true);
            setCountdown(60);
        } catch (error) {
            message.error(error instanceof Error ? error.message : "发送失败");
        } finally {
            setSendingCode(false);
        }
    };

    const submit = async (values: LoginFormValues) => {
        try {
            if (mode === "register" && !allowRegister) {
                message.error("当前未开放注册");
                return;
            }
            if (mode === "register" && values.password !== values.confirmPassword) {
                message.error("两次输入的密码不一致");
                return;
            }
            const action = mode === "register" ? register : login;
            const payload = mode === "register"
                ? smtpEnabled
                    ? { username: values.username, password: values.password, email: values.email, verificationCode: values.verificationCode }
                    : { username: values.username, password: values.password }
                : { username: values.username, password: values.password };
            const user = await action(payload);
            message.success(mode === "register" ? "注册成功" : "登录成功");
            router.replace(redirect.startsWith("/") ? redirect : "/");
            router.refresh();
            if (user.role !== "admin") router.replace("/");
        } catch (error) {
            message.error(error instanceof Error ? error.message : "操作失败");
        }
    };

    return (
        <main className="flex h-full min-h-0 items-center justify-center overflow-y-auto bg-background bg-[radial-gradient(#e5e7eb_1px,transparent_1px)] px-6 py-10 [background-size:16px_16px] dark:bg-[radial-gradient(rgba(245,245,244,.16)_1px,transparent_1px)]">
            <section className="w-full max-w-[420px]">
                <div className="mb-7 text-center">
                    <span
                        className="mx-auto mb-4 block size-12 bg-stone-950 dark:bg-stone-100"
                        style={{
                            mask: "url(/logo.svg) center / contain no-repeat",
                            WebkitMask: "url(/logo.svg) center / contain no-repeat",
                        }}
                        aria-label="灵感事务所"
                    />
                    <h1 className="text-3xl font-semibold tracking-normal text-stone-950 dark:text-stone-100">账号登录</h1>
                    <p className="mt-3 text-base leading-7 text-stone-500 dark:text-stone-400">在这里，灵感生根发芽。</p>
                </div>

                <Form<LoginFormValues> form={form} layout="vertical" size="large" requiredMark={false} onFinish={submit}>
                    <Form.Item>
                        <Segmented
                            block
                            value={mode}
                            onChange={(value) => setMode(value as "login" | "register")}
                            options={allowRegister ? [{ label: "登录", value: "login" }, { label: "注册", value: "register" }] : [{ label: "登录", value: "login" }]}
                        />
                    </Form.Item>

                    {/* 注册时先填邮箱 — 仅 SMTP 启用时显示 */}
                    {smtpEnabled && mode === "register" ? (
                        <Form.Item label={<span className="font-medium text-stone-800 dark:text-stone-200">邮箱</span>} style={{ marginBottom: 16 }}>
                            <Space.Compact style={{ width: "100%" }}>
                                <Form.Item name="email" noStyle rules={[{ required: true, message: "请输入邮箱" }, { type: "email", message: "邮箱格式不正确" }]}>
                                    <Input prefix={<MailOutlined />} placeholder="请输入邮箱" style={{ width: "calc(100% - 120px)" }} />
                                </Form.Item>
                                <Button
                                    style={{ width: 120 }}
                                    loading={sendingCode}
                                    disabled={countdown > 0}
                                    onClick={() => {
                                        const email = form.getFieldValue("email");
                                        void handleSendCode(email);
                                    }}
                                >
                                    {countdown > 0 ? `${countdown}s` : codeSent ? "重新发送" : "发送验证码"}
                                </Button>
                            </Space.Compact>
                        </Form.Item>
                    ) : null}

                    {smtpEnabled && mode === "register" ? (
                        <Form.Item name="verificationCode" label={<span className="font-medium text-stone-800 dark:text-stone-200">验证码</span>} rules={[{ required: true, message: "请输入验证码" }]}>
                            <Input placeholder="请输入6位验证码" maxLength={6} />
                        </Form.Item>
                    ) : null}

                    <Form.Item name="username" label={<span className="font-medium text-stone-800 dark:text-stone-200">用户名</span>} rules={[{ required: true, message: "请输入用户名" }]}>
                        <Input prefix={<UserOutlined />} autoComplete="username" />
                    </Form.Item>
                    <Form.Item name="password" label={<span className="font-medium text-stone-800 dark:text-stone-200">密码</span>} rules={[{ required: true, message: "请输入密码" }]}>
                        <Input.Password prefix={<LockOutlined />} autoComplete={mode === "register" ? "new-password" : "current-password"} />
                    </Form.Item>
                    {mode === "register" ? (
                        <Form.Item name="confirmPassword" label={<span className="font-medium text-stone-800 dark:text-stone-200">确认密码</span>} rules={[{ required: true, message: "请再次输入密码" }]}>
                            <Input.Password prefix={<LockOutlined />} autoComplete="new-password" />
                        </Form.Item>
                    ) : null}
                    <Space orientation="vertical" size={12} style={{ width: "100%" }}>
                        <Button block type="primary" htmlType="submit" loading={isLoading}>
                            {mode === "register" ? "注册" : "登录"}
                        </Button>
                        {linuxDoEnabled ? (
                            <Button block href={`/api/auth/linux-do/authorize?redirect=${encodeURIComponent(redirect)}`} icon={<img src="/icons/linuxdo.svg" alt="" width={18} height={18} />}>
                                使用 Linux.do 登录
                            </Button>
                        ) : null}
                    </Space>
                </Form>
            </section>
        </main>
    );
}
