'use client';

import { useCallback, useEffect, useState } from 'react';
import { ArrowPathIcon } from '@heroicons/react/24/outline';
import { Alert } from '@/shared/ui';
import { ACCENT, ActionButton, Card, Pill } from './ui-kit';
import {
    getWhatsAppTemplatesStatusAction,
    provisionWhatsAppTemplatesAction,
} from '../../infra/actions';
import { WhatsAppTemplateStatus } from '../../domain/types';

interface WhatsAppTemplatesPanelProps {
    businessId?: number;
    enabled: boolean;
}

const STATUS_STYLES: Record<string, string> = {
    APPROVED: 'bg-green-100 text-green-800',
    PENDING: 'bg-yellow-100 text-yellow-800',
    REJECTED: 'bg-red-100 text-red-700',
    DISABLED: 'bg-gray-200 text-gray-700',
    ERROR: 'bg-red-100 text-red-700',
};

const STATUS_LABELS: Record<string, string> = {
    APPROVED: 'Aprobada',
    PENDING: 'En revisión',
    REJECTED: 'Rechazada',
    DISABLED: 'Deshabilitada',
    ERROR: 'Error',
    UNKNOWN: 'Sin estado',
};

export default function WhatsAppTemplatesPanel({ businessId, enabled }: WhatsAppTemplatesPanelProps) {
    const [templates, setTemplates] = useState<WhatsAppTemplateStatus[]>([]);
    const [wabaId, setWabaId] = useState('');
    const [hosted, setHosted] = useState(false);
    const [loading, setLoading] = useState(false);
    const [provisioning, setProvisioning] = useState(false);
    const [message, setMessage] = useState<{ type: 'success' | 'error' | 'info'; text: string } | null>(null);

    const load = useCallback(
        async (refresh: boolean) => {
            setLoading(true);
            setMessage(null);
            try {
                const result = await getWhatsAppTemplatesStatusAction(businessId, refresh);
                if (result.success && result.data) {
                    setTemplates(result.data.templates || []);
                    setWabaId(result.data.waba_id || '');
                    setHosted(Boolean(result.data.hosted_by_platform));
                } else {
                    setMessage({ type: 'error', text: result.message || 'No se pudo consultar el estado' });
                }
            } catch (err: any) {
                setMessage({ type: 'error', text: err?.message || 'No se pudo consultar el estado' });
            } finally {
                setLoading(false);
            }
        },
        [businessId]
    );

    useEffect(() => {
        if (enabled) {
            load(false);
        }
    }, [enabled, load]);

    const handleProvision = async () => {
        setProvisioning(true);
        setMessage(null);
        try {
            const result = await provisionWhatsAppTemplatesAction(businessId);
            if (result.success && result.data) {
                setTemplates(result.data.templates || []);
                setWabaId(result.data.waba_id || '');
                const failed = Object.keys(result.data.failed || {});
                setMessage({
                    type: failed.length > 0 ? 'error' : 'success',
                    text: `Creadas: ${result.data.created.length}. Ya exist\u00edan: ${
                        result.data.already_exists.length
                    }. Omitidas: ${(result.data.skipped || []).length}.${
                        failed.length > 0 ? ` Fallaron: ${failed.join(', ')}` : ''
                    }`,
                });
            } else {
                setMessage({ type: 'error', text: result.message || 'No se pudieron crear las plantillas' });
            }
        } catch (err: any) {
            setMessage({ type: 'error', text: err?.message || 'No se pudieron crear las plantillas' });
        } finally {
            setProvisioning(false);
        }
    };

    if (enabled && hosted) {
        return (
            <Card
                icon={<ArrowPathIcon style={{ color: ACCENT, width: 16, height: 16 }} />}
                title="Plantillas"
                description={'Tu n\u00famero vive en la cuenta de WhatsApp de Probability, as\u00ed que usa las plantillas que ya est\u00e1n aprobadas. No hay nada que crear ni que esperar.'}
                action={<Pill tone="ok">Aprobadas</Pill>}
            />
        );
    }

    if (!enabled) {
        return (
            <Card
                icon={<ArrowPathIcon style={{ color: ACCENT, width: 16, height: 16 }} />}
                title="Plantillas"
                description={'Est\u00e1s usando el n\u00famero de Probability, que ya tiene todas las plantillas aprobadas. Conecta tu propio n\u00famero para gestionar las tuyas.'}
                action={<Pill tone="ok">Aprobadas</Pill>}
            />
        );
    }

    const pending = templates.filter((t) => t.status !== 'APPROVED').length;

    return (
        <Card
            icon={<ArrowPathIcon style={{ color: ACCENT, width: 16, height: 16 }} />}
            title="Plantillas"
            description={
                <>
                    {wabaId ? `Cuenta de WhatsApp ${wabaId}. ` : ''}
                    {'Meta puede tardar horas en aprobarlas. Mientras una plantilla no est\u00e9 aprobada, los mensajes que la usan no se pueden enviar.'}
                </>
            }
            action={
                <button
                    type="button"
                    onClick={() => load(true)}
                    disabled={loading}
                    className="p-2 rounded-lg bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50"
                    style={{ border: '1px solid #e9e9f0' }}
                    title="Actualizar estado"
                >
                    <ArrowPathIcon className={`w-4 h-4 text-gray-500 ${loading ? 'animate-spin' : ''}`} />
                </button>
            }
        >
            <div className="space-y-3">
                {templates.length === 0 ? (
                    <p className="text-[13px] text-gray-500 dark:text-gray-400">
                        {loading ? 'Consultando plantillas...' : 'Todav\u00eda no hay plantillas en tu cuenta.'}
                    </p>
                ) : (
                    <>
                        <p className="text-[11px] text-gray-500 dark:text-gray-400">
                            {templates.length} plantillas, {pending} pendientes de {'aprobaci\u00f3n'}.
                        </p>
                        <div
                            className="max-h-72 overflow-y-auto rounded-lg bg-white dark:bg-gray-800 divide-y divide-gray-100 dark:divide-gray-700"
                            style={{ border: '1px solid #e9e9f0' }}
                        >
                            {templates.map((template) => (
                                <div
                                    key={`${template.name}-${template.language}`}
                                    className="px-3 py-2.5 flex items-center justify-between gap-3"
                                >
                                    <div className="min-w-0">
                                        <p className="text-[13px] text-gray-900 dark:text-white font-mono truncate">
                                            {template.name}
                                        </p>
                                        {template.reason && (
                                            <p className="text-[11px] text-red-600 dark:text-red-400 truncate">
                                                {template.reason}
                                            </p>
                                        )}
                                    </div>
                                    <span
                                        className={`px-2 py-1 rounded-full text-[11px] font-semibold shrink-0 ${
                                            STATUS_STYLES[template.status] || 'bg-gray-100 text-gray-700'
                                        }`}
                                    >
                                        {STATUS_LABELS[template.status] || template.status}
                                    </span>
                                </div>
                            ))}
                        </div>
                    </>
                )}

                <ActionButton
                    variant="ghost"
                    onClick={handleProvision}
                    disabled={provisioning}
                    loading={provisioning}
                    className="w-full"
                >
                    {'Crear las plantillas que faltan'}
                </ActionButton>

                {message && (
                    <Alert type={message.type === 'info' ? 'success' : message.type} onClose={() => setMessage(null)}>
                        {message.text}
                    </Alert>
                )}
            </div>
        </Card>
    );
}
