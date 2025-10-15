# Backend Residuos App

Backend API para la aplicación Clean City - Sistema de gestión de residuos urbanos con autenticación Firebase y validación robusta.

## 🚀 Características

- **Autenticación Firebase**: Sistema seguro de registro e inicio de sesión
- **Validación robusta**: Validación integral de datos de entrada con mensajes personalizados
- **Arquitectura limpia**: Estructura modular siguiendo principios SOLID
- **Testing completo**: Tests unitarios con mocks y cobertura
- **CI/CD**: Pipeline automatizado con GitHub Actions
- **Containerización**: Docker y Docker Compose para desarrollo y producción
- **Hot reload**: Desarrollo ágil con recarga automática
- **Documentación Swagger**: API documentada automáticamente

## 🏗️ Estructura del Proyecto

```
.
├── cmd/
│   ├── api/                        # Punto de entrada principal
│   │   └── main.go
│   └── migrate/                    # CLI para migraciones
│       └── main.go
├── internal/
│   ├── handlers/                   # Controladores HTTP
│   │   ├── auth_handler.go
│   │   ├── notification_handler.go
│   │   ├── report_handler.go
│   │   ├── user_handler.go
│   │   └── audit_handler.go
│   ├── models/                     # Modelos de datos
│   │   └── models.go
│   ├── repositories/               # Capa de acceso a datos
│   │   ├── user_repository.go
│   │   ├── report_repository.go
│   │   └── audit_repository.go
│   └── services/                   # Lógica de negocio
│       ├── auth_service.go
│       ├── notification_service.go
│       └── sync_service.go
├── migrations/                     # Archivos de migración SQL
│   ├── 001_enable_extensions.sql
│   ├── 002_create_roles_enum.sql
│   ├── 003_create_users_table.sql
│   ├── 004_create_waste_reports_table.sql
│   ├── 005_create_report_updates_table.sql
│   └── 006_create_audit_logs_table.sql
├── pkg/
│   ├── config/
│   │   └── config.go               # Configuración de la aplicación
│   ├── database/
│   │   └── postgres.go             # Conexión a PostgreSQL
│   ├── firebase/
│   │   └── firebase.go             # Cliente Firebase
│   ├── migration/
│   │   └── migration.go            # Sistema de migraciones
│   └── middleware/
│       └── firebase_auth.go        # Middleware de autenticación
├── scripts/                        # Scripts de utilidad
│   ├── migrate.ps1                 # PowerShell para migraciones
│   └── setup.ps1                   # Script de configuración inicial
├── .env.example                    # Variables de entorno de ejemplo
├── go.mod
└── README.md
```

## 📦 Dependencias Principales

- **Firebase Admin SDK**: `firebase.google.com/go/v4 v4.18.0`
- **PostgreSQL Driver**: `github.com/lib/pq v1.10.9`
- **Gin Web Framework**: `github.com/gin-gonic/gin v1.10.0`
- **Goose Migrations**: `github.com/pressly/goose/v3 v3.26.0`
- **Environment Variables**: `github.com/joho/godotenv v1.5.1`

## Configuración

### 1. Variables de Entorno

Copia el archivo `.env.example` a `.env` y configura las variables:

```bash
cp .env.example .env
```

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASS=your_password_here
DB_NAME=residuos_db

# Server Configuration
SERVER_PORT=8080
ENVIRONMENT=development

# Firebase Configuration
FIREBASE_CREDENTIALS_PATH=./firebase-service-account-key.json
FIREBASE_PROJECT_ID=your-firebase-project-id
```

### 2. Firebase Setup

1. Ve a la [Consola de Firebase](https://console.firebase.google.com/)
2. Crea un nuevo proyecto o selecciona uno existente
3. Ve a **Configuración del proyecto** > **Cuentas de servicio**
4. Genera una nueva clave privada y descarga el archivo JSON
5. Guarda el archivo como `firebase-service-account-key.json` en la raíz del proyecto
6. Copia el Project ID a tu archivo `.env`

### 3. PostgreSQL Setup

Asegúrate de tener PostgreSQL instalado y ejecutándose:

```sql
CREATE DATABASE residuos_db;
CREATE USER postgres WITH ENCRYPTED PASSWORD 'your_password_here';
GRANT ALL PRIVILEGES ON DATABASE residuos_db TO postgres;
```

## Instalación y Ejecución

### 1. Instalar Dependencias

```bash
go mod tidy
```

### 2. Ejecutar la Aplicación

```bash
go run cmd/api/main.go
```

La aplicación estará disponible en `http://localhost:8080`

