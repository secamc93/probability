'use client';

import { useState } from 'react';
import { Alert } from '@/shared/ui';
import { ChatBubbleLeftRightIcon, PhoneIcon } from '@heroicons/react/24/outline';
import { ACCENT, ACCENT_BORDER, ACCENT_SOFT, ActionButton, Card, Pill, fieldHint, fieldLabel, inputCls } from './ui-kit';
import WhatsAppConnectionForm from './WhatsAppConnectionForm';
import WhatsAppEmbeddedSignup from './WhatsAppEmbeddedSignup';
import WhatsAppNumberWizard from './WhatsAppNumberWizard';
import WhatsAppTemplatesPanel from './WhatsAppTemplatesPanel';

interface WhatsAppIntegrationViewProps {
    integration: {
        id: number;
        name: string;
        code: string;
        business_id?: number | null;
        config?: Record<string, any>;
        credentials?: Record<string, any>;
        is_active: boolean;
        created_at: string;
        updated_at: string;
    };
    imageUrl?: string;
    onToggleActive?: (id: number, currentlyActive: boolean) => Promise<boolean>;
    onUpdateConfig?: (id: number, config: Record<string, any>) => Promise<{ success: boolean; message?: string }>;
    onTestConnection?: (id: number) => Promise<{ success: boolean; message?: string }>;
    onRefresh?: () => void;
}

