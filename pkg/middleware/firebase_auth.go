package middleware

import (
	"context"
	"net/http"
	"strings"

	"backend-residuos-app/pkg/firebase"

	"github.com/gin-gonic/gin"
)

// Context keys
const UserContextKey = "user"

// FirebaseAuth middleware para autenticación con Firebase
func FirebaseAuth(firebaseClient *firebase.Firebase) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Obtener el token del header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authorization header required",
				"message": "No authorization header provided",
			})
			c.Abort()
			return
		}

		// Verificar que el header tenga el formato correcto (Bearer token)
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid authorization header format",
				"message": "Expected format: Bearer <token>",
			})
			c.Abort()
			return
		}

		idToken := tokenParts[1]

		// Verificar el token con Firebase
		ctx := context.Background()
		token, err := firebaseClient.VerifyIDToken(ctx, idToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// Agregar información del usuario al contexto
		c.Set("userID", token.UID)
		c.Set("userEmail", token.Claims["email"])
		c.Set("firebaseToken", token)

		c.Next()
	}
}

// OptionalFirebaseAuth middleware opcional para autenticación con Firebase
// No bloquea la request si no hay token, pero lo valida si está presente
func OptionalFirebaseAuth(firebaseClient *firebase.Firebase) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// Si no hay header, continuar sin autenticación
			c.Next()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			// Si el formato es incorrecto, continuar sin autenticación
			c.Next()
			return
		}

		idToken := tokenParts[1]
		ctx := context.Background()
		token, err := firebaseClient.VerifyIDToken(ctx, idToken)
		if err != nil {
			// Si el token es inválido, continuar sin autenticación
			c.Next()
			return
		}

		// Si el token es válido, agregar información del usuario al contexto
		c.Set("userID", token.UID)
		c.Set("userEmail", token.Claims["email"])
		c.Set("firebaseToken", token)

		c.Next()
	}
}

// GetUserID helper para obtener el ID del usuario del contexto
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return "", false
	}
	uid, ok := userID.(string)
	return uid, ok
}

// GetUserEmail helper para obtener el email del usuario del contexto
func GetUserEmail(c *gin.Context) (string, bool) {
	userEmail, exists := c.Get("userEmail")
	if !exists {
		return "", false
	}
	email, ok := userEmail.(string)
	return email, ok
}

// RequireAuth middleware que requiere autenticación
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists || userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authentication required",
				"message": "User must be authenticated to access this resource",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
