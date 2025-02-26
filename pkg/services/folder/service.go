package folder

import (
	"context"
)

type Service interface {
	// GetChildren returns an array containing all child folders.
	GetChildren(ctx context.Context, cmd *GetChildrenQuery) ([]*Folder, error)
	// GetParents returns an array containing add parent folders if nested folders are enabled
	// otherwise it returns an empty array
	GetParents(ctx context.Context, q GetParentsQuery) ([]*Folder, error)
	Create(ctx context.Context, cmd *CreateFolderCommand) (*Folder, error)

	// GetFolder takes a GetFolderCommand and returns a folder matching the
	// request. One of ID, UID, or Title must be included. If multiple values
	// are included in the request, Grafana will select one in order of
	// specificity (ID, UID, Title).
	Get(ctx context.Context, cmd *GetFolderQuery) (*Folder, error)

	// Update is used to update a folder's UID, Title and Description. To change
	// a folder's parent folder, use Move.
	Update(ctx context.Context, cmd *UpdateFolderCommand) (*Folder, error)
	Delete(ctx context.Context, cmd *DeleteFolderCommand) error
	MakeUserAdmin(ctx context.Context, orgID int64, userID, folderID int64, setViewAndEditPermissions bool) error
	// Move changes a folder's parent folder to the requested new parent.
	Move(ctx context.Context, cmd *MoveFolderCommand) (*Folder, error)
}
