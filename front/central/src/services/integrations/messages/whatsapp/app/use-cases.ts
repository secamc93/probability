import { IWhatsAppRepository } from '../domain/ports';
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
} from '../domain/types';

export class WhatsAppUseCases {
    constructor(private readonly repository: IWhatsAppRepository) {}

    async getTemplatesStatus(businessId?: number, refresh = false): Promise<WhatsAppTemplatesResponse> {
        return this.repository.getTemplatesStatus(businessId, refresh);
    }

    async provisionTemplates(businessId?: number): Promise<WhatsAppProvisionResponse> {
        return this.repository.provisionTemplates(businessId);
    }

    async saveConnection(
        values: WhatsAppConnectionValues,
        businessId?: number
    ): Promise<WhatsAppConnectionResponse> {
        return this.repository.saveConnection(values, businessId);
    }

    async getNumberState(businessId?: number): Promise<WhatsAppNumberResponse> {
        return this.repository.getNumberState(businessId);
    }

    async addNumber(values: WhatsAppAddNumberValues, businessId?: number): Promise<WhatsAppNumberResponse> {
        return this.repository.addNumber(values, businessId);
    }

    async requestNumberCode(method: string, businessId?: number): Promise<WhatsAppNumberResponse> {
        return this.repository.requestNumberCode(method, businessId);
    }

    async verifyNumberCode(code: string, businessId?: number): Promise<WhatsAppNumberResponse> {
        return this.repository.verifyNumberCode(code, businessId);
    }

    async registerNumber(businessId?: number): Promise<WhatsAppNumberResponse> {
        return this.repository.registerNumber(businessId);
    }

    async getEmbeddedSignupConfig(): Promise<WhatsAppEmbeddedSignupConfigResponse> {
        return this.repository.getEmbeddedSignupConfig();
    }

    async completeEmbeddedSignup(
        payload: WhatsAppEmbeddedSignupPayload,
        businessId?: number
    ): Promise<WhatsAppEmbeddedSignupResponse> {
        return this.repository.completeEmbeddedSignup(payload, businessId);
    }
}
