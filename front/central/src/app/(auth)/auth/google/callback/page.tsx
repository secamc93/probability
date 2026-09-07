'use client';

import { Suspense, useEffect, useRef, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { getSessionAction, getRolesPermissionsAction } from '@/services/auth/login/infra/actions';
import { TokenStorage } from '@/shared/config';
import { applyBusinessTheme, resetTheme } from '@/shared/utils/apply-business-theme';

function GoogleCallbackContent() {
    const router = useRouter();
    const searchParams = useSearchParams();
    const [error, setError] = useState<string | null>(null);
    const yaCorrio = useRef(false);

    useEffect(() => {
        if (yaCorrio.current) return;
        yaCorrio.current = true;

        const hidratarSesion = async () => {
            try {
                const sesion = await getSessionAction();
                if (!sesion.success || !sesion.data) {
                    throw new Error('No se pudo recuperar la sesión');
                }

                const datos = sesion.data;

                TokenStorage.setUser({
                    userId: datos.user.id.toString(),
                    name: datos.user.name,
                    email: datos.user.email,
                    role: 'user',
                    avatarUrl: datos.user.avatar_url,
                    is_super_admin: datos.is_super_admin,
                    scope: datos.scope,
                });

                if (datos.businesses) {
                    TokenStorage.setBusinessesData(datos.businesses);
                }

                if (!datos.is_super_admin && datos.businesses?.length > 0) {
                    applyBusinessTheme(datos.businesses[0]);
                } else {
                    resetTheme();
                }

                try {
                    const permisos = await getRolesPermissionsAction();
                    if (permisos.success && permisos.data) {
                        TokenStorage.setPermissions({
                            is_super: permisos.data.is_super,
                            business_id: permisos.data.business_id,
                            business_name: permisos.data.business_name,
                            role_id: permisos.data.role?.id || 0,
                            role_name: permisos.data.role?.name || '',
                            resources: permisos.data.resources || [],
                            subscription_status: permisos.data.subscription_status,
                        });
                    }
                } catch (permErr) {
                    console.warn('No se pudieron obtener los permisos:', permErr);
                    if (datos.is_super_admin) {
                        TokenStorage.setPermissions({
                            is_super: true,
                            business_id: 0,
                            business_name: '',
                            role_id: 0,
                            role_name: 'Super Admin',
                            resources: [],
                            subscription_status: 'active',
                        });
                    }
                }

                if (searchParams.get('require_password_change') === 'true') {
                    router.replace('/profile');
                    return;
                }

                router.replace(datos.is_super_admin ? '/tickets' : '/home');
            } catch (err: any) {
                console.error('Error hidratando la sesión de Google:', err);
                setError('No pudimos completar el inicio de sesión con Google.');
                setTimeout(() => router.replace('/login'), 2500);
            }
        };

        hidratarSesion();
    }, [router, searchParams]);

    return (
        <div className="flex h-screen w-screen flex-col items-center justify-center gap-4 bg-white">
            {error ? (
                <p className="text-sm text-red-600">{error}</p>
            ) : (
                <>
                    <div className="h-12 w-12 animate-spin rounded-full border-b-2 border-[#8B5CF6]" />
                    <p className="text-gray-600">{'Iniciando sesión con Google...'}</p>
                </>
            )}
        </div>
    );
}

export default function GoogleCallbackPage() {
    return (
        <Suspense
            fallback={
                <div className="flex h-screen w-screen items-center justify-center bg-white">
                    <div className="h-12 w-12 animate-spin rounded-full border-b-2 border-[#8B5CF6]" />
                </div>
            }
        >
            <GoogleCallbackContent />
        </Suspense>
    );
}
