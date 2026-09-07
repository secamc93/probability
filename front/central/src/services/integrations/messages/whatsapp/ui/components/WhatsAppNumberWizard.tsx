'use client';

import { useCallback, useEffect, useState } from 'react';
import { Alert, Button, Input } from '@/shared/ui';
import {
    addWhatsAppNumberAction,
    getWhatsAppNumberStateAction,
    registerWhatsAppNumberAction,
    requestWhatsAppNumberCodeAction,
    verifyWhatsAppNumberCodeAction,
} from '../../infra/actions';
import { WhatsAppNumberState } from '../../domain/types';

interface WhatsAppNumberWizardProps {
    businessId?: number;
    onChanged?: () => void;
}

const PASOS = [
    { key: 'sin_numero', label: 'Tu número' },
    { key: 'esperando_codigo', label: 'Código' },
    { key: 'verificado', label: 'Activación' },
    { key: 'registrado', label: 'Listo' },
];

function indiceDePaso(status: string): number {
    const idx = PASOS.findIndex((paso) => paso.key === status);
    if (idx >= 0) return idx;
    return status === 'nombre_en_revision' ? 3 : 0;
}

export default function WhatsAppNumberWizard({ businessId, onChanged }: WhatsAppNumberWizardProps) {
    const [state, setState] = useState<WhatsAppNumberState | null>(null);
    const [countryCode, setCountryCode] = useState('57');
    const [phone, setPhone] = useState('');
    const [name, setName] = useState('');
    const [code, setCode] = useState('');
    const [pin, setPin] = useState('');
    const [busy, setBusy] = useState(false);
    const [loading, setLoading] = useState(true);
    const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

    const load = useCallback(async () => {
        setLoading(true);
        const result = await getWhatsAppNumberStateAction(businessId);
        if (result.success && result.data) {
            setState(result.data);
        }
        setLoading(false);
    }, [businessId]);

    useEffect(() => {
        load();
    }, [load]);

    const correr = async (accion: () => Promise<any>, exito: string) => {
        setBusy(true);
        setMessage(null);
        try {
            const result = await accion();
            if (result.success && result.data) {
                setState(result.data);
                if (result.data.pin) setPin(result.data.pin);
                setMessage({ type: 'success', text: exito });
                onChanged?.();
            } else {
                setMessage({ type: 'error', text: result.message || 'No se pudo completar el paso' });
            }
        } catch (err: any) {
            setMessage({ type: 'error', text: err?.message || 'No se pudo completar el paso' });
        } finally {
            setBusy(false);
        }
    };

    const status = state?.status || 'sin_numero';
    const pasoActual = indiceDePaso(status);

    return (
        <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-4 space-y-4">
            <div>
                <h4 className="text-sm font-medium text-gray-700 dark:text-gray-200">
                    {'Conectar tu propio número'}
                </h4>
                <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                    {'Tu número queda dentro de la cuenta de WhatsApp de Probability: usa las plantillas ya aprobadas y nosotros gestionamos el consumo. El número no puede estar activo en WhatsApp normal.'}
                </p>
            </div>

            <ol className="flex flex-wrap gap-2 text-xs">
                {PASOS.map((paso, idx) => (
                    <li
                        key={paso.key}
                        className={`px-2 py-1 rounded-full ${
                            idx < pasoActual
                                ? 'bg-green-100 text-green-800'
                                : idx === pasoActual
                                  ? 'bg-blue-100 text-blue-800'
                                  : 'bg-gray-100 text-gray-500'
                        }`}
                    >
                        {idx + 1}. {paso.label}
                    </li>
                ))}
            </ol>

            {loading ? (
                <p className="text-sm text-gray-500 dark:text-gray-400">{'Consultando el estado...'}</p>
            ) : (
                <>
                    {status === 'sin_numero' && (
                        <div className="space-y-3">
                            <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                                <div>
                                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1">
                                        {'Indicativo'}
                                    </label>
                                    <Input value={countryCode} onChange={(e) => setCountryCode(e.target.value)} placeholder="57" />
                                </div>
                                <div className="md:col-span-2">
                                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1">
                                        {'Número de teléfono'}
                                    </label>
                                    <Input value={phone} onChange={(e) => setPhone(e.target.value)} placeholder="3001234567" />
                                </div>
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1">
                                    {'Nombre que verán tus clientes'}
                                </label>
                                <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="Mi Tienda" />
                                <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                                    {'Meta revisa este nombre. Debe corresponder a tu negocio real.'}
                                </p>
                            </div>
                            <Button
                                type="button"
                                variant="primary"
                                disabled={busy}
                                loading={busy}
                                onClick={() =>
                                    correr(
                                        () =>
                                            addWhatsAppNumberAction(
                                                { country_code: countryCode, phone_number: phone, verified_name: name },
                                                businessId
                                            ),
                                        'Número agregado. Pide el código para verificarlo.'
                                    )
                                }
                            >
                                {'Agregar número'}
                            </Button>
                        </div>
                    )}

                    {status === 'esperando_codigo' && (
                        <div className="space-y-3">
                            <p className="text-sm text-gray-700 dark:text-gray-200">
                                {'Te enviaremos un código de 6 dígitos al teléfono para confirmar que es tuyo.'}
                            </p>
                            <div className="flex flex-wrap gap-2">
                                <Button
                                    type="button"
                                    variant="outline"
                                    disabled={busy}
                                    onClick={() => correr(() => requestWhatsAppNumberCodeAction('SMS', businessId), 'Código enviado por SMS')}
                                >
                                    {'Enviar por SMS'}
                                </Button>
                                <Button
                                    type="button"
                                    variant="outline"
                                    disabled={busy}
                                    onClick={() => correr(() => requestWhatsAppNumberCodeAction('VOICE', businessId), 'Te llamaremos con el código')}
                                >
                                    {'Recibir llamada'}
                                </Button>
                            </div>
                            <div className="flex items-end gap-2">
                                <div className="flex-1">
                                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-1">
                                        {'Código recibido'}
                                    </label>
                                    <Input value={code} onChange={(e) => setCode(e.target.value)} placeholder="123456" className="font-mono" />
                                </div>
                                <Button
                                    type="button"
                                    variant="primary"
                                    disabled={busy}
                                    loading={busy}
                                    onClick={() => correr(() => verifyWhatsAppNumberCodeAction(code, businessId), 'Número verificado')}
                                >
                                    {'Verificar'}
                                </Button>
                            </div>
                        </div>
                    )}

                    {status === 'verificado' && (
                        <div className="space-y-3">
                            <p className="text-sm text-gray-700 dark:text-gray-200">
                                {'El número es tuyo. Falta activarlo para enviar y recibir mensajes.'}
                            </p>
                            <Button
                                type="button"
                                variant="primary"
                                disabled={busy}
                                loading={busy}
                                onClick={() => correr(() => registerWhatsAppNumberAction(businessId), 'Número activado')}
                            >
                                {'Activar número'}
                            </Button>
                        </div>
                    )}

                    {(status === 'registrado' || status === 'nombre_en_revision') && (
                        <div className="space-y-2 text-sm text-gray-700 dark:text-gray-200">
                            <p>
                                {'Tus mensajes salen desde '}
                                <span className="font-mono">{state?.display_phone_number || state?.phone_number_id}</span>
                                {state?.quality_rating ? ` (calidad ${state.quality_rating})` : ''}
                            </p>
                            {status === 'nombre_en_revision' && (
                                <p className="text-xs text-yellow-700 dark:text-yellow-500">
                                    {'Meta todavía está revisando el nombre que verán tus clientes.'}
                                </p>
                            )}
                        </div>
                    )}

                    {pin && (
                        <Alert type="success">
                            {'Guarda este PIN de dos pasos, no se vuelve a mostrar: '}
                            <span className="font-mono font-semibold">{pin}</span>
                        </Alert>
                    )}
                </>
            )}

            {message && (
                <Alert type={message.type} onClose={() => setMessage(null)}>
                    {message.text}
                </Alert>
            )}
        </div>
    );
}
