import 'package:dio/dio.dart';

/// Traduce excepciones de Dio a mensajes claros en espa\u00f1ol para el usuario.
String parseError(dynamic e) {
  if (e is DioException) {
    switch (e.type) {
      case DioExceptionType.connectionError:
        return 'No se pudo conectar al servidor. Verifica tu conexi\u00f3n a internet o que el servidor est\u00e9 activo.';
      case DioExceptionType.connectionTimeout:
        return 'La conexi\u00f3n con el servidor tard\u00f3 demasiado. Intenta de nuevo.';
      case DioExceptionType.sendTimeout:
        return 'No se pudo enviar la solicitud. Verifica tu conexi\u00f3n a internet.';
      case DioExceptionType.receiveTimeout:
        return 'El servidor tard\u00f3 demasiado en responder. Intenta de nuevo.';
      case DioExceptionType.badResponse:
        final statusCode = e.response?.statusCode;
        final data = e.response?.data;
        if (data is Map) {
          if (data.containsKey('error')) return data['error'].toString();
          if (data.containsKey('message')) return data['message'].toString();
        }
        switch (statusCode) {
          case 400:
            return 'Datos inv\u00e1lidos. Verifica la informaci\u00f3n ingresada.';
          case 401:
            return 'Sesi\u00f3n expirada o credenciales incorrectas. Inicia sesi\u00f3n de nuevo.';
          case 403:
            return 'No tienes permisos para realizar esta acci\u00f3n.';
          case 404:
            return 'Recurso no encontrado.';
          case 409:
            return 'Conflicto: el recurso ya existe o fue modificado.';
          case 422:
            return 'Los datos enviados no son v\u00e1lidos. Revisa el formulario.';
          case 500:
            return 'Error interno del servidor. Intenta m\u00e1s tarde.';
          case 502:
          case 503:
            return 'Servidor no disponible. Intenta m\u00e1s tarde.';
          default:
            return 'Error del servidor (c\u00f3digo $statusCode). Intenta m\u00e1s tarde.';
        }
      case DioExceptionType.cancel:
        return 'La solicitud fue cancelada.';
      case DioExceptionType.badCertificate:
        return 'Error de seguridad en la conexi\u00f3n. Contacta al administrador.';
      case DioExceptionType.unknown:
        return 'Error de conexi\u00f3n. Verifica tu conexi\u00f3n a internet.';
    }
  }
  return 'Ocurri\u00f3 un error inesperado. Intenta de nuevo.';
}