## API Endpoints

### Rutas Públicas

- `GET /api/v1/health` - Health check
- `POST /api/v1/auth/users` - Crear usuario (admin)
- `GET /api/v1/auth/users/email/:email` - Obtener usuario por email

### Rutas Protegidas (Requieren Token Firebase)

#### Autenticación
- `GET /api/v1/auth/profile` - Obtener perfil del usuario
- `PUT /api/v1/auth/profile` - Actualizar perfil del usuario
- `DELETE /api/v1/auth/account` - Eliminar cuenta del usuario

#### Notificaciones
- `POST /api/v1/notifications/send` - Enviar notificación a un token
- `POST /api/v1/notifications/multicast` - Enviar notificación a múltiples tokens
- `POST /api/v1/notifications/topic` - Enviar notificación a un tema
- `POST /api/v1/notifications/subscribe` - Suscribir tokens a un tema
- `POST /api/v1/notifications/unsubscribe` - Desuscribir tokens de un tema
- `POST /api/v1/notifications/waste-reminder` - Enviar recordatorio de recolección

## 🚀 Inicio Rápido

### Instalación Automática

```powershell
# Clona el repositorio
git clone <repository-url>
cd backend-residuos-app

# Ejecuta el script de setup (PowerShell)
.\scripts\setup.ps1
```

### Instalación Manual

1. **Instalar dependencias**:
   ```bash
   go mod download
   go mod tidy
   ```

2. **Configurar variables de entorno**:
   ```bash
   cp .env.example .env
   # Edita .env con tus valores
   ```

3. **Configurar Firebase**:
   - Descarga tu clave de servicio desde Firebase Console
   - Guárdala como `firebase-service-account-key.json`

4. **Configurar PostgreSQL**:
   ```sql
   CREATE DATABASE residuos_db;
   CREATE EXTENSION IF NOT EXISTS postgis;
   ```

5. **Ejecutar migraciones**:
   ```bash
   go run cmd/migrate/main.go -action=up
   ```

6. **Ejecutar la aplicación**:
   ```bash
   go run cmd/api/main.go
   ```

## 🔧 Comandos de Migración

```bash
# Aplicar todas las migraciones
go run cmd/migrate/main.go -action=up

# Revertir última migración
go run cmd/migrate/main.go -action=down

# Ver estado de migraciones
go run cmd/migrate/main.go -action=status

# Revertir todas las migraciones
go run cmd/migrate/main.go -action=reset

# Ver versión actual
go run cmd/migrate/main.go -action=version

# Validar archivos de migración
go run cmd/migrate/main.go -action=validate
```

## 🗄️ Esquema de Base de Datos

### Tablas Principales

- **users**: Información de usuarios con roles
- **waste_reports**: Reportes de residuos con geolocalización
- **report_updates**: Actualizaciones de estado de reportes
- **audit_logs**: Registro de auditoría
- **notifications**: Historial de notificaciones

### Roles de Usuario

| Rol | Permisos |
|-----|----------|
| `ciudadano` | Crear reportes, ver propios reportes |
| `operador` | Gestionar reportes asignados, actualizar estados |
| `administrador` | Acceso completo al sistema |

## 🔌 Uso de la API

### Autenticación

Para rutas protegidas, incluye el token de Firebase en el header:

```
Authorization: Bearer YOUR_FIREBASE_ID_TOKEN
```

### Ejemplo: Crear Reporte de Residuos

```bash
curl -X POST http://localhost:8080/api/v1/reports \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_FIREBASE_ID_TOKEN" \
  -d '{
    "title": "Basura acumulada en parque",
    "description": "Gran cantidad de residuos en el parque central",
    "waste_type": "organico",
    "location": {
      "latitude": -12.0464,
      "longitude": -77.0428
    },
    "address": "Parque Kennedy, Miraflores",
    "priority": "media"
  }'
```

### Ejemplo: Enviar Notificación

