'use server';

import { cookies } from 'next/headers';
import { env } from '@/shared/config/env';
import { LoginRepository } from '../repository';
import { LoginUseCase } from '../../app';
import {
    LoginRequest,
    ChangePasswordRequest,
    GeneratePasswordRequest,
    GenerateBusinessTokenRequest
} from '../repository/mapper/request';
import {
    LoginSuccessResponse,
    UserRolesPermissionsSuccessResponse,
    ChangePasswordResponse,
    GeneratePasswordResponse,
    GenerateBusinessTokenSuccessResponse
} from '../repository/mapper/response';

const repository = new LoginRepository();
const useCase = new LoginUseCase(repository);

export const loginAction = async (credentials: LoginRequest): Promise<LoginSuccessResponse> => {
    try {
        const response = await useCase.login(credentials);


        return response;
    } catch (error: any) {
        console.error('Login Action Error:', error.message);
        throw new Error(error.message); 
    }
};

export const changePasswordAction = async (data: ChangePasswordRequest, token?: string): Promise<ChangePasswordResponse> => {
    try {
        if (!token) {
            const cookieStore = await cookies();
            token = cookieStore.get('session_token')?.value;
        }

        if (!token) {
            throw new Error('No se encontró el token de sesión. Por favor, inicia sesión nuevamente.');
        }

        return await useCase.changePassword(data, token);
    } catch (error: any) {
        console.error('Change Password Action Error:', error.message);
        throw new Error(error.message);
    }
};

export const generatePasswordAction = async (data: GeneratePasswordRequest, token: string): Promise<GeneratePasswordResponse> => {
    try {
        return await useCase.generatePassword(data, token);
    } catch (error: any) {
        console.error('Generate Password Action Error:', error.message);
        throw new Error(error.message);
    }
};

export const getRolesPermissionsAction = async (): Promise<UserRolesPermissionsSuccessResponse> => {
    try {
        const cookieStore = await cookies();
        const token = cookieStore.get('session_token')?.value;

        if (!token) {
            throw new Error('No session token found');
        }

        return await useCase.getRolesPermissions(token);
    } catch (error: any) {
        console.error('Get Roles Permissions Action Error:', error.message);
        throw new Error(error.message);
    }
};

export const getSessionAction = async (): Promise<LoginSuccessResponse> => {
    try {
        const cookieStore = await cookies();
        const token = cookieStore.get('session_token')?.value;

        if (!token) {
            throw new Error('No session token found');
        }

        return await useCase.getSession(token);
    } catch (error: any) {
        console.error('Get Session Action Error:', error.message);
        throw new Error(error.message);
    }
};

export async function loginServerAction(email: string, password: string) {
    try {
        const response = await fetch('http://localhost:3050/api/v1/auth/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password }),
        });

        if (!response.ok) {
            const errorData = await response.json();
            return {
                success: false,
                error: errorData.error || errorData.message || 'Error al iniciar sesión',
            };
        }

        const setCookieHeader = response.headers.get('set-cookie');

        if (setCookieHeader) {
            const tokenMatch = setCookieHeader.match(/session_token=([^;]+)/);
            const maxAgeMatch = setCookieHeader.match(/Max-Age=(\d+)/);

            if (tokenMatch && tokenMatch[1]) {
                const cookieStore = await cookies();

                cookieStore.set('session_token', tokenMatch[1], {
                    maxAge: maxAgeMatch ? parseInt(maxAgeMatch[1]) : 7 * 24 * 60 * 60, 
                    path: '/',
                    httpOnly: true,
                    secure: false, 
                    sameSite: 'lax', 
                });
            }
        }

        const data = await response.json();
        return {
            success: true,
            data,
        };
    } catch (error: any) {
        return {
            success: false,
            error: error.message || 'Error al conectar con el servidor',
        };
    }
}



const DEMO_API_BASE = process.env.API_BASE_URL || process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:3050/api/v1';

