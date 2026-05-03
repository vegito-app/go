package firebase

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func FirestoreIsNotFound(err error) bool {
	return status.Code(err) == codes.NotFound
}
