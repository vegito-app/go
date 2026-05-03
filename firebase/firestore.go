package firebase

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func FirestoreIsNotFound(err error) bool {
	return status.Code(err) == codes.NotFound

}
func FirestoreIsAlreadyExists(err error) bool {
	return status.Code(err) == codes.AlreadyExists
}
