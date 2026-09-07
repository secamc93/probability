import {
    WhatsAppAddNumberValues,
    WhatsAppConnectionResponse,
    WhatsAppConnectionValues,
    WhatsAppEmbeddedSignupConfigResponse,
    WhatsAppEmbeddedSignupPayload,
    WhatsAppEmbeddedSignupResponse,
    WhatsAppNumberResponse,
    WhatsAppProvisionResponse,
    WhatsAppTemplatesResponse,
} from './types';

export interface IWhatsAppRepository {
    getTemplatesStatus(businessId?: number, refresh?: boolean): Promise<WhatsAppTemplatesResponse>;
    provisionTemplates(businessId?: number): Promise<WhatsAppProvisionResponse>;
    saveConnection(values: WhatsAppConnectionValues, businessId?: number): Promise<WhatsAppConnectionResponse>;
    getNumberState(businessId?: number): Promise<WhatsAppNumberResponse>;
    addNumber(values: WhatsAppAddNumberValues, businessId?: number): Promise<WhatsAppNumberResponse>;
    requestNumberCode(method: string, businessId?: number): Promise<WhatsAppNumberResponse>;
    verifyNumberCode(code: string, businessId?: number): Promise<WhatsAppNumberResponse>;
    registerNumber(businessId?: number): Promise<WhatsAppNumberResponse>;
    getEmbeddedSignupConfig(): Promise<WhatsAppEmbeddedSignupConfigResponse>;
    completeEmbeddedSignup(
        payload: WhatsAppEmbeddedSignupPayload,
        businessId?: number
    ): Promise<WhatsAppEmbeddedSignupResponse>;
}
