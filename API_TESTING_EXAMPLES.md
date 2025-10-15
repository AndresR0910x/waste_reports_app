# Ejemplo de registro de usuario ciudadano

## URL del endpoint
POST http://localhost:8082/api/v1/auth/register

## Headers
Content-Type: application/json

## Body
{
  "email": "juan.perez@example.com",
  "password": "MiPassword123!",
  "full_name": "Juan Pérez",
  "phone": "+51987654321",
  "cedula": "12345678",
  "language": "es"
}

## Respuesta esperada (HTTP 201)
{
  "success": true,
  "message": "User registered successfully",
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "user": {
    "id": 1,
    "firebase_uid": "firebase_uid_123",
    "email": "juan.perez@example.com",
    "full_name": "Juan Pérez",
    "role": "ciudadano",
    "role_id": 3,
    "is_active": true,
    "language": "es",
    "created_at": "2024-10-15T09:00:00Z"
  },
  "expires_at": "2024-10-16T09:00:00Z"
}

---

# Ejemplo de login

## URL del endpoint  
POST http://localhost:8082/api/v1/auth/login

## Headers
Content-Type: application/json

## Body
{
  "email": "juan.perez@example.com",
  "password": "MiPassword123!"
}

## Respuesta esperada (HTTP 200)
{
  "success": true,
  "message": "Login successful",
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "user": {
    "id": 1,
    "firebase_uid": "firebase_uid_123",
    "email": "juan.perez@example.com",
    "full_name": "Juan Pérez",
    "role": "ciudadano",
    "role_id": 3,
    "is_active": true,
    "language": "es",
    "created_at": "2024-10-15T09:00:00Z"
  },
  "expires_at": "2024-10-16T09:00:00Z"
}

---

# Errores comunes y soluciones

## Error de CORS
- Asegúrate de usar el puerto correcto (8082)
- Los orígenes permitidos están configurados en main.go

## Error de Firebase
- Verifica que Email/Password esté habilitado en Firebase Console
- Confirma que las credenciales de Firebase sean correctas

## Error de base de datos 
- Verifica que la conexión a Neon esté funcionando
- Confirma que las migraciones se ejecutaron correctamente

## Error de validación
- Email: debe ser formato válido (user@domain.com)
- Password: mínimo 6 caracteres, debe contener letra, número y carácter especial
- Cédula: 8 dígitos para DNI peruano
- Teléfono: formato +51XXXXXXXXX

---

# Comandos útiles para debugging

## Verificar roles en base de datos
```bash
psql "postgresql://neondb_owner:npg_jnw3bVupEP5i@ep-gentle-pond-adcmrdsv.c-2.us-east-1.aws.neon.tech:5432/waste_reports_db?sslmode=require" -c "SELECT * FROM roles;"
```

## Verificar usuarios creados
```bash
psql "postgresql://neondb_owner:npg_jnw3bVupEP5i@ep-gentle-pond-adcmrdsv.c-2.us-east-1.aws.neon.tech:5432/waste_reports_db?sslmode=require" -c "SELECT id, firebase_uid, email, full_name, role_id, is_active FROM users;"
```

## Health check
```bash
curl http://localhost:8082/api/v1/health
```