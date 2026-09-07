'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { CheckCircleIcon } from '@heroicons/react/24/outline';
import { GoogleLogo } from '@/services/auth/login/ui/components/GoogleButton';
import { demoRegisterWithGoogleAction, getGoogleSignupInfoAction } from '@/services/auth/login/infra/actions';

const VENTAJAS = ['Datos de ejemplo cargados', 'Sin tarjeta', 'Listo en un minuto'];

function RegistroDemoContent() {
    const router = useRouter();
    const [cuenta, setCuenta] = useState<{ email: string; name: string } | null>(null);
    const [cargando, setCargando] = useState(true);
    const [businessName, setBusinessName] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    useEffect(() => {
        getGoogleSignupInfoAction()
            .then(setCuenta)
            .catch(() => setCuenta(null))
            .finally(() => setCargando(false));
    }, []);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');

        if (businessName.trim().length < 2) {
            setError('Escribe el nombre de tu negocio');
            return;
        }

        setLoading(true);
        try {
            const resultado = await demoRegisterWithGoogleAction(businessName.trim());
            if (!resultado.success) {
                setError(resultado.error || 'No se pudo crear la demo');
                return;
            }
            router.replace('/auth/google/callback?status=ok');
        } catch {
            setError('Error al conectar con el servidor');
        } finally {
            setLoading(false);
        }
    };

    if (cargando) {
        return (
            <div className="flex min-h-screen items-center justify-center bg-[#140c2d]">
                <div className="h-12 w-12 animate-spin rounded-full border-b-2 border-white" />
            </div>
        );
    }

    if (!cuenta) {
        return (
            <div className="flex min-h-screen items-center justify-center bg-[#140c2d] p-4">
                <div className="w-full max-w-[440px] rounded-3xl bg-white p-7 text-center shadow-2xl">
                    <h1 className="text-xl font-extrabold text-gray-900">{'La solicitud expiró'}</h1>
                    <p className="mt-2 text-sm text-gray-500">
                        {'Vuelve al login y crea tu demo con Google de nuevo.'}
                    </p>
                    <button
                        type="button"
                        onClick={() => router.replace('/login')}
                        className="mt-6 w-full rounded-xl bg-[#5b21b6] py-3 text-sm font-semibold text-white transition-colors hover:bg-[#4c1d95]"
                    >
                        Volver al login
                    </button>
                </div>
            </div>
        );
    }

    return (
        <div className="flex min-h-screen items-center justify-center bg-[#140c2d] p-4">
            <div className="w-full max-w-[440px] rounded-3xl bg-white p-7 shadow-2xl">
                <h1 className="text-[22px] font-extrabold leading-tight text-gray-900">
                    {'¿Cómo se llama tu negocio?'}
                </h1>
                <p className="mt-1.5 text-sm text-gray-500">
                    {'Es lo último que necesitamos para armar tu demo.'}
                </p>

                <div className="mt-4 flex items-center gap-2 rounded-xl bg-gray-50 px-3.5 py-2.5">
                    <GoogleLogo />
                    <span className="truncate text-sm text-gray-600">{cuenta.email}</span>
                </div>

                <ul className="mt-3.5 flex flex-wrap gap-x-4 gap-y-1.5">
                    {VENTAJAS.map((v) => (
                        <li key={v} className="flex items-center gap-1.5 text-xs font-medium text-gray-600">
                            <CheckCircleIcon className="h-4 w-4 text-[#8B5CF6]" />
                            {v}
                        </li>
                    ))}
                </ul>

                <form onSubmit={handleSubmit} className="mt-6">
                    <label className="mb-1.5 block text-xs font-semibold uppercase tracking-wide text-gray-500">
                        Nombre del negocio
                    </label>
                    <input
                        type="text"
                        autoFocus
                        required
                        value={businessName}
                        onChange={(e) => setBusinessName(e.target.value)}
                        placeholder="Mi Tienda"
                        className="w-full rounded-xl border border-gray-200 bg-gray-50/60 px-3.5 py-2.5 text-sm text-gray-900 outline-none transition-colors placeholder:text-gray-400 focus:border-[#8B5CF6] focus:bg-white focus:ring-2 focus:ring-[#8B5CF6]/20"
                    />

                    {error && (
                        <div className="mt-4 rounded-xl border border-red-200 bg-red-50 p-3 text-sm text-red-700">
                            {error}
                        </div>
                    )}

                    <button
                        type="submit"
                        disabled={loading}
                        className="mt-5 w-full rounded-xl bg-[#5b21b6] py-3 text-sm font-semibold text-white transition-colors hover:bg-[#4c1d95] disabled:opacity-60"
                    >
                        {loading ? 'Creando tu demo...' : 'Crear mi demo'}
                    </button>
                </form>
            </div>
        </div>
    );
}

export default function RegistroDemoPage() {
    return <RegistroDemoContent />;
}
