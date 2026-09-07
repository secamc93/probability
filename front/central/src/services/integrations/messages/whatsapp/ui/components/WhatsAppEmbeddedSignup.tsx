'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import { Alert } from '@/shared/ui';
import { BoltIcon } from '@heroicons/react/24/outline';
import { ACCENT, ActionButton, Card, Pill } from './ui-kit';
import {
    completeWhatsAppEmbeddedSignupAction,
    getWhatsAppEmbeddedSignupConfigAction,
} from '../../infra/actions';
import { WhatsAppEmbeddedSignupConfig } from '../../domain/types';

interface WhatsAppEmbeddedSignupProps {
    businessId?: number;
    onConnected?: () => void;
}

declare global {
    interface Window {
        FB?: any;
        fbAsyncInit?: () => void;
    }
}

function cargarSDK(appId: string, version: string): Promise<void> {
    return new Promise((resolve, reject) => {
        if (window.FB) {
            resolve();
            return;
        }

        const existente = document.getElementById('facebook-jssdk');
        if (existente) {
            existente.addEventListener('load', () => resolve());
            existente.addEventListener('error', () => reject(new Error('No se pudo cargar el SDK de Meta')));
            return;
        }

        window.fbAsyncInit = () => {
            window.FB?.init({ appId, cookie: true, xfbml: false, version });
            resolve();
        };

        const script = document.createElement('script');
        script.id = 'facebook-jssdk';
        script.src = 'https://connect.facebook.net/es_LA/sdk.js';
        script.async = true;
        script.defer = true;
        script.crossOrigin = 'anonymous';
        script.onerror = () => reject(new Error('No se pudo cargar el SDK de Meta'));
        document.body.appendChild(script);
    });
}

export default function WhatsAppEmbeddedSignup({ businessId, onConnected }: WhatsAppEmbeddedSignupProps) {
    const [config, setConfig] = useState<WhatsAppEmbeddedSignupConfig | null>(null);
    const [busy, setBusy] = useState(false);
    const [pin, setPin] = useState('');
    const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
    const datosSesion = useRef<{ waba_id: string; phone_number_id: string } | null>(null);

    useEffect(() => {
        getWhatsAppEmbeddedSignupConfigAction().then((result) => {
            setConfig(result.data || { enabled: false });
        });
    }, []);

    useEffect(() => {
        const onMessage = (event: MessageEvent) => {
            if (event.origin !== 'https://www.facebook.com' && event.origin !== 'https://web.facebook.com') {
                return;
            }
            try {
                const data = JSON.parse(event.data);
                if (data.type !== 'WA_EMBEDDED_SIGNUP') return;
                if (data.event === 'FINISH' || data.event === 'FINISH_ONLY_WABA') {
                    datosSesion.current = {
                        waba_id: data.data?.waba_id || '',
                        phone_number_id: data.data?.phone_number_id || '',
                    };
                } else if (data.event === 'CANCEL') {
                    setMessage({ type: 'error', text: `Cancelaste el registro en el paso: ${data.data?.current_step || ''}` });
                } else if (data.event === 'ERROR') {
                    setMessage({ type: 'error', text: data.data?.error_message || 'Meta reportó un error en el registro' });
                }
            } catch {
                return;
            }
        };

        window.addEventListener('message', onMessage);
        return () => window.removeEventListener('message', onMessage);
    }, []);

    const conectar = useCallback(async () => {
        if (!config?.enabled || !config.app_id || !config.config_id) return;

        setBusy(true);
        setMessage(null);
        datosSesion.current = null;

        try {
            await cargarSDK(config.app_id, config.graph_version || 'v22.0');

            const respuesta: any = await new Promise((resolve) => {
                window.FB.login(resolve, {
                    config_id: config.config_id,
                    response_type: 'code',
                    override_default_response_type: true,
                    extras: { setup: {}, featureType: '', sessionInfoVersion: '3' },
                });
            });

            const code = respuesta?.authResponse?.code;
            if (!code) {
                setMessage({ type: 'error', text: 'No completaste el registro con Meta' });
                return;
            }

            const sesion = datosSesion.current as { waba_id: string; phone_number_id: string } | null;
            const result = await completeWhatsAppEmbeddedSignupAction(
                {
                    code,
                    waba_id: sesion?.waba_id || '',
                    phone_number_id: sesion?.phone_number_id || '',
                },
                businessId
            );

            if (!result.success || !result.data) {
                setMessage({ type: 'error', text: result.message || 'No se pudo completar la conexión' });
                return;
            }

            if (result.data.pin) setPin(result.data.pin);

            setMessage({
                type: result.data.warning ? 'error' : 'success',
                text:
                    result.data.warning ||
                    `Conectado: ${result.data.display_phone_number || result.data.phone_number_id}`,
            });
            onConnected?.();
        } catch (err: any) {
            setMessage({ type: 'error', text: err?.message || 'No se pudo abrir el registro de Meta' });
        } finally {
            setBusy(false);
        }
    }, [businessId, config, onConnected]);

    if (!config?.enabled) return null;

    return (
        <Card
            icon={<BoltIcon style={{ color: ACCENT, width: 16, height: 16 }} />}
            title={'Conectar con Meta (recomendado)'}
            description={'Inicias sesi\u00f3n con tu cuenta de Facebook en una ventana de Meta, eliges o creas tu cuenta de WhatsApp Business y tu n\u00famero queda conectado.'}
            action={<Pill tone="ok">Recomendado</Pill>}
        >
            <div className="space-y-3">
                <ul className="space-y-1 text-[11px] text-gray-600 dark:text-gray-300">
                    <li>{'\u2713 No nos das tu contrase\u00f1a ni copias identificadores a mano.'}</li>
                    <li>{'\u2713 Meta nos entrega un acceso propio de la aplicaci\u00f3n, que no caduca a los pocos d\u00edas.'}</li>
                    <li>{'\u2713 Dejamos configurado el webhook y el registro del n\u00famero en un solo paso.'}</li>
                    <li>{'\u2713 Puedes revocarnos el acceso cuando quieras desde tu Business Manager.'}</li>
                </ul>

                <ActionButton onClick={conectar} disabled={busy} loading={busy}>
                    {'Conectar cuenta de WhatsApp'}
                </ActionButton>

                {pin && (
                    <Alert type="success">
                        {'Guarda este PIN de dos pasos, no se vuelve a mostrar: '}
                        <span className="font-mono font-semibold">{pin}</span>
                    </Alert>
                )}

                {message && (
                    <Alert type={message.type} onClose={() => setMessage(null)}>
                        {message.text}
                    </Alert>
                )}
            </div>
        </Card>
    );
}