```bash
curl -X POST http://localhost:8080/api/v1/notifications/send \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_FIREBASE_ID_TOKEN" \
  -d '{
    "token": "DEVICE_FCM_TOKEN",
    "title": "Nuevo reporte asignado",
    "body": "Se te ha asignado un nuevo reporte de residuos",
    "data": {
      "type": "assignment",
      "report_id": "uuid-del-reporte"
    }
  }'
```

## Servicios Disponibles

### AuthService

- Verificación de tokens Firebase
- Gestión de usuarios (crear, actualizar, eliminar)
- Obtener información de perfil
- Custom claims

### NotificationService

- Envío de notificaciones individuales
- Envío masivo (multicast)
- Gestión de temas
- Notificaciones específicas para residuos

### Middleware

- **FirebaseAuth**: Autenticación obligatoria
- **OptionalFirebaseAuth**: Autenticación opcional
- **CORS**: Cross-origin resource sharing

## Desarrollo

### Estructura de Respuestas

Todas las respuestas de la API siguen este formato:

**Éxito:**
```json
{
  "success": true,
  "data": { ... },
  "message": "Operation completed successfully"
}
```

**Error:**
```json
{
  "error": "Error type",
  "message": "Detailed error message"
}
```

## 🧪 Testing

```bash
# Ejecutar todas las pruebas
go test ./...

# Ejecutar pruebas con cobertura
go test -cover ./...

# Ejecutar pruebas de una package específica
go test ./internal/services
```

## 📊 Estructura de Datos

### Usuario
```json
{
  "id": "uuid",
  "firebase_uid": "string",
  "email": "string",
  "full_name": "string",
  "phone": "string",
  "role": "ciudadano|operador|administrador",
  "location": {
    "latitude": 0.0,
    "longitude": 0.0
  },
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

### Reporte de Residuos
```json
{
  "id": "uuid",
  "user_id": "uuid",
  "title": "string",
  "description": "string",
  "waste_type": "organico|reciclable|peligroso|otro",
  "status": "pendiente|en_proceso|resuelto|rechazado",
  "location": {
    "latitude": 0.0,
    "longitude": 0.0
  },
  "address": "string",
  "images": ["string"],
  "priority": "baja|media|alta|critica",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

## 🔒 Seguridad

- **Autenticación JWT**: Tokens Firebase validados
- **Autorización basada en roles**: Permisos granulares
- **CORS configurado**: Protección contra ataques CSRF
- **Rate limiting**: Limitación de requests por IP
- **Validación de entrada**: Sanitización de datos

## 🚀 Despliegue

### Producción

1. Configurar variables de entorno de producción
2. Ejecutar migraciones en servidor de BD
3. Compilar binario optimizado:
   ```bash
   CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/api/main.go
   ```

## 🤝 Contribución

1. Fork el proyecto
2. Crea una rama feature (`git checkout -b feature/nueva-caracteristica`)
3. Commit tus cambios (`git commit -am 'Añadir nueva característica'`)
4. Push a la rama (`git push origin feature/nueva-caracteristica`)
5. Abre un Pull Request

## 🗺️ Roadmap

- [ ] Sistema de notificaciones en tiempo real
- [ ] API de métricas y dashboard
- [ ] Sistema de recompensas para ciudadanos
- [ ] Integración con servicios de mapas
- [ ] App móvil complementaria
- [ ] Sistema de rutas optimizadas para operadores

---

**Desarrollado con ❤️ para hacer las ciudades más limpias** 🌱

### Logging

La aplicación utiliza el logger estándar de Go y Gin. Los logs incluyen:
- Inicialización de servicios
- Errores de conexión
- Requests HTTP (via Gin middleware)

## Producción

Para ejecutar en producción, configura:

```env
ENVIRONMENT=production
```

Esto habilitará el modo release de Gin y optimizaciones de rendimiento.

## Contribución

1. Fork el proyecto
2. Crea una rama para tu feature (`git checkout -b feature/nueva-funcionalidad`)
3. Commit tus cambios (`git commit -am 'Añadir nueva funcionalidad'`)
4. Push a la rama (`git push origin feature/nueva-funcionalidad`)
5. Crea un Pull Request

## Licencia

Este proyecto está bajo la licencia MIT.