'use client';

import React from 'react';

export const ACCENT = 'var(--color-primary)';
export const ACCENT_SOFT = 'color-mix(in srgb, var(--color-primary) 10%, white)';
export const ACCENT_BORDER = 'color-mix(in srgb, var(--color-primary) 25%, white)';
export const ACCENT_DARK = 'color-mix(in srgb, var(--color-primary) 75%, black)';
export const CARD_BG = '#fafafd';
export const CARD_BORDER = '#eceaf3';
export const INPUT_BORDER = '#e9e9f0';

export const fieldLabel = 'block text-[13px] font-semibold text-gray-900 dark:text-gray-100 mb-1';
export const fieldHint = 'text-[11px] text-gray-400 dark:text-gray-500 mt-1';
export const inputCls =
    'w-full px-3 py-2 text-sm rounded-lg border bg-white dark:bg-gray-800 text-gray-900 dark:text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-[var(--color-primary)]/30 focus:border-[var(--color-primary)]';

interface CardProps {
    icon?: React.ReactNode;
    title: string;
    description?: React.ReactNode;
    action?: React.ReactNode;
    children?: React.ReactNode;
}

export function Card({ icon, title, description, action, children }: CardProps) {
    return (
        <div
            className="rounded-xl p-4 dark:bg-gray-800/60"
            style={{ backgroundColor: CARD_BG, border: `1px solid ${CARD_BORDER}` }}
        >
            <div className="flex items-start justify-between gap-3 mb-3">
                <div className="flex items-start gap-2 min-w-0">
                    {icon && (
                        <span
                            className="flex h-7 w-7 items-center justify-center rounded-md shrink-0"
                            style={{ backgroundColor: ACCENT_SOFT }}
                        >
                            {icon}
                        </span>
                    )}
                    <div className="min-w-0">
                        <h3 className="text-sm font-bold text-gray-900 dark:text-white leading-tight">{title}</h3>
                        {description && (
                            <p className="text-[11px] text-gray-500 dark:text-gray-400 mt-1 leading-relaxed">
                                {description}
                            </p>
                        )}
                    </div>
                </div>
                {action && <div className="shrink-0">{action}</div>}
            </div>
            {children}
        </div>
    );
}

interface PillProps {
    tone: 'ok' | 'warn' | 'off';
    children: React.ReactNode;
}

const PILL_TONES: Record<PillProps['tone'], { dot: string; text: string; border: string }> = {
    ok: { dot: ACCENT, text: ACCENT_DARK, border: ACCENT_BORDER },
    warn: { dot: '#f59e0b', text: '#b45309', border: '#fde68a' },
    off: { dot: '#9ca3af', text: '#6b7280', border: '#e5e7eb' },
};

export function Pill({ tone, children }: PillProps) {
    const style = PILL_TONES[tone];
    return (
        <span
            className="inline-flex items-center gap-2 rounded-full px-3 py-1 text-[11px] font-semibold bg-white dark:bg-gray-900"
            style={{ border: `1px solid ${style.border}`, color: style.text }}
        >
            <span className="h-2 w-2 rounded-full" style={{ backgroundColor: style.dot }} />
            {children}
        </span>
    );
}

interface StepsProps {
    pasos: { key: string; label: string }[];
    actual: number;
}

export function Steps({ pasos, actual }: StepsProps) {
    return (
        <ol className="flex flex-wrap items-center gap-1.5 mb-4">
            {pasos.map((paso, idx) => {
                const hecho = idx < actual;
                const activo = idx === actual;
                return (
                    <li
                        key={paso.key}
                        className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[11px] font-semibold"
                        style={
                            hecho
                                ? { backgroundColor: ACCENT_SOFT, color: ACCENT_DARK }
                                : activo
                                  ? { backgroundColor: 'white', color: ACCENT_DARK, border: `1px solid ${ACCENT_BORDER}` }
                                  : { backgroundColor: '#f3f4f6', color: '#9ca3af' }
                        }
                    >
                        <span
                            className="flex h-4 w-4 items-center justify-center rounded-full text-[10px] font-bold text-white"
                            style={{ backgroundColor: hecho || activo ? ACCENT : '#d1d5db' }}
                        >
                            {idx + 1}
                        </span>
                        {paso.label}
                    </li>
                );
            })}
        </ol>
    );
}

interface ActionButtonProps {
    children: React.ReactNode;
    onClick?: () => void;
    disabled?: boolean;
    loading?: boolean;
    variant?: 'primary' | 'ghost';
    className?: string;
    type?: 'button' | 'submit';
}

export function ActionButton({
    children,
    onClick,
    disabled,
    loading,
    variant = 'primary',
    className = '',
    type = 'button',
}: ActionButtonProps) {
    const base =
        'px-4 py-2 text-[13px] font-semibold rounded-lg inline-flex items-center justify-center gap-2 transition-colors disabled:opacity-60 disabled:cursor-not-allowed';

    if (variant === 'ghost') {
        return (
            <button
                type={type}
                onClick={onClick}
                disabled={disabled || loading}
                className={`${base} bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700 ${className}`}
                style={{ border: `1px solid ${INPUT_BORDER}` }}
            >
                {loading && <Spinner />}
                {children}
            </button>
        );
    }

    return (
        <button
            type={type}
            onClick={onClick}
            disabled={disabled || loading}
            className={`${base} text-white ${className}`}
            style={{ backgroundColor: ACCENT }}
        >
            {loading && <Spinner />}
            {children}
        </button>
    );
}

function Spinner() {
    return (
        <span className="h-3.5 w-3.5 rounded-full border-2 border-white/40 border-t-white animate-spin" />
    );
}
