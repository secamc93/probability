import 'package:flutter/material.dart';
import '../../../services/modules/storefront/ui/screens/storefront_screen.dart';
import '../../../services/modules/website_config/ui/screens/website_config_screen.dart';
import 'module_tabs_scaffold.dart';

class StorefrontModuleScreen extends StatelessWidget {
  const StorefrontModuleScreen({super.key, this.initialTab = 0});

  final int initialTab;

  @override
  Widget build(BuildContext context) {
    return ModuleTabsScaffold(
      title: 'Tienda online',
      subtitle: 'Catalogo p\u00fablico y sitio',
      initialTab: initialTab,
      tabs: const ['Catalogo', 'Configuraci\u00f3n'],
      builder: (context, businessId) => [
        StorefrontScreen(key: ValueKey('storefront_$businessId'), businessId: businessId),
        WebsiteConfigScreen(key: ValueKey('website_$businessId'), businessId: businessId),
      ],
    );
  }
}
