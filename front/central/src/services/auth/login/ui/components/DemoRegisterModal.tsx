'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { EyeIcon, EyeSlashIcon, CheckCircleIcon } from '@heroicons/react/24/outline';
import { demoRegisterAction } from '../../infra/actions';
import { GoogleLogo, googleLoginUrl } from './GoogleButton';

interface DemoRegisterModalProps {
  onClose: () => void;
}

const VENTAJAS = ['Datos de ejemplo cargados', 'Sin tarjeta', 'Listo en un minuto'];

const inputClass =
  'w-full rounded-xl border border-gray-200 bg-gray-50/60 px-3.5 py-2.5 text-sm text-gray-900 outline-none transition-colors placeholder:text-gray-400 focus:border-[#8B5CF6] focus:bg-white focus:ring-2 focus:ring-[#8B5CF6]/20';

const labelClass = 'mb-1.5 block text-xs font-semibold uppercase tracking-wide text-gray-500';

export const DemoRegisterModal = ({ onClose }: DemoRegisterModalProps) => {
  const router = useRouter();
  const [fullName, setFullName] = useState('');
  const [businessName, setBusinessName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [phone, setPhone] = useState('');
  const [channel, setChannel] = useState<'email' | 'whatsapp'>('email');
  const [loading, setLoading] = useState(false);
  const [done, setDone] = useState(false);
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    if (password.length < 6) {
      setError('La contraseña debe tener al menos 6 caracteres');
      return;
    }
    if (channel === 'whatsapp' && phone.trim().length < 7) {
      setError('Ingresa un teléfono válido para verificar por WhatsApp');
      return;
    }
    setLoading(true);
    try {
      const result = await demoRegisterAction({
        full_name: fullName.trim(),
        business_name: businessName.trim(),
        email: email.trim(),
        password,
        phone: channel === 'whatsapp' ? phone.trim() : undefined,
        channel,
      });
      if (result.success) {
        if (channel === 'whatsapp') {
          router.push(`/verify-demo?email=${encodeURIComponent(email.trim())}`);
          return;
        }
        setDone(true);
        setMessage(result.message || 'Cuenta creada. Revisa tu correo para verificar tu cuenta.');
      } else {
        setError(result.error || 'No se pudo crear la demo');
      }
    } catch {
      setError('Error al conectar con el servidor');
    } finally {
      setLoading(false);
    }
  };

  const channelBtn = (value: 'email' | 'whatsapp', label: string, hint: string) => (
    <button
      type="button"
      onClick={() => setChannel(value)}
      className={`flex-1 rounded-xl border px-3 py-2 text-left transition-colors ${
        channel === value
          ? 'border-[#8B5CF6] bg-[#8B5CF6]/8 ring-1 ring-[#8B5CF6]'
          : 'border-gray-200 hover:border-gray-300'
      }`}
    >
      <span className="block text-sm font-semibold text-gray-900">{label}</span>
      <span className="block text-xs text-gray-500">{hint}</span>
    </button>
  );

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-[#140c2d]/70 p-4 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="relative flex max-h-[92vh] w-full max-w-[440px] flex-col overflow-hidden rounded-3xl bg-white shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        <button
          type="button"
          onClick={onClose}
          aria-label="Cerrar"
          className="absolute right-4 top-4 z-10 flex h-8 w-8 items-center justify-center rounded-full text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-700"
        >
          <span aria-hidden="true" className="text-lg leading-none">&times;</span>
        </button>

        <div className="shrink-0 px-7 pb-5 pt-7">
          <h2 className="text-[22px] font-extrabold leading-tight text-gray-900">Crea tu demo gratis</h2>
          <p className="mt-1.5 text-sm text-gray-500">
            Prueba Probability con datos simulados de una tienda real.
          </p>
          <ul className="mt-3.5 flex flex-wrap gap-x-4 gap-y-1.5">
            {VENTAJAS.map((v) => (
              <li key={v} className="flex items-center gap-1.5 text-xs font-medium text-gray-600">
                <CheckCircleIcon className="h-4 w-4 text-[#8B5CF6]" />
                {v}
              </li>
            ))}
          </ul>
        </div>

        {done ? (
          <div className="px-7 pb-7">
            <div className="rounded-xl border border-green-200 bg-green-50 p-4 text-sm text-green-800">
              {message}
            </div>
            <button
              type="button"
              onClick={onClose}
              className="mt-5 w-full rounded-xl bg-[#5b21b6] py-3 text-sm font-semibold text-white transition-colors hover:bg-[#4c1d95]"
            >
              Entendido
            </button>
          </div>
        ) : (
          <div className="min-h-0 flex-1 overflow-y-auto px-7 pb-7">
            <a
              href={googleLoginUrl('demo')}
              className="flex w-full items-center justify-center gap-3 rounded-xl border border-gray-300 bg-white px-4 py-3 text-sm font-semibold text-gray-700 transition-colors hover:bg-gray-50"
            >
              <GoogleLogo />
              {'Crear demo con Google'}
            </a>
            <p className="mt-2 text-center text-xs text-gray-400">
              {'Sin contraseña ni verificación: entras de una vez'}
            </p>

            <div className="my-5 flex items-center gap-3">
              <span className="h-px flex-1 bg-gray-200" />
              <span className="text-[11px] font-medium uppercase tracking-wide text-gray-400">
                {'o con tu correo'}
              </span>
              <span className="h-px flex-1 bg-gray-200" />
            </div>

            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className={labelClass}>Tu nombre</label>
                  <input
                    type="text"
                    required
                    value={fullName}
                    onChange={(e) => setFullName(e.target.value)}
                    placeholder={'Juan Pérez'}
                    className={inputClass}
                  />
                </div>
                <div>
                  <label className={labelClass}>Tu negocio</label>
                  <input
                    type="text"
                    required
                    value={businessName}
                    onChange={(e) => setBusinessName(e.target.value)}
                    placeholder="Mi Tienda"
                    className={inputClass}
                  />
                </div>
              </div>

              <div>
                <label className={labelClass}>{'Correo electrónico'}</label>
                <input
                  type="email"
                  required
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="tucorreo@ejemplo.com"
                  className={inputClass}
                />
              </div>

              <div>
                <label className={labelClass}>{'Contraseña'}</label>
                <div className="relative">
                  <input
                    type={showPassword ? 'text' : 'password'}
                    required
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder={'Mínimo 6 caracteres'}
                    className={`${inputClass} pr-11`}
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword((v) => !v)}
                    aria-label={showPassword ? 'Ocultar contraseña' : 'Mostrar contraseña'}
                    className="absolute inset-y-0 right-0 flex items-center pr-3.5 text-gray-400 transition-colors hover:text-gray-600"
                  >
                    {showPassword ? <EyeSlashIcon className="h-5 w-5" /> : <EyeIcon className="h-5 w-5" />}
                  </button>
                </div>
              </div>

              <div>
                <label className={labelClass}>{'Cómo verificamos tu cuenta'}</label>
                <div className="flex gap-2.5">
                  {channelBtn('email', 'Correo', 'Enlace por email')}
                  {channelBtn('whatsapp', 'WhatsApp', 'Código al celular')}
                </div>
              </div>

              {channel === 'whatsapp' && (
                <div>
                  <label className={labelClass}>{'Teléfono de WhatsApp'}</label>
                  <input
                    type="tel"
                    inputMode="tel"
                    value={phone}
                    onChange={(e) => setPhone(e.target.value)}
                    placeholder="3001234567"
                    className={inputClass}
                  />
                </div>
              )}

              {error && (
                <div className="rounded-xl border border-red-200 bg-red-50 p-3 text-sm text-red-700">
                  {error}
                </div>
              )}

              <button
                type="submit"
                disabled={loading}
                className="w-full rounded-xl bg-[#5b21b6] py-3 text-sm font-semibold text-white transition-colors hover:bg-[#4c1d95] disabled:opacity-60"
              >
                {loading ? 'Creando...' : 'Crear mi demo'}
              </button>
            </form>
          </div>
        )}
      </div>
    </div>
  );
};
