import { IWhatsAppRepository } from '../domain/ports';
import {
    WhatsAppConnectionResponse,
    WhatsAppConnectionValues,
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
}