export async function demoRegisterAction(payload: { full_name: string; business_name: string; email: string; password: string; phone?: string; channel?: 'email' | 'whatsapp'; }): Promise<{ success: boolean; message?: string; error?: string }> {
    try {
        const res = await fetch(`${DEMO_API_BASE}/auth/demo-register`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
            cache: 'no-store',
        });
        const data = await res.json();
        if (!res.ok) {
            return { success: false, error: data.error || data.message || 'No se pudo crear la demo' };
        }
        return { success: true, message: data.message };
    } catch (error: any) {
        return { success: false, error: error.message || 'Error al conectar con el servidor' };
    }
}

export async function verifyEmailAction(token: string): Promise<{ success: boolean; message?: string; error?: string }> {
    try {
        const res = await fetch(`${DEMO_API_BASE}/auth/verify-email`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ token }),
            cache: 'no-store',
        });
        const data = await res.json();
        if (!res.ok) {
            return { success: false, error: data.error || data.message || 'No se pudo verificar la cuenta' };
        }
        return { success: true, message: data.message };
    } catch (error: any) {
        return { success: false, error: error.message || 'Error al conectar con el servidor' };
    }
}

export async function demoVerifyOtpAction(email: string, code: string): Promise<{ success: boolean; message?: string; error?: string }> {
    try {
        const res = await fetch(`${DEMO_API_BASE}/auth/demo-verify-otp`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, code }),
            cache: 'no-store',
        });
        const data = await res.json();
        if (!res.ok || !data.success) {
            return { success: false, error: data.message || data.error || 'Código inválido o expirado' };
        }
        return { success: true, message: data.message };
    } catch (error: any) {
        return { success: false, error: error.message || 'Error al conectar con el servidor' };
    }
}

export async function recoveryChannelsAction(email: string): Promise<{ email: boolean; whatsapp: { available: boolean; masked_phone: string }; error?: string }> {
    try {
        const res = await fetch(`${env.API_BASE_URL}/auth/recovery-channels`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email }),
            cache: 'no-store',
        });
        const data = await res.json();
        if (!res.ok) {
            return { email: true, whatsapp: { available: false, masked_phone: '' }, error: data.error || data.message };
        }
        return { email: data.email, whatsapp: data.whatsapp };
    } catch (error: any) {
        return { email: true, whatsapp: { available: false, masked_phone: '' }, error: error.message || 'Error al conectar con el servidor' };
    }
}

export async function forgotPasswordAction(email: string, channel: 'email' | 'whatsapp' = 'email'): Promise<{ success: boolean; message?: string; error?: string }> {
    try {
        const res = await fetch(`${env.API_BASE_URL}/auth/forgot-password`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, channel }),
            cache: 'no-store',
        });
        const data = await res.json();
        if (!res.ok) {
            return { success: false, error: data.error || data.message || 'No se pudo procesar la solicitud' };
        }
        return { success: true, message: data.message };
    } catch (error: any) {
        return { success: false, error: error.message || 'Error al conectar con el servidor' };
    }
}

export async function verifyOtpAction(email: string, code: string): Promise<{ success: boolean; token?: string; message?: string; error?: string }> {
    try {
        const res = await fetch(`${env.API_BASE_URL}/auth/verify-otp`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, code }),
            cache: 'no-store',
        });
        const data = await res.json();
        if (!res.ok || !data.success) {
            return { success: false, error: data.message || data.error || 'Código inválido o expirado' };
        }
        return { success: true, token: data.token, message: data.message };
    } catch (error: any) {
        return { success: false, error: error.message || 'Error al conectar con el servidor' };
    }
}

export async function resetPasswordAction(token: string, newPassword: string): Promise<{ success: boolean; message?: string; error?: string }> {
    try {
        const res = await fetch(`${env.API_BASE_URL}/auth/reset-password`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ token, new_password: newPassword }),
            cache: 'no-store',
        });
        const data = await res.json();
        if (!res.ok) {
            return { success: false, error: data.error || data.message || 'No se pudo restablecer la contraseña' };
        }
        return { success: true, message: data.message };
    } catch (error: any) {
        return { success: false, error: error.message || 'Error al conectar con el servidor' };
    }
}