export default function WhatsAppIntegrationView({
    integration,
    imageUrl,
    onToggleActive,
    onUpdateConfig,
    onTestConnection,
    onRefresh,
}: WhatsAppIntegrationViewProps) {
    const [isActive, setIsActive] = useState(integration.is_active);
    const [toggling, setToggling] = useState(false);
    const [testPhone, setTestPhone] = useState(integration.config?.test_phone_number || '');
    const [saving, setSaving] = useState(false);
    const [testing, setTesting] = useState(false);
    const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

    const savedPhone = integration.config?.test_phone_number || '';
    const usesOwnNumber =
        integration.config?.use_platform_token === false && Boolean(integration.config?.phone_number_id);
    const hasUnsavedChanges = testPhone !== savedPhone;

    const handleToggle = async () => {
        if (!onToggleActive) return;
        setToggling(true);
        try {
            const success = await onToggleActive(integration.id, isActive);
            if (success) {
                setIsActive((prev) => !prev);
            }
        } finally {
            setToggling(false);
        }
    };

    const handleSavePhone = async () => {
        if (!onUpdateConfig || !testPhone.trim()) return;
        setSaving(true);
        setMessage(null);
        try {
            const updatedConfig = { ...integration.config, test_phone_number: testPhone.trim() };
            const result = await onUpdateConfig(integration.id, updatedConfig);
            if (result.success) {
                setMessage({ type: 'success', text: 'Número de prueba guardado' });
                onRefresh?.();
            } else {
                setMessage({ type: 'error', text: result.message || 'Error al guardar' });
            }
        } catch (err: any) {
            setMessage({ type: 'error', text: err.message || 'Error al guardar' });
        } finally {
            setSaving(false);
        }
    };

    const handleTestConnection = async () => {
        if (!onTestConnection) return;

        if (hasUnsavedChanges && onUpdateConfig && testPhone.trim()) {
            setSaving(true);
            try {
                const updatedConfig = { ...integration.config, test_phone_number: testPhone.trim() };
                const saveResult = await onUpdateConfig(integration.id, updatedConfig);
                if (!saveResult.success) {
                    setMessage({ type: 'error', text: saveResult.message || 'Error al guardar antes de probar' });
                    setSaving(false);
                    return;
                }
            } catch (err: any) {
                setMessage({ type: 'error', text: err.message || 'Error al guardar' });
                setSaving(false);
                return;
            }
            setSaving(false);
        }

        setTesting(true);
        setMessage(null);
        try {
            const result = await onTestConnection(integration.id);
            if (result.success) {
                setMessage({ type: 'success', text: 'Mensaje de prueba enviado correctamente' });
            } else {
                setMessage({ type: 'error', text: result.message || 'Error en la prueba de conexión' });
            }
        } catch (err: any) {
            setMessage({ type: 'error', text: err.message || 'Error en la prueba' });
        } finally {
            setTesting(false);
        }
    };

    const numeroPropio = integration.config?.phone_number_id;
    const estadoNumero = integration.config?.number_status;

    return (
        <div className="space-y-3 w-full">
            {message && (
                <Alert type={message.type} onClose={() => setMessage(null)}>
                    {message.text}
                </Alert>
            )}

            <div
                className="flex flex-col gap-3 rounded-xl p-4 sm:flex-row sm:items-center sm:justify-between dark:bg-gray-800/60"
                style={{ backgroundColor: ACCENT_SOFT, border: `1px solid ${ACCENT_BORDER}` }}
            >
                <div className="flex items-center gap-3">
                    <span
                        className="flex h-11 w-11 items-center justify-center rounded-xl overflow-hidden shrink-0 bg-white dark:bg-gray-900"
                        style={{ border: `1px solid ${ACCENT_BORDER}` }}
                    >
                        {imageUrl ? (
                            <img src={imageUrl} alt={integration.name} className="h-8 w-8 object-contain" />
                        ) : (
                            <ChatBubbleLeftRightIcon className="h-6 w-6" style={{ color: ACCENT }} />
                        )}
                    </span>
                    <div className="min-w-0">
                        <h2 className="text-base font-bold text-gray-900 dark:text-white leading-tight">
                            {integration.name}
                        </h2>
                        <p className="text-xs text-gray-600 dark:text-gray-300 mt-0.5">
                            {usesOwnNumber
                                ? 'Los mensajes de tus pedidos salen desde tu propio n\u00famero.'
                                : 'Los mensajes de tus pedidos salen desde el n\u00famero de Probability.'}
                        </p>
                        <p className="text-[11px] text-gray-500 dark:text-gray-400 font-mono mt-0.5">
                            {integration.code}
                        </p>
                    </div>
                </div>

                <div className="flex items-center gap-2 self-start shrink-0">
                    <Pill tone={isActive ? 'ok' : 'off'}>{isActive ? 'Activa' : 'Inactiva'}</Pill>
                    {onToggleActive && (
                        <button
                            type="button"
                            role="switch"
                            aria-checked={isActive}
                            onClick={handleToggle}
                            disabled={toggling}
                            className="relative inline-flex h-7 w-12 items-center rounded-full transition-colors focus:outline-none disabled:opacity-50"
                            style={{ backgroundColor: isActive ? ACCENT : '#e5e7eb' }}
                        >
                            <span
                                className={`inline-block h-5 w-5 transform rounded-full bg-white shadow-md transition-transform ${
                                    isActive ? 'translate-x-6' : 'translate-x-1'
                                }`}
                            />
                        </button>
                    )}
                </div>
            </div>

            {!isActive && (
                <Card
                    icon={<ChatBubbleLeftRightIcon style={{ color: ACCENT, width: 16, height: 16 }} />}
                    title={'Integraci\u00f3n desactivada'}
                    description={'Act\u00edvala con el interruptor de arriba para configurar el n\u00famero y enviar mensajes.'}
                />
            )}

            {isActive && (
                <>
                    <WhatsAppEmbeddedSignup
                        businessId={integration.business_id || undefined}
                        onConnected={onRefresh}
                    />

                    <WhatsAppNumberWizard
                        businessId={integration.business_id || undefined}
                        onChanged={onRefresh}
                    />

                    {(onUpdateConfig || onTestConnection) && (
                        <Card
                            icon={<PhoneIcon style={{ color: ACCENT, width: 16, height: 16 }} />}
                            title={'Probar el env\u00edo'}
                            description={'Mandamos un mensaje de prueba a este n\u00famero para confirmar que la integraci\u00f3n responde. Usa indicativo de pa\u00eds, por ejemplo 573001234567.'}
                        >
                            <div className="space-y-3">
                                <div>
                                    <label className={fieldLabel}>{'N\u00famero de prueba'}</label>
                                    <div className="flex flex-col gap-2 sm:flex-row">
                                        <input
                                            type="text"
                                            value={testPhone}
                                            onChange={(e) => setTestPhone(e.target.value)}
                                            placeholder="573001234567"
                                            className={`${inputCls} font-mono flex-1`}
                                            style={{ borderColor: '#e9e9f0' }}
                                        />
                                        {onUpdateConfig && (
                                            <ActionButton
                                                variant="ghost"
                                                onClick={handleSavePhone}
                                                disabled={!hasUnsavedChanges || !testPhone.trim()}
                                                loading={saving}
                                            >
                                                Guardar
                                            </ActionButton>
                                        )}
                                    </div>
                                    {hasUnsavedChanges && testPhone.trim() && (
                                        <p className={fieldHint}>Sin guardar. Al probar se guarda solo.</p>
                                    )}
                                </div>

                                {onTestConnection && (
                                    <ActionButton
                                        onClick={handleTestConnection}
                                        disabled={!testPhone.trim()}
                                        loading={testing}
                                        className="w-full"
                                    >
                                        Enviar mensaje de prueba
                                    </ActionButton>
                                )}
                            </div>
                        </Card>
                    )}

                    <WhatsAppTemplatesPanel
                        businessId={integration.business_id || undefined}
                        enabled={usesOwnNumber}
                    />

                    <details
                        className="rounded-xl p-4 dark:bg-gray-800/60"
                        style={{ backgroundColor: '#fafafd', border: '1px solid #eceaf3' }}
                    >
                        <summary className="text-sm font-bold text-gray-900 dark:text-white cursor-pointer">
                            {'Ya tengo mi propia cuenta de WhatsApp Business'}
                        </summary>
                        <div className="mt-3">
                            <WhatsAppConnectionForm
                                config={integration.config || {}}
                                businessId={integration.business_id || undefined}
                                onSaved={onRefresh}
                            />
                        </div>
                    </details>

                    <div className="flex flex-wrap items-center justify-between gap-2 px-1 text-[11px] text-gray-400 dark:text-gray-500">
                        <span>Creada {new Date(integration.created_at).toLocaleDateString()}</span>
                        {numeroPropio && (
                            <span className="font-mono">
                                phone_number_id {numeroPropio}
                                {estadoNumero ? ` \u00b7 ${estadoNumero}` : ''}
                            </span>
                        )}
                        <span>Actualizada {new Date(integration.updated_at).toLocaleDateString()}</span>
                    </div>
                </>
            )}
        </div>
    );
}
