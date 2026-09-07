'use client';

import { useState } from 'react';
import { Alert, SecretInput } from '@/shared/ui';
import { ACCENT, ActionButton, fieldHint, fieldLabel, inputCls } from './ui-kit';
import { saveWhatsAppConnectionAction } from '../../infra/actions';
import { WhatsAppConnectionValues } from '../../domain/types';

interface WhatsAppConnectionFormProps {
    config: Record<string, any>;
    businessId?: number;
    onSaved?: () => void;
}

function readConfig(config: Record<string, any>): WhatsAppConnectionValues {
    return {
        waba_id: config?.waba_id ? String(config.waba_id) : '',
        phone_number_id: config?.phone_number_id ? String(config.phone_number_id) : '',
        access_token: '',
        use_platform_token: config?.use_platform_token !== false,
        hosted: config?.hosted_by_platform !== false,
    };
}

export default function WhatsAppConnectionForm({ config, businessId, onSaved }: WhatsAppConnectionFormProps) {
    const [values, setValues] = useState<WhatsAppConnectionValues>(() => readConfig(config));
    const [saving, setSaving] = useState(false);
    const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

    const hasStoredToken = Boolean(config?.phone_number_id) && config?.use_platform_token === false;

    const handleChange = (field: keyof WhatsAppConnectionValues, value: string | boolean) => {
        setValues((prev) => ({ ...prev, [field]: value }));
    };

    const handleSubmit = async () => {
        setMessage(null);

        if (!values.use_platform_token) {
            if (!values.phone_number_id.trim()) {
                setMessage({ type: 'error', text: 'Falta el Phone Number ID del n\u00famero' });
                return;
            }
            if (!values.hosted && !values.waba_id.trim()) {
                setMessage({ type: 'error', text: 'Con cuenta propia tambi\u00e9n necesitas el WABA ID' });
                return;
            }
        }

        setSaving(true);
        try {
            const result = await saveWhatsAppConnectionAction(
                {
                    hosted: values.hosted,
                    use_platform_token: values.use_platform_token,
                    waba_id: values.use_platform_token || values.hosted ? '' : values.waba_id.trim(),
                    phone_number_id: values.use_platform_token ? '' : values.phone_number_id.trim(),
                    access_token: values.hosted ? '' : values.access_token.trim(),
                },
                businessId
            );

            if (result.success) {
                const numero = result.data?.display_phone_number;
                setMessage({
                    type: 'success',
                    text: values.use_platform_token
                        ? 'Los mensajes vuelven a salir del n\u00famero de Probability'
                        : `N\u00famero verificado con Meta${numero ? `: ${numero}` : ''}`,
                });
                setValues((prev) => ({ ...prev, access_token: '' }));
                onSaved?.();
            } else {
                setMessage({ type: 'error', text: result.message || 'Error al guardar la conexi\u00f3n' });
            }
        } catch (err: any) {
            setMessage({ type: 'error', text: err?.message || 'Error al guardar la conexi\u00f3n' });
        } finally {
            setSaving(false);
        }
    };

    return (
        <div className="space-y-3">
            <p className="text-[11px] text-gray-500 dark:text-gray-400 leading-relaxed">
                {'Solo si ya tienes tu propia cuenta de WhatsApp Business en Meta. Comp\u00e1rtela con Probability desde tu Business Manager y pega aqu\u00ed los identificadores. En este caso las conversaciones las pagas t\u00fa y tus plantillas se crean en tu cuenta.'}
            </p>

            <div className="flex items-center gap-3">
                <button
                    type="button"
                    role="switch"
                    aria-checked={!values.use_platform_token}
                    onClick={() => handleChange('use_platform_token', !values.use_platform_token)}
                    className="relative inline-flex h-7 w-12 items-center rounded-full transition-colors focus:outline-none shrink-0"
                    style={{ backgroundColor: !values.use_platform_token ? ACCENT : '#e5e7eb' }}
                >
                    <span
                        className={`inline-block h-5 w-5 transform rounded-full bg-white shadow-md transition-transform ${
                            !values.use_platform_token ? 'translate-x-6' : 'translate-x-1'
                        }`}
                    />
                </button>
                <div>
                    <p className="text-[13px] font-semibold text-gray-800 dark:text-gray-100 leading-tight">
                        {'Usar mi propio n\u00famero'}
                    </p>
                    <p className="text-[11px] text-gray-500 dark:text-gray-400 leading-tight mt-0.5">
                        {'Apagado: los mensajes salen del n\u00famero de Probability, como hasta ahora.'}
                    </p>
                </div>
            </div>

            {!values.use_platform_token && (
                <div className="space-y-3">
                    <div className="space-y-2">
                        <p className={fieldLabel}>{'\u00bfD\u00f3nde vive el n\u00famero?'}</p>

                        <label className="flex items-start gap-2 text-[13px] text-gray-700 dark:text-gray-200">
                            <input
                                type="radio"
                                className="mt-1"
                                checked={values.hosted}
                                onChange={() => handleChange('hosted', true)}
                            />
                            <span>
                                {'En la cuenta de Probability (recomendado)'}
                                <span className="block text-[11px] text-gray-500 dark:text-gray-400">
                                    {'Nosotros pagamos las conversaciones y te las cobramos, y el n\u00famero usa las plantillas ya aprobadas.'}
                                </span>
                            </span>
                        </label>

                        <label className="flex items-start gap-2 text-[13px] text-gray-700 dark:text-gray-200">
                            <input
                                type="radio"
                                className="mt-1"
                                checked={!values.hosted}
                                onChange={() => handleChange('hosted', false)}
                            />
                            <span>
                                {'En mi propia cuenta de WhatsApp Business'}
                                <span className="block text-[11px] text-gray-500 dark:text-gray-400">
                                    {'T\u00fa pagas las conversaciones a Meta y tus plantillas se crean en tu cuenta.'}
                                </span>
                            </span>
                        </label>
                    </div>

                    <div className="grid grid-cols-1 gap-x-4 gap-y-3 sm:grid-cols-2">
                        {!values.hosted && (
                            <div>
                                <label className={fieldLabel}>
                                    WABA ID <span style={{ color: ACCENT }}>*</span>
                                </label>
                                <input
                                    className={`${inputCls} font-mono`}
                                    style={{ borderColor: '#e9e9f0' }}
                                    value={values.waba_id}
                                    onChange={(e) => handleChange('waba_id', e.target.value)}
                                    placeholder="123456789012345"
                                />
                                <p className={fieldHint}>{'ID de tu cuenta de WhatsApp Business en Meta'}</p>
                            </div>
                        )}

                        <div>
                            <label className={fieldLabel}>
                                Phone Number ID <span style={{ color: ACCENT }}>*</span>
                            </label>
                            <input
                                className={`${inputCls} font-mono`}
                                style={{ borderColor: '#e9e9f0' }}
                                value={values.phone_number_id}
                                onChange={(e) => handleChange('phone_number_id', e.target.value)}
                                placeholder="123456789012345"
                            />
                            <p className={fieldHint}>{'ID del n\u00famero desde el que se env\u00edan los mensajes'}</p>
                        </div>

                        {!values.hosted && (
                            <div className="sm:col-span-2">
                                <label className={fieldLabel}>{'Token de acceso (opcional)'}</label>
                                <SecretInput
                                    value={values.access_token}
                                    onChange={(e) => handleChange('access_token', e.target.value)}
                                    placeholder={hasStoredToken ? 'Gu\u00e1rdalo vac\u00edo para no cambiarlo' : 'EAAxxxxxxxxx...'}
                                    className="font-mono"
                                />
                                <p className={fieldHint}>
                                    {'D\u00e9jalo vac\u00edo si compartiste tu cuenta con Probability desde tu Business Manager: en ese caso enviamos con nuestras credenciales.'}
                                </p>
                            </div>
                        )}
                    </div>
                </div>
            )}

            <ActionButton onClick={handleSubmit} disabled={saving} loading={saving}>
                {'Guardar conexi\u00f3n'}
            </ActionButton>

            {message && (
                <Alert type={message.type} onClose={() => setMessage(null)}>
                    {message.text}
                </Alert>
            )}
        </div>
    );
}
