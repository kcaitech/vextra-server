package storage

import (
	"kcaitech.com/kcserver/services"
)

func TidyTeamAvatars(removedTeamAvatars []string) {

	storageClient := services.GetStorageClient()

	for _, avatar := range removedTeamAvatars {
		storageClient.AttatchBucket.DeleteObject(avatar)
	}

}
