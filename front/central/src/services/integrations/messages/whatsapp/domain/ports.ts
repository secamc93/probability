import {
    WhatsAppConnectionResponse,
    WhatsAppConnectionValues,
    WhatsAppProvisionResponse,
    WhatsAppTemplatesResponse,
} from './types';

export interface IWhatsAppRepository {
    getTemplatesStatus(businessId?: number, refresh?: boolean): Promise<WhatsAppTemplatesResponse>;
    provisionTemplates(businessId?: number): Promise<WhatsAppProvisionResponse>;
    saveConnection(values: WhatsAppConnectionValues, businessId?: number): Promise<WhatsAppConnectionResponse>;
}
