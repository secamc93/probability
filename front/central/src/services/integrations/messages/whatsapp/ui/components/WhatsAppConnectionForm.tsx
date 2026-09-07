'use client';

import { useState } from 'react';
import { Alert, SecretInput } from '@/shared/ui';
import { ACCENT, ActionButton, fieldHint, fieldLabel, inputCls } from './ui-kit';
import { saveWhatsAppConnectionAction } from '../../infra/actions';

interface WhatsAppConnectionFormProps {
    config: Record<string, any>;
    businessId?: number;
    onSaved?: () => void;
}

export default function WhatsAppConnectionForm({ config, businessId, onSaved }: WhatsAppConnectionFormProps) {
    const yaConectada = config?.use_platform_token === false && !config?.hosted_by_platform;

    const [wabaId, setWabaId] = useState(config?.waba_id ? String(config.waba_id) : '');
    const [phoneNumberId, setPhoneNumberId] = useState(
        yaConectada && config?.phone_number_id ? String(config.phone_number_id) : ''
    );
    const [accessToken, setAccessToken] = useState('');
    const [saving, setSaving] = useState(false);
    const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

    const hasStoredToken = yaConectada;

    const handleSubmit = async () => {
        setMessage(null);

        if (!wabaId.trim() || !phoneNumberId.trim()) {
            setMessage({ type: 'error', text: 'Necesitas el WABA ID y el Phone Number ID de tu cuenta' });
            return;
        }

        setSaving(true);
        try {
            const result = await saveWhatsAppConnectionAction(
                {
                    hosted: false,
                    use_platform_token: false,
                    waba_id: wabaId.trim(),
                    phone_number_id: phoneNumberId.trim(),
                    access_token: accessToken.trim(),
                },
                businessId
            );

            if (result.success) {
                const numero = result.data?.display_phone_number;
                setMessage({
                    type: 'success',
                    text: `N\u00famero verificado con Meta${numero ? `: ${numero}` : ''}`,
                });
                setAccessToken('');
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
                {'Si el bot\u00f3n de arriba no te sirve, comparte tu cuenta con Probability desde tu Business Manager y pega aqu\u00ed los identificadores. T\u00fa pagas las conversaciones a Meta y tus plantillas se crean en tu cuenta.'}
            </p>

            <div className="grid grid-cols-1 gap-x-4 gap-y-3 sm:grid-cols-2">
                <div>
                    <label className={fieldLabel}>
                        WABA ID <span style={{ color: ACCENT }}>*</span>
                    </label>
                    <input
                        className={`${inputCls} font-mono`}
                        style={{ borderColor: '#e9e9f0' }}
                        value={wabaId}
                        onChange={(e) => setWabaId(e.target.value)}
                        placeholder="123456789012345"
                    />
                    <p className={fieldHint}>{'ID de tu cuenta de WhatsApp Business en Meta'}</p>
                </div>

                <div>
                    <label className={fieldLabel}>
                        Phone Number ID <span style={{ color: ACCENT }}>*</span>
                    </label>
                    <input
                        className={`${inputCls} font-mono`}
                        style={{ borderColor: '#e9e9f0' }}
                        value={phoneNumberId}
                        onChange={(e) => setPhoneNumberId(e.target.value)}
                        placeholder="123456789012345"
                    />
                    <p className={fieldHint}>{'ID del n\u00famero desde el que se env\u00edan los mensajes'}</p>
                </div>

                <div className="sm:col-span-2">
                    <label className={fieldLabel}>{'Token de acceso (opcional)'}</label>
                    <SecretInput
                        value={accessToken}
                        onChange={(e) => setAccessToken(e.target.value)}
                        placeholder={hasStoredToken ? 'Gu\u00e1rdalo vac\u00edo para no cambiarlo' : 'EAAxxxxxxxxx...'}
                        className="font-mono"
                    />
                    <p className={fieldHint}>
                        {'D\u00e9jalo vac\u00edo si compartiste tu cuenta con Probability: en ese caso enviamos con nuestras credenciales.'}
                    </p>
                </div>
            </div>

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
