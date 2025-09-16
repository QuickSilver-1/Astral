package filesrepo

import "fmt"

func getFilePath(userID, fileID string) string {
	return fmt.Sprintf("%s/%s", userID, fileID)
}
