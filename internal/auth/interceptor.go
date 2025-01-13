package auth

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryAuthInterceptor creates a new unary server interceptor for authentication
func UnaryAuthInterceptor(jwtManager *JWTManager) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip authentication for login and register
		if info.FullMethod == "/bookmanagement.AuthService/Login" ||
			info.FullMethod == "/bookmanagement.AuthService/Register" {
			return handler(ctx, req)
		}

		// Get token from metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
		}

		values := md.Get("authorization")
		if len(values) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
		}

		token := values[0]

		// Verify token
		claims, err := jwtManager.Verify(token)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", err)
		}

		// Add user ID to context
		newCtx := context.WithValue(ctx, "user_id", claims.UserID)
		return handler(newCtx, req)
	}
}
